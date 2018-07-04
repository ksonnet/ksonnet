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
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/pkg/errors"
)

// RunPrototypeSearch runs `prototype search`
func RunPrototypeSearch(m map[string]interface{}) error {
	ps, err := NewPrototypeSearch(m)
	if err != nil {
		return err
	}

	return ps.Run()
}

// PrototypeSearch searches for prototypes by name.
type PrototypeSearch struct {
	app        app.App
	query      string
	outputType string

	out            io.Writer
	packageManager registry.PackageManager
	protoSearchFn  func(string, prototype.Prototypes) (prototype.Prototypes, error)
}

// NewPrototypeSearch creates an instance of PrototypeSearch
func NewPrototypeSearch(m map[string]interface{}) (*PrototypeSearch, error) {
	ol := newOptionLoader(m)

	app := ol.LoadApp()
	ps := &PrototypeSearch{
		app:        app,
		query:      ol.LoadString(OptionQuery),
		outputType: ol.LoadOptionalString(OptionOutput),

		out:            os.Stdout,
		packageManager: registry.NewPackageManager(app),
		protoSearchFn:  protoSearch,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return ps, nil
}

// Run runs the env list action.
func (ps *PrototypeSearch) Run() error {
	prototypes, err := ps.packageManager.Prototypes()
	if err != nil {
		return err
	}

	results, err := ps.protoSearchFn(ps.query, prototypes)
	if err != nil {
		return err
	}

	if len(results) == 1 {
		return fmt.Errorf("failed to find any search results for query %q", ps.query)
	}

	var rows [][]string
	for _, p := range results {
		rows = append(rows, []string{p.Name, p.Template.ShortDescription})
	}

	t := table.New("prototypeSearch", ps.out)
	t.SetHeader([]string{"name", "description"})

	f, err := table.DetectFormat(ps.outputType)
	if err != nil {
		return errors.Wrap(err, "detecting output format")
	}
	t.SetFormat(f)

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	t.AppendBulk(rows)

	return t.Render()
}

func protoSearch(query string, prototypes prototype.Prototypes) (prototype.Prototypes, error) {
	index, err := prototype.NewIndex(prototypes, prototype.DefaultBuilder)
	if err != nil {
		return nil, err
	}
	return index.SearchNames(query, prototype.Substring)
}
