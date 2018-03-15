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
	"os"

	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// ParamList lists params.
func ParamList(fs afero.Fs, appRoot, componentName, nsName, envName string) error {
	pl, err := newParamList(fs, appRoot, componentName, nsName, envName)
	if err != nil {
		return err
	}

	return pl.run()
}

type paramList struct {
	nsName        string
	componentName string
	envName       string

	*base
}

func newParamList(fs afero.Fs, appRoot, componentName, nsName, envName string) (*paramList, error) {
	b, err := new(fs, appRoot)
	if err != nil {
		return nil, err
	}

	pl := &paramList{
		nsName:        nsName,
		componentName: componentName,
		envName:       envName,
		base:          b,
	}

	return pl, nil
}

func (pl *paramList) run() error {
	// if you want env params, call app.EnvParams for the name space.
	// then you could convert that into the list

	ns, err := component.GetNamespace(pl.app, pl.nsName)
	if err != nil {
		return errors.Wrap(err, "could not find namespace")
	}

	var params []component.NamespaceParameter
	if pl.componentName == "" {
		cParams, err := ns.Params(pl.envName)
		if err != nil {
			return err
		}
		params = append(params, cParams...)
	} else {
		dm := component.DefaultManager
		c, err := dm.Component(pl.app, pl.nsName, pl.componentName)
		if err != nil {
			return err
		}

		cParams, err := c.Params(pl.envName)
		if err != nil {
			return err
		}
		params = append(params, cParams...)
	}

	table := table.New(os.Stdout)

	table.SetHeader([]string{"COMPONENT", "INDEX", "PARAM", "VALUE"})
	for _, data := range params {
		table.Append([]string{data.Component, data.Index, data.Key, data.Value})
	}

	table.Render()

	return nil
}
