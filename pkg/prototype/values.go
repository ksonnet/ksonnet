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

package prototype

import (
	"bufio"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

const (
	// OptionValuesFile is flag key containing the path to a values file
	OptionValuesFile = "values-file"
)

// Values extracts values for a prototype from supplied flags.
func Values(fs afero.Fs, p *Prototype, flags *pflag.FlagSet) (map[string]string, error) {
	if p == nil {
		return nil, errors.New("prototype is required")
	}

	if flags == nil {
		return nil, errors.New("flags is required")
	}

	valuesFile, err := flags.GetString(OptionValuesFile)
	if err != nil {
		return nil, err
	}

	name, err := flags.GetString("name")
	if err != nil {
		return nil, errors.New("name for prototype is required")
	}

	if valuesFile != "" {
		vff, err := newValuesFromFile(fs, p, name, valuesFile)
		if err != nil {
			return nil, err
		}

		return vff.Extract()
	}

	vff := newValuesFromFlags(p, flags)
	return vff.Extract()
}

type valuesFromFile struct {
	p    *Prototype
	name string
	r    io.Reader
}

func newValuesFromFile(fs afero.Fs, p *Prototype, name, path string) (*valuesFromFile, error) {
	if fs == nil {
		return nil, errors.New("fs is required")
	}

	if name == "" {
		return nil, errors.New("name is required")
	}

	if path == "" {
		return nil, errors.New("path to values file is required")
	}

	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}

	return &valuesFromFile{
		p:    p,
		name: name,
		r:    f,
	}, nil
}

func (vff *valuesFromFile) Extract() (map[string]string, error) {
	m := make(map[string]string)

	if vff.name == "" {
		return nil, errors.New("name for prototype is required")
	}

	values, err := vff.values()
	if err != nil {
		return nil, errors.Wrap(err, "reading values file")
	}

	values["name"] = vff.name

	missingRequired := ParamSchemas{}
	for _, param := range vff.p.Params {
		v := values[param.Name]
		if v == "" && param.IsRequired() {
			missingRequired = append(missingRequired, param)
		}

		quoted, err := param.QuotedValue(values[param.Name])
		if err != nil {
			return nil, err
		}

		m[param.Name] = quoted
	}

	if len(missingRequired) > 0 {
		return nil, errors.Errorf("failed to instantiate prototype '%s'. The following required parameters are missing:\n%s", vff.p.Name, missingRequired.PrettyString(""))
	}

	return m, nil
}

func (vff *valuesFromFile) values() (map[string]string, error) {
	m := make(map[string]string)

	scanner := bufio.NewScanner(vff.r)

	for scanner.Scan() {
		s := scanner.Text()

		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			return nil, errors.Errorf("%q is invalid. value format is key=value", s)
		}

		if parts[1] == "" {
			return nil, errors.Errorf("%q is invalid. value format is key=value", s)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		m[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanning error")
	}

	return m, nil
}

type valuesFromFlags struct {
	p     *Prototype
	flags *pflag.FlagSet
}

func newValuesFromFlags(p *Prototype, flags *pflag.FlagSet) *valuesFromFlags {
	return &valuesFromFlags{
		p:     p,
		flags: flags,
	}
}

func (vff *valuesFromFlags) Extract() (map[string]string, error) {
	values := map[string]string{}
	missingRequired := ParamSchemas{}

	for _, param := range vff.p.Params {
		val, err := vff.flags.GetString(param.Name)
		if err != nil {
			return nil, err
		}

		if param.Default != nil {
			if val == "" {
				missingRequired = append(missingRequired, param)
				continue
			}
		}

		if _, ok := values[param.Name]; ok {
			return nil, errors.Errorf("prototype %q has multiple parameters with name %q", vff.p.Name, param.Name)
		}

		quoted, err := param.Quote(val)
		if err != nil {
			return nil, err
		}
		values[param.Name] = quoted
	}

	if len(missingRequired) > 0 {
		return nil, errors.Errorf("failed to instantiate prototype %q. The following required parameters are missing:\n%s", vff.p.Name, missingRequired.PrettyString(""))
	}

	return values, nil
}

func BindFlags(p *Prototype) *pflag.FlagSet {
	fs := pflag.NewFlagSet("preview", pflag.ContinueOnError)

	fs.String(OptionValuesFile, "", "Supply a values file")

	for _, param := range p.RequiredParams() {
		fs.String(param.Name, "", param.Description)
	}

	for _, param := range p.OptionalParams() {
		fs.String(param.Name, *param.Default, param.Description)
	}

	return fs
}
