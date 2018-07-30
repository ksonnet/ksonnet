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
	"path/filepath"
	"testing"

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/component"
	cmocks "github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestComponentList_wide(t *testing.T) {

	validComponentManager := func() component.Manager {
		summary1 := component.Summary{ComponentName: "ingress"}
		c1 := &cmocks.Component{}
		c1.On("Summarize").Return(summary1, nil)

		summary2 := component.Summary{ComponentName: "deployment"}
		c2 := &cmocks.Component{}
		c2.On("Summarize").Return(summary2, nil)

		cs := []component.Component{c1, c2}

		cm := &cmocks.Manager{}
		cm.On("Components", mock.Anything, "/").Return(cs, nil)

		return cm
	}

	cannotLoadComponents := func() component.Manager {
		cm := &cmocks.Manager{}
		cm.On("Components", mock.Anything, "/").Return(nil, errors.New("can't load components"))

		return cm
	}

	cannotLoadSummary := func() component.Manager {
		c1 := &cmocks.Component{}

		c1.On("Summarize").Return(component.Summary{}, errors.New("can't load summary"))

		cs := []component.Component{c1}

		cm := &cmocks.Manager{}
		cm.On("Components", mock.Anything, "/").Return(cs, nil)

		return cm
	}

	cases := []struct {
		name             string
		componentManager component.Manager
		output           string
		expectedFile     string
		isErr            bool
	}{
		{
			name:             "with json format",
			componentManager: validComponentManager(),
			output:           "json",
			expectedFile:     filepath.ToSlash("component/list/output.json"),
		},
		{
			name:             "with table format",
			componentManager: validComponentManager(),
			output:           "table",
			expectedFile:     filepath.ToSlash("component/list/table.txt"),
		},
		{
			name:             "with unspecified format",
			componentManager: validComponentManager(),
			expectedFile:     filepath.ToSlash("component/list/table.txt"),
		},
		{
			name:             "with invalid format",
			componentManager: validComponentManager(),
			output:           "invalid",
			isErr:            true,
		},
		{
			name:             "can't load components",
			componentManager: cannotLoadComponents(),
			isErr:            true,
		},
		{
			name:             "can't load summary",
			componentManager: cannotLoadSummary(),
			isErr:            true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				moduleName := "/"

				in := map[string]interface{}{
					OptionApp:    appMock,
					OptionModule: moduleName,
					OptionOutput: tc.output,
				}

				a, err := NewComponentList(in)
				require.NoError(t, err)

				a.cm = tc.componentManager

				var buf bytes.Buffer
				a.out = &buf

				err = a.Run()
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				assertOutput(t, tc.expectedFile, buf.String())
			})
		})
	}

}

func TestComponentList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewComponentList(in)
	require.Error(t, err)
}
