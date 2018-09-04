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
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/component"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/ksonnet/ksonnet/pkg/schema"
	utilstrings "github.com/ksonnet/ksonnet/pkg/util/strings"
	utilyaml "github.com/ksonnet/ksonnet/pkg/util/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// RunImport runs `import`
func RunImport(m map[string]interface{}) error {
	i, err := NewImport(m)
	if err != nil {
		return err
	}

	return i.Run()
}

// Import imports files or directories into ksonnet.
type Import struct {
	app    app.App
	module string
	path   string

	createComponentFn func(a app.App, module, name, text string, p params.Params, templateType prototype.TemplateType) (string, error)
}

// NewImport creates an instance of Import. `module` is the name of the component and
// entity is the file or directory to import.
func NewImport(m map[string]interface{}) (*Import, error) {
	ol := newOptionLoader(m)

	i := &Import{
		app:    ol.LoadApp(),
		module: ol.LoadString(OptionModule),
		path:   ol.LoadString(OptionPath),

		createComponentFn: component.Create,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return i, nil
}

// Run runs the import process.
func (i *Import) Run() error {
	if i.path == "" {
		return errors.New("path is required")
	}

	if strings.HasPrefix(i.path, "http") {
		return i.handleURL()
	}

	return i.handleLocal()
}

func (i *Import) handleURL() error {
	resp, err := http.Get(i.path)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unable to download %s: %s", i.path, resp.Status)
	}

	dir, err := afero.TempDir(i.app.Fs(), "", "")
	if err != nil {
		return err
	}

	defer i.app.Fs().RemoveAll(dir)

	filename, err := extractFilename(resp)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, filename)
	f, err := i.app.Fs().Create(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return i.importFile(path)
}

func extractFilename(resp *http.Response) (string, error) {
	filename := resp.Request.URL.Path
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		if _, contentDisposition, err := mime.ParseMediaType(cd); err == nil {
			filename = contentDisposition["filename"]
		}
	}

	if filename == "" || strings.HasSuffix(filename, "/") || strings.Contains(filename, "\x00") {
		return "", errors.New("unable to find name for file")
	}

	filename = filepath.Base(path.Clean("/" + filename))
	if filename == "" || filename == "." || filename == "/" {
		return "", errors.New("unable to find name for file in path")
	}

	return filename, nil
}

func (i *Import) handleLocal() error {
	pathFi, err := i.app.Fs().Stat(i.path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("%s does not exist", i.path)
		}
		return err
	}

	var paths []string
	if pathFi.IsDir() {
		fis, err := afero.ReadDir(i.app.Fs(), i.path)
		if err != nil {
			return err
		}

		for _, fi := range fis {
			path := filepath.Join(i.path, fi.Name())
			paths = append(paths, path)
		}
	} else {
		paths = append(paths, i.path)
	}

	for _, path := range paths {
		if err := i.importFile(path); err != nil {
			return err
		}
	}

	return nil
}

func (i *Import) importFile(fileName string) error {
	base := filepath.Base(fileName)
	ext := filepath.Ext(base)

	templateType, err := prototype.ParseTemplateType(strings.TrimPrefix(ext, "."))
	if err != nil {
		return errors.Wrap(err, "parse template type")
	}

	switch templateType {
	default:
		return errors.Errorf("unable to handle components of type %s", templateType)
	case prototype.YAML:
		return i.createYAML(fileName, base, ext)
	case prototype.JSON, prototype.Jsonnet:
		return i.createComponent(fileName, base, ext, templateType)
	}
}

func (i *Import) createYAML(fileName, base, ext string) error {
	f, err := i.app.Fs().Open(fileName)
	if err != nil {
		return errors.Wrapf(err, "opening %q", fileName)
	}

	readers, err := utilyaml.Decode(f)
	if err != nil {
		return err
	}

	for _, r := range readers {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		dataReader := bytes.NewReader(data)
		ts, props, err := schema.ImportYaml(dataReader)
		if err != nil {
			if err == schema.ErrEmptyYAML {
				continue
			}
			return err
		}

		val, err := props.Value([]string{"metadata", "name"})
		if err != nil {
			return err
		}

		name, ok := val.(string)
		if !ok {
			return errors.Errorf("unable to find metadata name of object in %s", fileName)
		}

		componentName := fmt.Sprintf("%s-%s-%s", strings.ToLower(ts.Kind()), name, utilstrings.LowerRand(5))
		if err = i.createComponentFromData(componentName, string(data), prototype.YAML); err != nil {
			return err
		}
	}

	return nil
}

func (i *Import) createComponentFromData(name, data string, templateType prototype.TemplateType) error {
	componentParams := params.Params{}

	var moduleName string
	switch i.module {
	case "", "/":
		// do nothing
	default:
		moduleName = i.module
	}

	_, err := i.createComponentFn(i.app, moduleName, name, data, componentParams, templateType)
	if err != nil {
		return errors.Wrap(err, "create component")
	}

	return nil
}

func (i *Import) createComponent(fileName, base, ext string, templateType prototype.TemplateType) error {
	var name bytes.Buffer

	name.WriteString(strings.TrimSuffix(base, ext))

	contents, err := afero.ReadFile(i.app.Fs(), fileName)
	if err != nil {
		return errors.Wrap(err, "read manifest")
	}

	sourcePath := filepath.Clean(name.String())

	componentParams := params.Params{}

	_, err = i.createComponentFn(i.app, i.module, sourcePath, string(contents), componentParams, templateType)
	if err != nil {
		return errors.Wrap(err, "create component")
	}

	return nil

}
