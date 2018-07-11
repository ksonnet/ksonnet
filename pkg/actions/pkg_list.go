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
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/pkg/errors"
)

const (
	// pkgInstalled denotes a package is installed
	pkgInstalled = "*"
)

// RunPkgList runs `pkg list`
func RunPkgList(m map[string]interface{}) error {
	rl, err := NewPkgList(m)
	if err != nil {
		return err
	}

	return rl.Run()
}

// PkgList lists available registries
type PkgList struct {
	app           app.App
	pm            registry.PackageManager
	onlyInstalled bool
	outputType    string

	registryListFn func(ksApp app.App) ([]registry.Registry, error)
	out            io.Writer
}

// NewPkgList creates an instance of PkgList
func NewPkgList(m map[string]interface{}) (*PkgList, error) {
	ol := newOptionLoader(m)

	app := ol.LoadApp()
	rl := &PkgList{
		app:           app,
		pm:            registry.NewPackageManager(app),
		onlyInstalled: ol.LoadBool(OptionInstalled),
		outputType:    ol.LoadOptionalString(OptionOutput),

		registryListFn: registry.List,
		out:            os.Stdout,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return rl, nil
}

// Run runs the env list action.
func (pl *PkgList) Run() error {
	pkgs, err := pl.pm.Packages()
	if err != nil {
		return err
	}

	index := make(map[string]pkg.Package)
	for _, p := range pkgs {
		index[p.String()] = p
	}

	if !pl.onlyInstalled {
		// Merge in remote packages
		remote, err := pl.pm.RemotePackages()
		if err != nil {
			return errors.Wrap(err, "listing remote packages")
		}

		for _, p := range remote {
			if _, ok := index[p.String()]; ok {
				continue
			}

			index[p.String()] = p
		}
	}

	// Build output
	var rows [][]string
	for _, p := range index {
		isInstalled, err := p.IsInstalled()
		if err != nil {
			return err
		}
		rows = append(rows, pl.addRow(p.RegistryName(), p.Name(), p.Version(), isInstalled))
	}

	sort.Slice(rows, func(i, j int) bool {
		nameI := strings.Join([]string{rows[i][0], rows[i][1], rows[i][2]}, "-")
		nameJ := strings.Join([]string{rows[j][0], rows[j][1], rows[j][2]}, "-")

		return nameI < nameJ
	})

	t := table.New("pkgList", pl.out)

	f, err := table.DetectFormat(pl.outputType)
	if err != nil {
		return errors.Wrap(err, "detecting output format")
	}
	t.SetFormat(f)

	t.SetHeader([]string{"registry", "name", "version", "installed"})
	t.AppendBulk(rows)
	return t.Render()
}

func (pl *PkgList) addRow(regName, libName, version string, isInstalled bool) []string {
	row := []string{regName, libName, version}
	installedText := ""
	if isInstalled {
		installedText = pkgInstalled
	}

	row = append(row, installedText)

	return row
}
