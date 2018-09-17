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

package params

import (
	"encoding/json"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
)

// JsonnetEnvObject creates an object with the current ksonnet environment.
// This object includes the current server and namespace. The object
// is suitable to use as a Jsonnet ext code option.
func JsonnetEnvObject(a app.App, envName string) (string, error) {
	envDetails, err := a.Environment(envName)
	if err != nil {
		return "", err
	}
	if envDetails.Destination == nil {
		return "", errors.Errorf("environment lacks destination: %s", envName)
	}

	dest := map[string]string{
		"server":    envDetails.Destination.Server,
		"namespace": envDetails.Destination.Namespace,
	}

	marshalledDestination, err := json.Marshal(&dest)
	if err != nil {
		return "", err
	}

	return string(marshalledDestination), nil
}
