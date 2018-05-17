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
	"io"
	"strings"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	cases := []struct {
		name       string
		src1       string
		src2       string
		eLocation1 string
		eLocation2 string
		isNewError bool
		isRunError bool
	}{
		{
			name:       "default",
			src1:       "default",
			eLocation1: "local:default",
			eLocation2: "remote:default",
		},
		{
			name:       "local:default remote:default",
			src1:       "local:default",
			src2:       "remote:default",
			eLocation1: "local:default",
			eLocation2: "remote:default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				in := map[string]interface{}{
					OptionApp:            appMock,
					OptionClientConfig:   &client.Config{},
					OptionComponentNames: []string{},
					OptionSrc1:           tc.src1,
					OptionSrc2:           tc.src2,
				}

				d, err := NewDiff(in)
				if tc.isNewError {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)

				var buf bytes.Buffer
				d.out = &buf

				d.diffFn = func(a app.App, c *client.Config, l1 *diff.Location, l2 *diff.Location) (io.Reader, error) {
					assert.Equal(t, tc.eLocation1, l1.String(), "location1")
					assert.Equal(t, tc.eLocation2, l2.String(), "location2")

					r := strings.NewReader("")
					return r, nil
				}

				err = d.Run()
				if tc.isRunError {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
			})
		})
	}
}
