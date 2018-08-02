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

package cluster

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghodss/yaml"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ShowConfig is configuration for Show.
type ShowConfig struct {
	App            app.App
	ComponentNames []string
	EnvName        string
	Format         string
	Out            io.Writer
}

// ShowOpts is an option for configuring Show.
type ShowOpts func(*Show)

// Show shows objects.
type Show struct {
	ShowConfig

	// these make it easier to test Show.
	findObjectsFn findObjectsFn
}

// RunShow shows objects for a given configuration.
func RunShow(config ShowConfig, opts ...ShowOpts) error {
	s := &Show{
		ShowConfig:    config,
		findObjectsFn: findObjects,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s.Show()
}

// Show shows objects.
func (s *Show) Show() error {
	apiObjects, err := s.findObjectsFn(s.App, s.EnvName, s.ComponentNames)
	if err != nil {
		return errors.Wrap(err, "find objects")
	}

	sorted := make([]*unstructured.Unstructured, len(apiObjects))
	copy(sorted, apiObjects)
	UnstructuredSlice(sorted).Sort()

	switch s.Format {
	case "yaml":
		return s.showYAML(sorted)
	case "json":
		return s.showJSON(sorted)
	default:
		return fmt.Errorf("Unknown --format: %s", s.Format)
	}
}

func (s *Show) showYAML(apiObjects []*unstructured.Unstructured) error {
	return ShowYAML(s.Out, apiObjects)
}

func (s *Show) showJSON(apiObjects []*unstructured.Unstructured) error {
	enc := json.NewEncoder(s.Out)
	enc.SetIndent("", "  ")

	m := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "List",
	}

	items := make([]interface{}, 0)

	for _, obj := range apiObjects {
		items = append(items, obj.Object)
	}

	m["items"] = items

	return enc.Encode(m)
}

// ShowYAML shows YAML objects.
func ShowYAML(out io.Writer, apiObjects []*unstructured.Unstructured) error {
	for _, obj := range apiObjects {
		fmt.Fprintln(out, "---")
		buf, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		_, err = out.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}
