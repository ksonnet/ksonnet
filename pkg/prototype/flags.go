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
