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
	"io"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/upgrade"
	"github.com/stretchr/testify/require"
)

func TestUpgrade(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionDryRun: true,
		}

		var called bool
		u, err := newUpgrade(in)
		u.upgradeFn = func(a app.App, out io.Writer, pl upgrade.PackageLister, dryRun bool) error {
			called = true
			return nil
		}

		require.NoError(t, err)

		err = u.run()
		require.NoError(t, err)
		require.True(t, called)
	})
}

func TestUpgrade_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := newUpgrade(in)
	require.Error(t, err)
}
