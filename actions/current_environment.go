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

import "github.com/pkg/errors"

type environmentMetadata interface {
	CurrentEnvironment() string
}

type currentEnver interface {
	setCurrentEnv(name string)
}

func setCurrentEnv(em environmentMetadata, ce currentEnver, ol *optionLoader) error {
	envName := ol.LoadOptionalString(OptionEnvName)
	if envName == "" {
		envName = em.CurrentEnvironment()
	}

	if envName == "" {
		return errors.Errorf("environment is not set; use `env list` to see available environments")
	}

	ce.setCurrentEnv(envName)
	return nil
}
