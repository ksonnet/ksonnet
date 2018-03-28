// Copyright 2018 The kubecfg authors
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

package cmd

import (
	"testing"

	"github.com/ksonnet/ksonnet/client"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validateCmd(t *testing.T) {
	override := func(ksApp app.App, envName, nsName string, componentNames []string, clientConfig *client.Config) error {
		assert.Equal(t, "my-namespace", envName)
		assert.Equal(t, "", nsName)

		exectedComponents := []string{"module1", "module2"}
		assert.Equal(t, exectedComponents, componentNames)
		assert.Equal(t, validateClientConfig, clientConfig)

		return nil
	}

	withCmd(t, validateCmd, actionValidate, override, func() {
		args := []string{"validate", "my-namespace", "-c", "module1", "-c", "module2"}
		RootCmd.SetArgs(args)

		err := RootCmd.Execute()
		require.NoError(t, err)
	})
}
