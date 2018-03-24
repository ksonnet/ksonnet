// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Upless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package actions

import (
	"io"
	"os"
	"sort"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/ksonnet/ksonnet/prototype"
)

// RunPrototypeList runs `prototype list`
func RunPrototypeList(ksApp app.App) error {
	pl, err := NewPrototypeList(ksApp)
	if err != nil {
		return err
	}

	return pl.Run()
}

// PrototypeList lists available namespaces
type PrototypeList struct {
	app        app.App
	out        io.Writer
	prototypes func(app.App, pkg.Descriptor) (prototype.SpecificationSchemas, error)
}

// NewPrototypeList creates an instance of PrototypeList
func NewPrototypeList(ksApp app.App) (*PrototypeList, error) {
	pl := &PrototypeList{
		app:        ksApp,
		out:        os.Stdout,
		prototypes: pkg.LoadPrototypes,
	}

	return pl, nil
}

// Run runs the env list action.
func (pl *PrototypeList) Run() error {
	libraries, err := pl.app.Libraries()
	if err != nil {
		return err
	}

	var prototypes prototype.SpecificationSchemas

	for _, library := range libraries {
		d := pkg.Descriptor{
			Registry: library.Registry,
			Part:     library.Name,
		}

		p, err := pl.prototypes(pl.app, d)
		if err != nil {
			return err
		}

		prototypes = append(prototypes, p...)
	}

	index := prototype.NewIndex(prototypes)
	prototypes, err = index.List()
	if err != nil {
		return nil
	}

	var rows [][]string
	for _, p := range prototypes {
		rows = append(rows, []string{p.Name, p.Template.ShortDescription})
	}

	t := table.New(pl.out)
	t.SetHeader([]string{"name", "description"})

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	t.AppendBulk(rows)

	return t.Render()
}
