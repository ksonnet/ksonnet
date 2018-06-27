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

package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	goyaml "github.com/ghodss/yaml"
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet/pkg/app"
	ksstrings "github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Renderer renders helm charts.
type Renderer struct {
	app app.App
}

// NewRenderer creates an instance of Renderer.
func NewRenderer(a app.App) *Renderer {
	return &Renderer{
		app: a,
	}
}

// JsonnetNativeFunc is a jsonnet native function that renders helm charts.
func (r *Renderer) JsonnetNativeFunc() *jsonnet.NativeFunction {
	fn := func(input []interface{}) (interface{}, error) {
		repoName, ok := input[0].(string)
		if !ok {
			return nil, errors.New("invalid repository name")
		}

		chartName, ok := input[1].(string)
		if !ok {
			return nil, errors.New("invalid Helm chart name")
		}

		chartVersion, ok := input[2].(string)
		if !ok {
			return nil, errors.New("invalid Helm chart version")
		}

		values, ok := input[3].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid Helm chart values")
		}

		componentName, ok := input[4].(string)
		if !ok {
			return nil, errors.New("invalid component name")
		}

		return r.Render(repoName, chartName, chartVersion, componentName, values)
	}

	nf := &jsonnet.NativeFunction{
		Name:   "renderHelmChart",
		Params: ast.Identifiers{"repository", "chart", "version", "params", "componentName"},
		Func:   fn,
	}

	return nf
}

// Render renders a Helm chart.
func (r *Renderer) Render(repoName, chartName, chartVersion, componentName string, values map[string]interface{}) ([]interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"repoName":     repoName,
		"chartName":    chartName,
		"chartVersion": chartVersion,
		"values":       values,
	}).Debug("rendering helm chart")

	if r.app == nil {
		return nil, errors.New("app object is nil")
	}

	if chartVersion == "" {
		var err error
		chartVersion, err = LatestChartVersion(r.app, repoName, chartName)
		if err != nil {
			return nil, errors.Wrapf(err, "finding latest chart release for %s/%s", repoName, chartName)
		}

		logrus.WithFields(logrus.Fields{
			"version":   chartVersion,
			"chartName": chartName,
			"repoName":  repoName}).Debug("using latest helm chart version because none was supplied")
	}

	chartPath := filepath.Join(r.app.Root(), "vendor", repoName, chartName, "helm", chartVersion, chartName)

	b, err := goyaml.Marshal(values)
	if err != nil {
		return nil, err
	}

	rendered, err := r.renderWithHelm(componentName, string(b), chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "rendering Helm chart")
	}

	var out []interface{}
	for name, s := range rendered {
		if !ksstrings.InSlice(filepath.Ext(name), []string{".yaml", ".yml"}) {
			continue
		}
		var m map[string]interface{}
		if err := goyaml.Unmarshal([]byte(s), &m); err != nil {
			return nil, errors.Wrapf(err, "unmarshalling %s", name)
		}

		out = append(out, m)
	}

	return out, nil
}

func (r *Renderer) renderWithHelm(componentName, raw, chartPath string) (map[string]string, error) {
	config := &chart.Config{Raw: raw, Values: map[string]*chart.Value{}}

	c, err := chartutil.LoadDir(chartPath)
	if err != nil {
		return nil, errors.Wrap(err, "loading Helm chart")
	}

	if req, err := chartutil.LoadRequirements(c); err == nil {
		if err := checkDependencies(c, req); err != nil {
			return nil, err
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return nil, fmt.Errorf("cannot load requirements: %v", err)
	}

	err = chartutil.ProcessRequirementsEnabled(c, config)
	if err != nil {
		return nil, err
	}

	err = chartutil.ProcessRequirementsImportValues(c)
	if err != nil {
		return nil, err
	}

	options := chartutil.ReleaseOptions{
		Name: componentName,
	}
	vals, err := chartutil.ToRenderValuesCaps(c, config, options, nil)
	if err != nil {
		return nil, err
	}

	renderer := engine.New()
	rendered, err := renderer.Render(c, vals)
	if err != nil {
		return nil, errors.Wrap(err, "rendering Helm chart")
	}

	return rendered, nil
}

func checkDependencies(ch *chart.Chart, reqs *chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata != nil && d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}
