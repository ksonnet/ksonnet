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

	"github.com/ksonnet/ksonnet/component"
	cmocks "github.com/ksonnet/ksonnet/component/mocks"
	amocks "github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestComponentList(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		nsName := ""
		output := ""

		c := &cmocks.Component{}
		c.On("Name", false).Return("c1")

		cs := []component.Component{c}

		ns := &cmocks.Namespace{}
		ns.On("Components").Return(cs, nil)

		cm := &cmocks.Manager{}
		cm.On("Namespace", mock.Anything, "").Return(ns, nil)

		a, err := NewComponentList(appMock, nsName, output)
		require.NoError(t, err)

		a.cm = cm

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, "component/list/output.txt", buf.String())
	})
}

func TestComponentList_wide(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		nsName := ""
		output := "wide"

		summary := []component.Summary{
			{ComponentName: "deployment"},
		}
		c := &cmocks.Component{}
		c.On("Summarize").Return(summary, nil)

		cs := []component.Component{c}

		ns := &cmocks.Namespace{}
		ns.On("Components").Return(cs, nil)

		cm := &cmocks.Manager{}
		cm.On("Namespace", mock.Anything, "").Return(ns, nil)

		a, err := NewComponentList(appMock, nsName, output)
		require.NoError(t, err)

		a.cm = cm

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, "component/list/wide.txt", buf.String())
	})
}
