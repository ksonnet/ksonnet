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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/ksonnet"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/ksonnet/ksonnet/pkg/prototype/snippet"
	"github.com/ksonnet/ksonnet/pkg/prototype/snippet/jsonnet"
	"github.com/ksonnet/ksonnet/pkg/registry"
	strutil "github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

// RunPrototypePreview runs `prototype describe`
func RunPrototypePreview(m map[string]interface{}) error {
	pp, err := NewPrototypePreview(m)
	if err != nil {
		return err
	}

	return pp.Run()
}

// PrototypePreview lists available namespaces
type PrototypePreview struct {
	app   app.App
	out   io.Writer
	query string
	args  []string

	appPrototypesFn     func(app.App, pkg.Descriptor) (prototype.Prototypes, error)
	bindFlagsFn         func(p *prototype.Prototype) (*pflag.FlagSet, error)
	packageManager      registry.PackageManager
	extractParametersFn func(fs afero.Fs, p *prototype.Prototype, f *pflag.FlagSet) (map[string]string, error)
}

// NewPrototypePreview creates an instance of PrototypePreview
func NewPrototypePreview(m map[string]interface{}) (*PrototypePreview, error) {
	ol := newOptionLoader(m)

	app := ol.LoadApp()
	httpClientOpt := registry.HTTPClientOpt(ol.LoadHTTPClient())

	pp := &PrototypePreview{
		app:   app,
		query: ol.LoadString(OptionQuery),
		args:  ol.LoadStringSlice(OptionArguments),

		out:                 os.Stdout,
		packageManager:      registry.NewPackageManager(app, httpClientOpt),
		bindFlagsFn:         prototype.BindFlags,
		extractParametersFn: prototype.ExtractParameters,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return pp, nil
}

// Run runs the env list action.
func (pp *PrototypePreview) Run() error {
	prototypes, err := pp.packageManager.Prototypes()
	if err != nil {
		return err
	}

	index, err := prototype.NewIndex(prototypes, prototype.DefaultBuilder)
	if err != nil {
		return err
	}

	prototypes, err = index.List()
	if err != nil {
		return err
	}

	p, err := findUniquePrototype(pp.query, prototypes)
	if err != nil {
		return err
	}

	flags, err := pp.bindFlagsFn(p)
	if err != nil {
		return errors.Wrap(err, "binding prototype flags")
	}

	if err = flags.Parse(pp.args); err != nil {
		if strings.Contains(err.Error(), "help request") {
			return nil
		}
		return errors.Wrap(err, "parse preview args")
	}

	// NOTE: only supporting jsonnet templates
	templateType := prototype.Jsonnet

	params, err := pp.extractParametersFn(pp.app.Fs(), p, flags)
	if err != nil {
		return err
	}

	text, err := expandPrototype(p, templateType, params, "preview")
	if err != nil {
		return err
	}

	fmt.Fprintln(pp.out, text)
	return nil
}

// TODO: this doesn't belong here. Needs to be closer to where other jsonnet processing happens.
func expandPrototype(proto *prototype.Prototype, templateType prototype.TemplateType, params map[string]string, componentName string) (string, error) {
	template, err := proto.Template.Body(templateType)
	if err != nil {
		return "", err
	}
	if templateType == prototype.Jsonnet {
		componentsText := "components." + componentName
		if !strutil.IsASCIIIdentifier(componentName) {
			componentsText = fmt.Sprintf(`components["%s"]`, componentName)
		}
		template = append([]string{
			`local env = std.extVar("` + ksonnet.EnvExtCodeKey + `");`,
			`local params = std.extVar("` + ksonnet.ParamsExtCodeKey + `").` + componentsText + ";"},
			template...)
		return jsonnet.Parse(componentName, strings.Join(template, "\n"))
	}

	tm := snippet.Parse(strings.Join(template, "\n"))
	return tm.Evaluate(params)
}
