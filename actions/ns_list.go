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
	"io"
	"os"
	"sort"

	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/util/table"
)

// RunNsList runs `ns list`
func RunNsList(m map[string]interface{}) error {
	nl, err := NewNsList(m)
	if err != nil {
		return err
	}

	return nl.Run()
}

// NsList lists namespaces.
type NsList struct {
	app     app.App
	envName string
	out     io.Writer
	cm      component.Manager
}

// NewNsList creates an instance of NsList.
func NewNsList(m map[string]interface{}) (*NsList, error) {
	ol := newOptionLoader(m)

	nl := &NsList{
		app:     ol.loadApp(),
		envName: ol.loadString(OptionEnvName),

		out: os.Stdout,
		cm:  component.DefaultManager,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return nl, nil
}

// Run lists namespaces.
func (nl *NsList) Run() error {
	namespaces, err := nl.cm.Namespaces(nl.app, nl.envName)
	if err != nil {
		return err
	}

	t := table.New(nl.out)
	t.SetHeader([]string{"namespace"})

	names := make([]string, len(namespaces))
	for i := range namespaces {
		names[i] = namespaces[i].Name()
	}

	sort.Strings(names)

	for _, name := range names {
		t.Append([]string{name})
	}

	return t.Render()
}
