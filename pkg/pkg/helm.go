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

package pkg

import (
	"fmt"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	ksstrings "github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

type chartConfig struct {
	Description string `json:"description"`
}

// Helm is a package based on a Helm chart.
type Helm struct {
	pkg
	config chartConfig
}

var _ Package = (*Helm)(nil)

// NewHelm creates an instance of Helm.
func NewHelm(a app.App, name, registryName, version string, installChecker InstallChecker) (*Helm, error) {
	if installChecker == nil {
		installChecker = &DefaultInstallChecker{App: a}
	}

	cp, err := chartConfigPath(a, name, registryName, version)
	if err != nil {
		return nil, errors.Wrap(err, "finding chart path")
	}

	b, err := afero.ReadFile(a.Fs(), cp)
	if err != nil {
		return nil, errors.Wrap(err, "reading chart configuration")
	}

	var cc chartConfig
	if err = yaml.Unmarshal(b, &cc); err != nil {
		return nil, errors.Wrap(err, "unmarshalling chart configuration")
	}

	return &Helm{
		pkg: pkg{
			registryName: registryName,
			name:         name,
			version:      version,

			a:              a,
			installChecker: installChecker,
		},
		config: cc,
	}, nil
}

// chartConfigPath returns directory containing vendored chart manifest (Chart.yaml)
func chartConfigDir(a app.App, name, registryName, version string) (string, error) {
	var err error
	if version == "" {
		version, err = latestChartRelease(a, name, registryName)
		if err != nil {
			return "", errors.Wrapf(err, "finding latest %s chart release", name)
		}
	}

	// Construct path: `vendor/<registry>/<pkg>/helm/<version>/<pkg>`
	chartConfigPath := filepath.Join(a.VendorPath(), registryName, name, "helm", version, name)
	return chartConfigPath, nil
}

// chartConfigPath returns path to vendored chart manifest (Chart.yaml)
func chartConfigPath(a app.App, name, registryName, version string) (string, error) {
	dir, err := chartConfigDir(a, name, registryName, version)
	if err != nil {
		return "", err
	}

	chartConfigPath := filepath.Join(dir, "Chart.yaml")
	return chartConfigPath, nil
}

func latestChartRelease(a app.App, name, registryName string) (string, error) {
	chartPath := filepath.Join(a.VendorPath(), registryName, name, "helm")

	fis, err := afero.ReadDir(a.Fs(), chartPath)
	if err != nil {
		return "", err
	}

	var versions []semver.Version
	for _, fi := range fis {
		if fi.IsDir() {
			v, err := semver.Make(fi.Name())
			if err != nil {
				return "", err
			}

			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return "", errors.Errorf("chart %q doesn't have any releases", name)
	}

	semver.Sort(versions)
	return versions[0].String(), nil
}

// Description returns the description for the Helm chart. The description
// is retrieved from the chart's Chart.yaml file.
func (h *Helm) Description() string {
	return h.config.Description
}

func (h *Helm) prototypeName() string {
	return fmt.Sprintf("io.ksonnet.pkg.%s-%s", h.registryName, h.name)
}

// Prototypes returns prototypes for this package. Currently, it returns a single prototype.
func (h *Helm) Prototypes() (prototype.Prototypes, error) {

	shortDescription := fmt.Sprintf("Helm Chart %s from the %s registry",
		h.name, h.registryName)

	p := &prototype.Prototype{
		APIVersion: prototype.DefaultAPIVersion,
		Kind:       prototype.DefaultKind,
		Name:       h.prototypeName(),
		Template: prototype.SnippetSchema{
			Description:      shortDescription,
			ShortDescription: shortDescription,
			JsonnetBody:      []string{"{}"},
		},
		Params: prototype.ParamSchemas{
			{
				Name:        "name",
				Description: "Name of the component",
				Type:        prototype.String,
			},
			{
				Name:        "version",
				Description: "Version of the Helm chart",
				// TODO: find the latest version of the helm chart
				Default: ksstrings.Ptr("1.2.3"),
				Type:    prototype.String,
			},
		},
	}

	return prototype.Prototypes{p}, nil
}

// Path returns local directory for vendoring the package.
func (h *Helm) Path() string {
	if h == nil {
		return ""
	}
	if h.a == nil {
		return ""
	}

	path, err := chartConfigDir(h.a, h.name, h.registryName, h.version)
	if err != nil {
		return ""
	}
	return path
}
