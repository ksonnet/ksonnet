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
	"github.com/ksonnet/ksonnet/client"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/cluster"
)

type runDeleteFn func(cluster.DeleteConfig, ...cluster.DeleteOpts) error

// RunDelete runs `delete`.
func RunDelete(m map[string]interface{}) error {
	a, err := newDelete(m)
	if err != nil {
		return err
	}

	return a.run()
}

type deleteOpt func(*Delete)

// Delete collects options for applying objects to a cluster.
type Delete struct {
	app            app.App
	clientConfig   *client.Config
	componentNames []string
	envName        string
	gracePeriod    int64

	runDeleteFn runDeleteFn
}

// RunDelete runs `apply`
func newDelete(m map[string]interface{}, opts ...deleteOpt) (*Delete, error) {
	ol := newOptionLoader(m)

	a := &Delete{
		app:            ol.loadApp(),
		clientConfig:   ol.loadClientConfig(),
		componentNames: ol.loadStringSlice(OptionComponentNames),
		envName:        ol.loadString(OptionEnvName),
		gracePeriod:    ol.loadInt64(OptionGracePeriod),

		runDeleteFn: cluster.RunDelete,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	for _, opt := range opts {
		opt(a)
	}

	return a, nil
}

func (a *Delete) run() error {
	config := cluster.DeleteConfig{
		App:            a.app,
		ClientConfig:   a.clientConfig,
		ComponentNames: a.componentNames,
		EnvName:        a.envName,
		GracePeriod:    a.gracePeriod,
	}

	return a.runDeleteFn(config)
}
