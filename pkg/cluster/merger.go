// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cluster

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	oapi "k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

const (
	// maxPatchRetry is the maximum number of conflicts retry for during a patch operation before returning failure
	maxPatchRetry = 5
	// backOffPeriod is the period to back off when apply patch results in error.
	backOffPeriod = 1 * time.Second
	// how many times we can retry before back off
	triesBeforeBackOff = 1
)

// objectMerger merges an object with an object already in the cluster. This
// will ensure that important cluster values aren't overwritten.
type objectMerger interface {
	Merge(namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

// defaultObjectMerger merges an object with an object already in the cluster. This
// will ensure that important cluster values aren't overwritten.
type defaultObjectMerger struct {
	factory cmdutil.Factory
}

var _ objectMerger = (*defaultObjectMerger)(nil)

// newDefaultObjectMerger creates an instance of objectMerge.
func newDefaultObjectMerger(factory cmdutil.Factory) *defaultObjectMerger {
	p := &defaultObjectMerger{
		factory: factory,
	}

	return p
}

// Merge merges an object in a given namespace. It returns the merged object.
func (p *defaultObjectMerger) Merge(namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	file, err := p.stageInTempFile(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "staging %s/%s",
			obj.GroupVersionKind().GroupVersion().String(), obj.GetName())
	}

	defer os.Remove(file.Name())

	options := &resource.FilenameOptions{
		Filenames: []string{
			file.Name(),
		},
	}

	r := p.factory.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).DefaultNamespace().
		FilenameParam(false, options).
		Flatten().
		Latest().
		Do()

	if err = r.Err(); err != nil {
		return nil, errors.Wrap(err, "resource error")
	}

	encoder := scheme.DefaultJSONEncoder()
	deserializer := scheme.Codecs.UniversalDeserializer()

	infos, err := r.Infos()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving resource info")
	}

	if l := len(infos); l != 1 {
		return nil, errors.Errorf("expected resource info to be length 1, but was %d", l)
	}

	info := infos[0]

	if err = info.Get(); err != nil {
		if !kerrors.IsNotFound(err) {
			return nil, cmdutil.AddSourceToErr(fmt.Sprintf("retrieving current configuration of:\n%v\nfrom server for:", info), info.Source, err)
		}
	}

	modified, err := runtime.Encode(encoder, obj)
	if err != nil {
		return nil, errors.Wrap(err, "encode modified object")
	}

	helper := resource.NewHelper(info.Client, info.Mapping)
	patcher := &patcher{
		encoder:       encoder,
		decoder:       deserializer,
		mapping:       info.Mapping,
		helper:        helper,
		clientFunc:    p.factory.UnstructuredClientForMapping,
		clientsetFunc: p.factory.ClientSet,
		overwrite:     true,
		backOff:       clockwork.NewRealClock(),
		force:         false,
		cascade:       false,
		timeout:       0,
		gracePeriod:   0,
	}

	discoveryClient, err := p.factory.DiscoveryClient()
	if err == nil {
		openAPIGetter := openapi.NewOpenAPIGetter(discoveryClient)
		resources, err := openAPIGetter.Get()
		if err == nil {
			patcher.openapiSchema = resources
		}
	}

	patchBytes, patchedObject, err := patcher.patch(info.Object, modified, info.Source, info.Namespace, info.Name, os.Stderr)
	if err != nil {
		logrus.Debugf("applying patch:\n%s\nto:\n%v\nfor:\n", patchBytes, info)
		return nil, errors.Wrap(err, "path object")
	}

	u, ok := patchedObject.(*unstructured.Unstructured)
	if !ok {
		return nil, errors.New("patched object was not *unstructured.Unstructured")
	}

	return u, nil
}

// stageInTempFile stages an object in a temp file. The file will have to be
// manually removed once it is no longer needed.
func (p *defaultObjectMerger) stageInTempFile(obj *unstructured.Unstructured) (*os.File, error) {
	encoded, err := runtime.Encode(scheme.DefaultJSONEncoder(), obj)
	if err != nil {
		return nil, errors.Wrap(err, "encoding input")
	}

	tmpfile, err := ioutil.TempFile("", "ksonnet-mergepatch")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary file")
	}

	if _, err = tmpfile.Write(encoded); err != nil {
		return nil, errors.Wrap(err, "writing temporary file")
	}

	return tmpfile, nil
}

// patcher is the kubectl apply patcher.
type patcher struct {
	encoder runtime.Encoder
	decoder runtime.Decoder

	mapping       *meta.RESTMapping
	helper        *resource.Helper
	clientFunc    resource.ClientMapperFunc
	clientsetFunc func() (internalclientset.Interface, error)

	overwrite bool
	backOff   clockwork.Clock

	force       bool
	cascade     bool
	timeout     time.Duration
	gracePeriod int

	openapiSchema openapi.Resources
}

func (p *patcher) patchSimple(obj runtime.Object, modified []byte, source, namespace, name string, errOut io.Writer) ([]byte, runtime.Object, error) {
	// Serialize the current configuration of the object from the server.
	current, err := runtime.Encode(p.encoder, obj)
	if err != nil {
		return nil, nil, cmdutil.AddSourceToErr(fmt.Sprintf("serializing current configuration from:\n%v\nfor:", obj), source, err)
	}

	t := newDefaultAnnotationApplier()

	// Retrieve the original configuration of the object from the annotation.
	original, err := t.GetOriginalConfiguration(p.mapping, obj)
	if err != nil {
		return nil, nil, cmdutil.AddSourceToErr(fmt.Sprintf("retrieving original configuration from:\n%v\nfor:", obj), source, err)
	}

	var patchType types.PatchType
	var patch []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta
	var schema oapi.Schema
	createPatchErrFormat := "creating patch with:\noriginal:\n%s\nmodified:\n%s\ncurrent:\n%s\nfor:"

	// Create the versioned struct from the type defined in the restmapping
	// (which is the API version we'll be submitting the patch to)
	versionedObject, err := scheme.Scheme.New(p.mapping.GroupVersionKind)
	switch {
	case runtime.IsNotRegisteredError(err):
		// fall back to generic JSON merge patch
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"), mergepatch.RequireMetadataKeyUnchanged("name")}
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			if mergepatch.IsPreconditionFailed(err) {
				return nil, nil, fmt.Errorf("%s", "At least one of apiVersion, kind and name was changed")
			}
			return nil, nil, cmdutil.AddSourceToErr(fmt.Sprintf(createPatchErrFormat, original, modified, current), source, err)
		}
	case err != nil:
		return nil, nil, cmdutil.AddSourceToErr(fmt.Sprintf("getting instance of versioned object for %v:", p.mapping.GroupVersionKind), source, err)
	case err == nil:
		// Compute a three way strategic merge patch to send to server.
		patchType = types.StrategicMergePatchType

		// Try to use openapi first if the openapi spec is available and can successfully calculate the patch.
		// Otherwise, fall back to baked-in types.
		if p.openapiSchema != nil {
			if schema = p.openapiSchema.LookupResource(p.mapping.GroupVersionKind); schema != nil {
				lookupPatchMeta = strategicpatch.PatchMetaFromOpenAPI{Schema: schema}
				openapiPatch, err := strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, p.overwrite)
				if err != nil {
					fmt.Fprintf(errOut, "warning: error calculating patch from openapi spec: %v\n", err)
				} else {
					patchType = types.StrategicMergePatchType
					patch = openapiPatch
				}
			}
		}

		if patch == nil {
			lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
			if err != nil {
				return nil, nil, cmdutil.AddSourceToErr(fmt.Sprintf(createPatchErrFormat, original, modified, current), source, err)
			}

			patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, p.overwrite)
			if err != nil {
				return nil, nil, cmdutil.AddSourceToErr(fmt.Sprintf(createPatchErrFormat, original, modified, current), source, err)
			}
		}
	}

	if string(patch) == "{}" {
		return patch, obj, nil
	}

	patchedObj, err := p.helper.Patch(namespace, name, patchType, patch)
	if err != nil {
		return nil, nil, errors.Wrap(err, "patching existing object")
	}

	return patch, patchedObj, err
}

func (p *patcher) patch(current runtime.Object, modified []byte, source, namespace, name string, errOut io.Writer) ([]byte, runtime.Object, error) {
	var getErr error
	patchBytes, patchObject, err := p.patchSimple(current, modified, source, namespace, name, errOut)
	for i := 1; i <= maxPatchRetry && kerrors.IsConflict(err); i++ {
		if i > triesBeforeBackOff {
			p.backOff.Sleep(backOffPeriod)
		}
		current, getErr = p.helper.Get(namespace, name, false)
		if getErr != nil {
			return nil, nil, getErr
		}
		patchBytes, patchObject, err = p.patchSimple(current, modified, source, namespace, name, errOut)
	}
	if err != nil && kerrors.IsConflict(err) && p.force {
		patchBytes, patchObject, err = p.deleteAndCreate(current, modified, namespace, name)
	}
	return patchBytes, patchObject, err
}

func (p *patcher) deleteAndCreate(original runtime.Object, modified []byte, namespace, name string) ([]byte, runtime.Object, error) {
	err := p.delete(namespace, name)
	if err != nil {
		return modified, nil, err
	}
	err = wait.PollImmediate(kubectl.Interval, p.timeout, func() (bool, error) {
		if _, err := p.helper.Get(namespace, name, false); !kerrors.IsNotFound(err) {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return modified, nil, err
	}
	versionedObject, _, err := p.decoder.Decode(modified, nil, nil)
	if err != nil {
		return modified, nil, err
	}
	createdObject, err := p.helper.Create(namespace, true, versionedObject)
	if err != nil {
		// restore the original object if we fail to create the new one
		// but still propagate and advertise error to user
		recreated, recreateErr := p.helper.Create(namespace, true, original)
		if recreateErr != nil {
			err = fmt.Errorf("an error occurred force-replacing the existing object with the newly provided one:\n\n%v.\n\nAdditionally, an error occurred attempting to restore the original object:\n\n%v\n", err, recreateErr)
		} else {
			createdObject = recreated
		}
	}
	return modified, createdObject, err
}

func (p *patcher) delete(namespace, name string) error {
	c, err := p.clientFunc(p.mapping)
	if err != nil {
		return err
	}
	return runDelete(namespace, name, p.mapping, c, p.helper, p.cascade, p.gracePeriod, p.clientsetFunc)
}

func runDelete(namespace, name string, mapping *meta.RESTMapping, c resource.RESTClient, helper *resource.Helper, cascade bool, gracePeriod int, clientsetFunc func() (internalclientset.Interface, error)) error {
	if !cascade {
		if helper == nil {
			helper = resource.NewHelper(c, mapping)
		}
		return helper.Delete(namespace, name)
	}
	cs, err := clientsetFunc()
	if err != nil {
		return err
	}
	r, err := kubectl.ReaperFor(mapping.GroupVersionKind.GroupKind(), cs)
	if err != nil {
		if _, ok := err.(*kubectl.NoSuchReaperError); !ok {
			return err
		}
		return resource.NewHelper(c, mapping).Delete(namespace, name)
	}
	var options *metav1.DeleteOptions
	if gracePeriod >= 0 {
		options = metav1.NewDeleteOptions(int64(gracePeriod))
	}
	if err := r.Stop(namespace, name, 2*time.Minute, options); err != nil {
		return err
	}
	return nil
}
