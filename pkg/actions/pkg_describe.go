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
	"text/template"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/registry"
)

// RunPkgDescribe runs `pkg install`
func RunPkgDescribe(m map[string]interface{}) error {
	pd, err := NewPkgDescribe(m)
	if err != nil {
		return err
	}

	return pd.Run()
}

// PkgDescribe describes a package.
type PkgDescribe struct {
	app     app.App
	pkgName string

	templateSrc    string
	out            io.Writer
	packageManager registry.PackageManager
}

// NewPkgDescribe creates an instance of PkgDescribe.
func NewPkgDescribe(m map[string]interface{}) (*PkgDescribe, error) {
	ol := newOptionLoader(m)

	app := ol.LoadApp()
	pd := &PkgDescribe{
		app:     app,
		pkgName: ol.LoadString(OptionPackageName),

		templateSrc:    pkgDescribeTemplate,
		out:            os.Stdout,
		packageManager: registry.NewPackageManager(app),
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return pd, nil
}

// Run describes a package.
func (pd *PkgDescribe) Run() error {
	p, err := pd.packageManager.Find(pd.pkgName)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Name":        pd.pkgName,
		"Description": p.Description(),
	}

	isInstalled, err := p.IsInstalled()
	if err != nil {
		return err
	}

	data["IsInstalled"] = isInstalled

	if isInstalled {
		prototypes, err := p.Prototypes()
		if err != nil {
			return err
		}

		data["Prototypes"] = prototypes
	}

	t, err := template.New("pkg-describe").Parse(pd.templateSrc)
	if err != nil {
		return err
	}

	if err = t.Execute(pd.out, data); err != nil {
		return err
	}

	return nil
}

const pkgDescribeTemplate = `LIBRARY NAME:
{{.Name}}

DESCRIPTION:
{{.Description}}

{{- if .IsInstalled}}

PROTOTYPES:{{- range .Prototypes}}
  {{.Name}} - {{.Template.ShortDescription}}
{{- end}}{{- end}}
`
