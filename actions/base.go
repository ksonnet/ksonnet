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
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/spf13/afero"
)

type base struct {
	app app.App
}

func new(fs afero.Fs, appRoot string) (*base, error) {
	a, err := app.Load(fs, appRoot)
	if err != nil {
		return nil, err
	}

	return &base{
		app: a,
	}, nil
}
