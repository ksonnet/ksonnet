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
	"fmt"

	"github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

// FlagDefinitionError is an error returned when a flag
// definition fails.
type FlagDefinitionError struct {
	name string
}

func (e *FlagDefinitionError) Error() string {
	return fmt.Sprintf("unable to define flag %q", e.name)
}

// BindFlags creates a flag set using a prototype's parameters.
func BindFlags(p *Prototype) (fs *pflag.FlagSet, err error) {
	fs = pflag.NewFlagSet("prototype-flags", pflag.ContinueOnError)

	fs.String("values-file", "", "Prototype values file (file returns a Jsonnet object)")
	fs.String("module", "", "Component module")

	for _, param := range p.RequiredParams() {
		if fs.Lookup(param.Name) != nil {
			return nil, &FlagDefinitionError{name: param.Name}
		}
		fs.String(param.Name, "", param.Description)
	}

	for _, param := range p.OptionalParams() {
		if fs.Lookup(param.Name) != nil {
			return nil, &FlagDefinitionError{name: param.Name}
		}
		fs.String(param.Name, *param.Default, param.Description)
	}

	return fs, nil
}

// ExtractParameters extracts prototypes parameters from flags.
func ExtractParameters(fs afero.Fs, p *Prototype, flags *pflag.FlagSet) (map[string]string, error) {

	required := make(map[string]*ParamSchema)

	values := map[string]string{}
	for _, param := range p.Params {
		if err := updateValuesFromFlag(p, values, param, flags); err != nil {
			return nil, errors.Wrap(err, "reading value from flag")
		}

		if param.Default == nil {
			required[param.Name] = param
		}
	}

	valuesFilePath, err := flags.GetString("values-file")
	if err != nil {
		return nil, errors.Wrap(err, "finding values file flag")
	}

	if valuesFilePath != "" {
		updateValuesFromValuesFile(fs, values, valuesFilePath)
	}

	if err = checkMissingParameters(p, values, required); err != nil {
		return nil, err
	}

	return values, nil
}

// updateValuesFromFlag updates values from flags. It mutates the map which is passed in.
func updateValuesFromFlag(p *Prototype, values map[string]string, param *ParamSchema, flags *pflag.FlagSet) error {
	val, err := flags.GetString(param.Name)
	if err != nil {
		return err
	} else if _, ok := values[param.Name]; ok {
		return errors.Errorf("prototype %q has multiple parameters with name %q", p.Name, param.Name)
	}

	quoted, err := param.Quote(val)
	if err != nil {
		return err
	}
	values[param.Name] = quoted

	return nil
}

// updateValuesFromValuesFile updates values from a values file. It mutates the map which is passed in.
func updateValuesFromValuesFile(fs afero.Fs, values map[string]string, valuesFilePath string) error {
	f, err := fs.Open(valuesFilePath)
	if err != nil {
		return errors.Wrap(err, "opening values file")
	}

	defer f.Close()

	vf, err := ReadValues(f)
	if err != nil {
		return errors.Wrap(err, "reading values file")
	}

	keys, err := vf.Keys()
	if err != nil {
		return errors.Wrap(err, "finding keys in values file")
	}

	for _, k := range keys {
		v, err := vf.Get(k)
		if err != nil {
			return errors.Wrapf(err, "retrieving %q from values file", k)
		}
		values[k] = v
	}

	return nil
}

func checkMissingParameters(p *Prototype, values map[string]string, required map[string]*ParamSchema) error {
	missingRequired := ParamSchemas{}

	var keys []string
	for k := range values {
		keys = append(keys, k)
	}

	for k, v := range required {
		if strings.InSlice(k, keys) && values[k] == `""` {
			missingRequired = append(missingRequired, v)
		}
	}

	if len(missingRequired) > 0 {
		return errors.Errorf("failed to instantiate prototype %q. The following required parameters are missing:\n%s", p.Name, missingRequired.PrettyString(""))
	}

	return nil
}
