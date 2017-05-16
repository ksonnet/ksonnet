package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	jsonnet "github.com/strickyak/jsonnet_cgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Read fetches and decodes K8s objects by path.
// TODO: Replace this with something supporting more sophisticated
// content negotiation.
func Read(vm *jsonnet.VM, path string) ([]runtime.Unstructured, error) {
	ext := filepath.Ext(path)
	if ext == ".json" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return jsonReader(f)
	} else if ext == ".yaml" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return yamlReader(f)
	} else if ext == ".jsonnet" {
		return jsonnetReader(vm, path)
	}

	return nil, fmt.Errorf("Unknown file extension: %s", path)
}

func jsonReader(r io.Reader) ([]runtime.Unstructured, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	obj, _, err := unstructuredv1.UnstructuredJSONScheme.Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	return []runtime.Unstructured{obj.(runtime.Unstructured)}, nil
}

func yamlReader(r io.ReadCloser) ([]runtime.Unstructured, error) {
	decoder := yaml.NewDocumentDecoder(r)
	ret := []runtime.Unstructured{}
	buf := []byte{}
	for {
		_, err := decoder.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		jsondata, err := yaml.ToJSON(buf)
		if err != nil {
			return nil, err
		}
		obj, _, err := unstructuredv1.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj.(runtime.Unstructured))
	}
	return ret, nil
}

func jsonnetReader(vm *jsonnet.VM, path string) ([]runtime.Unstructured, error) {
	jsonstr, err := vm.EvaluateFile(path)
	if err != nil {
		return nil, err
	}

	glog.V(4).Infof("jsonnet result is: %s\n", jsonstr)

	return jsonReader(strings.NewReader(jsonstr))
}

// FlattenToV1 expands any List-type objects into their members, and
// cooerces everything to metav1.Objects.  Panics if coercion
// encounters an unexpected object type.
func FlattenToV1(objs []runtime.Unstructured) []metav1.Object {
	ret := make([]metav1.Object, 0, len(objs))
	for _, obj := range objs {
		switch o := obj.(type) {
		case *unstructuredv1.UnstructuredList:
			for _, item := range o.Items {
				ret = append(ret, &item)
			}
		case *unstructuredv1.Unstructured:
			ret = append(ret, o)
		default:
			panic("Unexpected unstructured object type")
		}
	}
	return ret
}
