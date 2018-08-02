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

package diff

import (
	"bytes"
	"io"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/cluster"
	"github.com/ksonnet/ksonnet/pkg/pipeline"
	"github.com/pkg/errors"
	godiff "github.com/shazow/go-diff"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Differ generates the differences between two Locations.
type Differ struct {
	App        app.App
	Config     *client.Config
	Components []string

	localGen  yamlGenerator
	remoteGen yamlGenerator
}

// DefaultDiff runs diff with default options.
func DefaultDiff(a app.App, config *client.Config, components []string, l1 *Location, l2 *Location) (io.Reader, error) {
	differ := New(a, config, components)
	return differ.Diff(l2, l1)
}

// New creates an instance of Differ.
func New(a app.App, config *client.Config, components []string) *Differ {
	yl := newYamlLocal(a)
	yr := newYamlRemote(a, config)

	d := &Differ{
		App:        a,
		Config:     config,
		Components: components,
		localGen:   yl,
		remoteGen:  yr,
	}

	return d
}

// Diff generates the differences between two locations.
func (d *Differ) Diff(location1, location2 *Location) (io.Reader, error) {
	logrus.WithFields(logrus.Fields{
		"src1": location1.String(),
		"src2": location2.String(),
	}).Debug("generating diff")

	r1, err := d.toYAML(location1)
	if err != nil {
		return nil, err
	}

	r2, err := d.toYAML(location2)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := godiff.DefaultDiffer().Diff(&buf, r1, r2); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (d *Differ) toYAML(location *Location) (io.ReadSeeker, error) {
	if err := location.Err(); err != nil {
		return nil, err
	}

	switch location.Destination() {
	default:
		return nil, errors.Errorf("unknown destation %q", location.Destination())
	case "local":
		return d.localGen.Generate(location, d.Components)
	case "remote":
		return d.remoteGen.Generate(location, d.Components)
	}
}

type yamlGenerator interface {
	Generate(*Location, []string) (io.ReadSeeker, error)
}

type yamlLocal struct {
	app              app.App
	collectObjectsFn func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error)
	showFn           func(io.Writer, []*unstructured.Unstructured) error
}

func newYamlLocal(a app.App) *yamlLocal {
	return &yamlLocal{
		app:              a,
		collectObjectsFn: localCollectObjects,
		showFn:           cluster.ShowYAML,
	}
}

func localCollectObjects(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
	p := pipeline.New(a, envName)
	return p.Objects(componentNames)
}

func (yl *yamlLocal) Generate(location *Location, components []string) (io.ReadSeeker, error) {
	var buf bytes.Buffer

	objects, err := yl.collectObjectsFn(yl.app, location.EnvName(), components)
	if err != nil {
		return nil, err

	}

	cluster.UnstructuredSlice(objects).Sort()

	if err := yl.showFn(&buf, objects); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

type yamlRemote struct {
	app              app.App
	config           *client.Config
	genClientsFn     func(a app.App, clientConfig *client.Config, envName string) (cluster.Clients, error)
	collectObjectsFn func(string, cluster.Clients, []string) ([]*unstructured.Unstructured, error)
	showFn           func(io.Writer, []*unstructured.Unstructured) error
}

func newYamlRemote(a app.App, config *client.Config) *yamlRemote {
	return &yamlRemote{
		app:              a,
		config:           config,
		genClientsFn:     cluster.GenClients,
		collectObjectsFn: cluster.CollectObjects,
		showFn:           cluster.ShowYAML,
	}
}

func (yr *yamlRemote) Generate(location *Location, components []string) (io.ReadSeeker, error) {
	var buf bytes.Buffer

	environment, err := yr.app.Environment(location.EnvName())
	if err != nil {
		return nil, err
	}

	// Create an environment-scoped set of cluster clients
	clients, err := yr.genClientsFn(yr.app, yr.config, location.EnvName())
	if err != nil {
		return nil, errors.Wrapf(err, "creating client for environment: %s", location.EnvName())
	}

	objects, err := yr.collectObjectsFn(environment.Destination.Namespace, clients, components)
	if err != nil {
		return nil, err
	}

	cluster.UnstructuredSlice(objects).Sort()

	if err := yr.showFn(&buf, objects); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}
