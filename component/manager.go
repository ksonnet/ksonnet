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

package component

import (
	"github.com/ksonnet/ksonnet/metadata/app"
)

var (
	// DefaultManager is the default manager for components.
	DefaultManager = &defaultManager{}
)

// Manager is an interface for interating with components.
type Manager interface {
	Namespaces(ksApp app.App, envName string) ([]Namespace, error)
	Namespace(ksApp app.App, nsName string) (Namespace, error)
	NSResolveParams(ns Namespace) (string, error)
	Components(ns Namespace) ([]Component, error)
	Component(ksApp app.App, nsName, componentName string) (Component, error)
}

type defaultManager struct{}

func (dm *defaultManager) Namespaces(ksApp app.App, envName string) ([]Namespace, error) {
	return NamespacesFromEnv(ksApp, envName)
}

func (dm *defaultManager) Namespace(ksApp app.App, nsName string) (Namespace, error) {
	return GetNamespace(ksApp, nsName)
}

func (dm *defaultManager) NSResolveParams(ns Namespace) (string, error) {
	return ns.ResolvedParams()
}

func (dm *defaultManager) Components(ns Namespace) ([]Component, error) {
	return ns.Components()
}

func (dm *defaultManager) Component(ksApp app.App, nsName, componentName string) (Component, error) {
	return LocateComponent(ksApp, nsName, componentName)
}
