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
	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
)

// RunNsCreate creates a namespace.
func RunNsCreate(m map[string]interface{}) error {
	nc, err := NewNsCreate(m)
	if err != nil {
		return err
	}

	return nc.Run()
}

// NsCreate creates a component namespace
type NsCreate struct {
	app    app.App
	nsName string
	cm     component.Manager
}

// NewNsCreate creates an instance of NsCreate.
func NewNsCreate(m map[string]interface{}) (*NsCreate, error) {
	ol := newOptionLoader(m)

	et := &NsCreate{
		app:    ol.loadApp(),
		nsName: ol.loadString(OptionNamespaceName),

		cm: component.DefaultManager,
	}

	return et, nil
}

// Run runs that ns create action.
func (nc *NsCreate) Run() error {
	_, err := nc.cm.Namespace(nc.app, nc.nsName)
	if err == nil {
		return errors.Errorf("namespace %q already exists", nc.nsName)
	}

	return nc.cm.CreateNamespace(nc.app, nc.nsName)
}
