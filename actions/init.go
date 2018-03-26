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

package actions

import (
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/appinit"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/spf13/afero"
)

const (
	defaultIncubatorRegName = "incubator"
	defaultIncubatorURI     = "github.com/ksonnet/parts/tree/master/" + defaultIncubatorRegName
)

// RunInit creates a namespace.
func RunInit(fs afero.Fs, name, rootPath, k8sSpecFlag, serverURI, namespace string) error {
	i, err := NewInit(fs, name, rootPath, k8sSpecFlag, serverURI, namespace)
	if err != nil {
		return err
	}

	return i.Run()
}

type appInitFn func(fs afero.Fs, name, rootPath, k8sSpecFlag, serverURI, namespace string, registries []registry.Registry) error

// Init creates a component namespace
type Init struct {
	fs          afero.Fs
	name        string
	rootPath    string
	k8sSpecFlag string
	serverURI   string
	namespace   string

	appInitFn appInitFn
}

// NewInit creates an instance of Init.
func NewInit(fs afero.Fs, name, rootPath, k8sSpecFlag, serverURI, namespace string) (*Init, error) {
	i := &Init{
		fs:          fs,
		name:        name,
		rootPath:    rootPath,
		k8sSpecFlag: k8sSpecFlag,
		serverURI:   serverURI,
		namespace:   namespace,
		appInitFn:   appinit.Init,
	}

	return i, nil
}

// Run runs that ns create action.
func (i *Init) Run() error {
	gh, err := registry.NewGitHub(&app.RegistryRefSpec{
		Name:     "incubator",
		Protocol: registry.ProtocolGitHub,
		URI:      defaultIncubatorURI,
	})
	if err != nil {
		return err
	}

	registries := []registry.Registry{gh}

	return i.appInitFn(
		i.fs,
		i.name,
		i.rootPath,
		i.k8sSpecFlag,
		i.serverURI,
		i.namespace,
		registries,
	)

}
