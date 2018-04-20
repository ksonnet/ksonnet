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

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
)

// RunEnvCurrent runs `env current`.
func RunEnvCurrent(m map[string]interface{}) error {
	a, err := newEnvCurrent(m)
	if err != nil {
		return err
	}

	return a.run()
}

// EnvCurrent sets/unsets the current environment
type EnvCurrent struct {
	app     app.App
	envName string
	unset   bool

	out io.Writer
}

// RunEnvCurrent runs `env current`
func newEnvCurrent(m map[string]interface{}) (*EnvCurrent, error) {
	ol := newOptionLoader(m)

	d := &EnvCurrent{
		app:     ol.LoadApp(),
		envName: ol.LoadOptionalString(OptionEnvName),
		unset:   ol.LoadBool(OptionUnset),

		out: os.Stdout,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return d, nil
}

func (e *EnvCurrent) run() error {
	if e.envName != "" && e.unset == true {
		return errors.New("set and unset are exclusive")
	}

	if e.unset {
		return e.app.SetCurrentEnvironment("")
	} else if e.envName != "" {
		return e.app.SetCurrentEnvironment(e.envName)
	} else {
		fmt.Fprintln(e.out, e.app.CurrentEnvironment())
		return nil
	}
}
