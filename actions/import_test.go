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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/ksonnet/ksonnet/metadata/app"
	amocks "github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/stretchr/testify/require"
)

func TestImport_http(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "file.yaml")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		f, err := os.Open(dataPath)
		require.NoError(t, err)

		defer f.Close()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Disposition", `attachment; filename="manifest.yaml"`)
			http.ServeContent(w, r, "file.yaml", time.Time{}, f)
		}))
		defer ts.Close()

		module := "/"

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   ts.URL,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			assert.Equal(t, "/service-my-service", name)
			assert.Equal(t, string(serviceData), text)
			assert.Equal(t, params.Params{}, p)
			assert.Equal(t, prototype.YAML, templateType)

			return "/", nil
		}

		err = a.Run()
		require.NoError(t, err)

	})
}

func TestImport_file(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "file.yaml")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		module := "/"
		path := "/file.yaml"

		stageFile(t, appMock.Fs(), "import/file.yaml", path)

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   path,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			assert.Equal(t, "/service-my-service", name)
			assert.Equal(t, string(serviceData), text)
			assert.Equal(t, params.Params{}, p)
			assert.Equal(t, prototype.YAML, templateType)

			return "/", nil
		}

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestImport_directory(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		dataPath := filepath.Join("testdata", "import", "file.yaml")
		serviceData, err := ioutil.ReadFile(dataPath)
		require.NoError(t, err)

		module := "/"
		path := "/import"

		stageFile(t, appMock.Fs(), "import/file.yaml", "/import/file.yaml")

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   path,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			assert.Equal(t, "/service-my-service", name)
			assert.Equal(t, string(serviceData), text)
			assert.Equal(t, params.Params{}, p)
			assert.Equal(t, prototype.YAML, templateType)

			return "/", nil
		}

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestImport_invalid_file(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		module := "/"
		path := "/import"

		in := map[string]interface{}{
			OptionApp:    appMock,
			OptionModule: module,
			OptionPath:   path,
		}

		a, err := NewImport(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, name, text string, p params.Params, templateType prototype.TemplateType) (string, error) {
			return "", errors.New("invalid")
		}

		err = a.Run()
		require.Error(t, err)
	})
}
