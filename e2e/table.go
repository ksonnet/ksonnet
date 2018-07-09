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

package e2e

import (
	"encoding/json"

	// gomega matchers
	// nolint: golint
	. "github.com/onsi/gomega"
)

type tableResponse struct {
	Kind string          `json:"kind,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

func loadTableResponse(s string) *tableResponse {
	b := []byte(s)
	var tr tableResponse
	err := json.Unmarshal(b, &tr)
	Expect(err).ToNot(HaveOccurred())

	return &tr
}

func (tr *tableResponse) data(kind string, dest interface{}) {
	Expect(tr.Kind).To(Equal(kind), "checking table kind")

	err := json.Unmarshal(tr.Data, dest)
	Expect(err).ToNot(HaveOccurred())
}

func (tr *tableResponse) componentList() []componentListRow {
	var rows []componentListRow
	tr.data("componentList", &rows)
	return rows
}

func (tr *tableResponse) envList() []envListRow {
	var rows []envListRow
	tr.data("envList", &rows)
	return rows
}

func (tr *tableResponse) paramList() []paramListRow {
	var rows []paramListRow
	tr.data("paramList", &rows)
	return rows
}

func (tr *tableResponse) pkgList() []pkgListRow {
	var rows []pkgListRow
	tr.data("pkgList", &rows)
	return rows
}

func (tr *tableResponse) prototypeList() []prototypeListRow {
	var rows []prototypeListRow
	tr.data("prototypeList", &rows)
	return rows
}

func (tr *tableResponse) registryList() []registryListRow {
	var rows []registryListRow
	tr.data("registryList", &rows)
	return rows
}

type componentListRow struct {
	APIVersion string `json:"apiversion,omitempty"`
	Component  string `json:"component,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	Type       string `json:"type,omitempty"`
}

type envListRow struct {
	KubernetesVersion string `json:"kubernetes-version,omitempty"`
	Name              string `json:"name,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	Override          string `json:"override,omitempty"`
	Server            string `json:"server,omitempty"`
}

type paramListRow struct {
	Component string `json:"component,omitempty"`
	Param     string `json:"param,omitempty"`
	Value     string `json:"value,omitempty"`
}

type pkgListRow struct {
	Registry  string `json:"registry,omitempty"`
	Name      string `json:"name,omitempty"`
	Version   string `json:"version,omitempty"`
	Installed string `json:"installed,omitempty"`
}

type prototypeListRow struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type registryListRow struct {
	Name     string `json:"name,omitempty"`
	Override string `json:"override,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	URI      string `json:"uri,omitempty"`
}
