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
	"github.com/ksonnet/ksonnet/pkg/component"
	cmocks "github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestComponentList(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		module := ""
		output := ""

		c := &cmocks.Component{}
		c.On("Name", false).Return("c1")

		cs := []component.Component{c}

		ns := &cmocks.Module{}
		ns.On("Components").Return(cs, nil)

		cm := &cmocks.Manager{}
		cm.On("Module", mock.Anything, "").Return(ns, nil)

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionOutput: output,
		}

		a, err := NewComponentList(in)
		require.NoError(t, err)

		a.cm = cm

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, "component/list/output.txt", buf.String())
	})
}

func TestComponentList_json(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		module := ""
		output := "json"

		summary1 := component.Summary{ComponentName: "ingress"}
		c1 := &cmocks.Component{}
		c1.On("Summarize").Return(summary1, nil)

		summary2 := component.Summary{ComponentName: "deployment"}
		c2 := &cmocks.Component{}
		c2.On("Summarize").Return(summary2, nil)

		cs := []component.Component{c1, c2}

		ns := &cmocks.Module{}
		ns.On("Components").Return(cs, nil)

		cm := &cmocks.Manager{}
		cm.On("Module", mock.Anything, "").Return(ns, nil)

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionOutput: output,
		}

		a, err := NewComponentList(in)
		require.NoError(t, err)

		a.cm = cm

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, "component/list/json.txt", buf.String())
	})
}

func TestComponentList_wide(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		module := ""
		output := "wide"

		summary1 := component.Summary{ComponentName: "ingress"}
		c1 := &cmocks.Component{}
		c1.On("Summarize").Return(summary1, nil)

		summary2 := component.Summary{ComponentName: "deployment"}
		c2 := &cmocks.Component{}
		c2.On("Summarize").Return(summary2, nil)

		cs := []component.Component{c1, c2}

		ns := &cmocks.Module{}
		ns.On("Components").Return(cs, nil)

		cm := &cmocks.Manager{}
		cm.On("Module", mock.Anything, "").Return(ns, nil)

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionOutput: output,
		}

		a, err := NewComponentList(in)
		require.NoError(t, err)

		a.cm = cm

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, "component/list/wide.txt", buf.String())
	})
}

func TestComponentList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewComponentList(in)
	require.Error(t, err)
}
