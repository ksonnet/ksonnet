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
	"bytes"
	"testing"

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	registrymocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/stretchr/testify/require"
)

func TestPrototypeDescribe(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		prototypes := prototype.Prototypes{}

		manager := &registrymocks.PackageManager{}
		manager.On("Prototypes").Return(prototypes, nil)

		in := map[string]interface{}{
			OptionApp:           appMock,
			OptionQuery:         "namespace",
			OptionTLSSkipVerify: false,
		}

		a, err := NewPrototypeDescribe(in)
		require.NoError(t, err)

		a.packageManager = manager

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, "prototype/describe/output.txt", buf.String())
	})
}

func TestPrototypeDescribe_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPrototypeDescribe(in)
	require.Error(t, err)
}
