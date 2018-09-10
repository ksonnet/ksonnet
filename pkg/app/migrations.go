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

package app

import (
	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type migrationFn func(fromVersion semver.Version, fromSchema interface{}) (semver.Version, interface{}, error)
type loaderFn func() (interface{}, error)

// loaderFn will load a particular schema version, while migrationFn will migrate it forward one step.
type migration struct {
	migrationFn migrationFn
	loaderFn    loaderFn
}

// Migrator describes a sequence of migrations for any backward supported schemas
// to reach the currently supported schema version
type Migrator struct {
	fs         afero.Fs
	migrations map[string]migration
}

// NewMigrator constructs a migrator with the provided dependencies
func NewMigrator(fs afero.Fs, root string) Migrator {
	m := Migrator{
		fs:         fs,
		migrations: make(map[string]migration),
	}

	m.migrations["0.1.0"] = migration{
		migrationFn: newMigration020(fs, root),
		loaderFn:    newLoader010(fs, root),
	}
	m.migrations["0.2.0"] = migration{
		migrationFn: newMigration030(fs, root),
		loaderFn:    newLoader020(fs, root),
	}
	// Current version just needs a loader
	m.migrations["0.3.0"] = migration{
		loaderFn: newLoader030(fs, root),
	}
	return m
}

// Load loads a schema, performing any necessary migrations to bring it to the most current version.
func (m Migrator) Load(fromVersion semver.Version, force bool) (interface{}, error) {
	version := fromVersion

	var cur interface{}
	for {
		var err error
		v := version.String()
		mig, ok := m.migrations[v]
		if !ok {
			return nil, errors.Errorf("migration not found for schema version %s", v)
		}
		if mig.loaderFn == nil {
			return nil, errors.Errorf("nil loaderFn for migration %s", v)
		}

		if cur == nil {
			log.Debugf("loading schema version %s", v)
			cur, err = mig.loaderFn()
			if err != nil {
				return nil, errors.Wrapf(err, "loading schema version %s", v)
			}
		}

		// Final (current) version will have no migrationFn. We are done.
		if mig.migrationFn == nil {
			break
		}

		version, cur, err = mig.migrationFn(version, cur)
		if err != nil {
			return nil, errors.Wrapf(err, "migrating from schema %s", v)
		}
		log.Debugf("migrated schema to version %s", version.String())
	}

	return cur, nil
}

func newLoader010(fs afero.Fs, root string) func() (interface{}, error) {
	return func() (interface{}, error) {
		contents, err := afero.ReadFile(fs, specPath(root))
		if err != nil {
			return nil, err
		}

		var spec = &Spec010{}

		err = yaml.Unmarshal(contents, spec)
		if err != nil {
			return nil, err
		}

		return spec, nil
	}
}

func newLoader020(fs afero.Fs, root string) func() (interface{}, error) {
	return func() (interface{}, error) {
		contents, err := afero.ReadFile(fs, specPath(root))
		if err != nil {
			return nil, err
		}

		var spec = &Spec020{}

		err = yaml.Unmarshal(contents, spec)
		if err != nil {
			return nil, err
		}

		return spec, nil
	}
}

func newLoader030(fs afero.Fs, root string) func() (interface{}, error) {
	return func() (interface{}, error) {
		contents, err := afero.ReadFile(fs, specPath(root))
		if err != nil {
			return nil, err
		}

		var spec = &Spec030{}

		err = yaml.Unmarshal(contents, spec)
		if err != nil {
			return nil, err
		}

		return spec, nil
	}
}

func newMigration020(fs afero.Fs, root string) migrationFn {
	expectedVersion := semver.MustParse("0.1.0")
	newVersion := semver.MustParse("0.2.0")
	return func(fromVersion semver.Version, fromSchema interface{}) (semver.Version, interface{}, error) {
		if !fromVersion.Equals(expectedVersion) {
			return semver.Version{}, nil, errors.Errorf("unexpected version: %v", fromVersion)
		}
		if fromSchema == nil {
			return semver.Version{}, nil, errors.New("input schema is nil")
		}
		src, ok := fromSchema.(*Spec010)
		if !ok {
			return semver.Version{}, nil, errors.Errorf("type mismatch on input schema: %T", fromSchema)
		}

		dst, err := migrateSchema010To020(src)
		if err != nil {
			return semver.Version{}, nil, err
		}
		return newVersion, dst, nil
	}
}

// Converts 010 schema to 020. Does not persist the changes.
func migrateSchema010To020(src *Spec010) (*Spec020, error) {
	var dst = &Spec020{}

	// Deep copy starts here
	dst.APIVersion = "0.2.0"
	dst.Kind = src.Kind
	dst.Name = src.Name
	dst.Version = src.Version
	dst.Description = src.Description
	dst.Authors = make([]string, len(src.Authors))
	copy(dst.Authors, src.Authors)

	dst.Contributors = make(ContributorSpecs020, len(src.Contributors))
	for i, c := range src.Contributors {
		dst.Contributors[i] = &ContributorSpec020{
			Name:  c.Name,
			Email: c.Email,
		}
	}
	if src.Repository != nil {
		dst.Repository = &RepositorySpec020{
			Type: src.Repository.Type,
			URI:  src.Repository.URI,
		}
	}
	dst.Bugs = src.Bugs
	dst.Keywords = make([]string, len(src.Keywords))
	copy(dst.Keywords, src.Keywords)
	dst.Registries = RegistryConfigs020{}
	for k, v := range src.Registries {
		dst.Registries[k] = &RegistryConfig020{
			Name:       v.Name,
			Protocol:   v.Protocol,
			URI:        v.URI,
			isOverride: false,
		}
		// NOTE GitVersion was removed from RegistryConfig in 0.2.0
	}

	dst.Environments = EnvironmentConfigs020{}
	for k, v := range src.Environments {
		targets := make([]string, len(v.Targets))
		copy(targets, v.Targets)

		dst.Environments[k] = &EnvironmentConfig020{
			Name:              v.Name,
			KubernetesVersion: v.KubernetesVersion,
			Path:              v.Path,
			Targets:           targets,
			Libraries:         LibraryConfigs020{},
			isOverride:        false,
		}

		if v.Destination != nil {
			dst.Environments[k].Destination = &EnvironmentDestinationSpec020{
				Server:    v.Destination.Server,
				Namespace: v.Destination.Namespace,
			}
		}

		// 010 did not have environment-scoped libraries
	}

	dst.Libraries = LibraryConfigs020{}
	for k, v := range src.Libraries {
		l := &LibraryConfig020{
			Name:     v.Name,
			Registry: v.Registry,
		}
		if v.GitVersion != nil {
			// 010 GitVersion.CommitSHA migrates to 020 Version field
			l.Version = v.GitVersion.CommitSHA
		}
		dst.Libraries[k] = l
	}

	dst.License = src.License

	return dst, nil
}

func newMigration030(fs afero.Fs, root string) migrationFn {
	expectedVersion := semver.MustParse("0.2.0")
	newVersion := semver.MustParse("0.3.0")
	return func(fromVersion semver.Version, fromSchema interface{}) (semver.Version, interface{}, error) {
		if !fromVersion.Equals(expectedVersion) {
			return semver.Version{}, nil, errors.Errorf("unexpected version: %v", fromVersion)
		}
		if fromSchema == nil {
			return semver.Version{}, nil, errors.New("input schema is nil")
		}
		src, ok := fromSchema.(*Spec020)
		if !ok {
			return semver.Version{}, nil, errors.Errorf("type mismatch on input schema: %T", fromSchema)
		}

		dst, err := migrateSchema020To030(src)
		if err != nil {
			return semver.Version{}, nil, err
		}
		return newVersion, dst, nil
	}
}

// Converts 020 schema to 030. Does not persist the changes.
func migrateSchema020To030(src *Spec020) (*Spec030, error) {
	var dst = &Spec030{}

	// Deep copy starts here
	dst.APIVersion = "0.3.0"
	dst.Kind = src.Kind
	dst.Name = src.Name
	dst.Version = src.Version
	dst.Description = src.Description
	dst.Authors = make([]string, len(src.Authors))
	copy(dst.Authors, src.Authors)

	dst.Contributors = make(ContributorSpecs030, len(src.Contributors))
	for i, c := range src.Contributors {
		dst.Contributors[i] = &ContributorSpec030{
			Name:  c.Name,
			Email: c.Email,
		}
	}
	if src.Repository != nil {
		dst.Repository = &RepositorySpec030{
			Type: src.Repository.Type,
			URI:  src.Repository.URI,
		}
	}
	dst.Bugs = src.Bugs
	dst.Keywords = make([]string, len(src.Keywords))
	copy(dst.Keywords, src.Keywords)
	dst.Registries = RegistryConfigs030{}
	for k, v := range src.Registries {
		dst.Registries[k] = &RegistryConfig030{
			Name:     v.Name,
			Protocol: v.Protocol,
			URI:      v.URI,
		}
	}

	dst.Environments = EnvironmentConfigs030{}
	for k, v := range src.Environments {
		libs := LibraryConfigs030{}
		targets := make([]string, len(v.Targets))
		copy(targets, v.Targets)

		dst.Environments[k] = &EnvironmentConfig030{
			Name:              v.Name,
			KubernetesVersion: v.KubernetesVersion,
			Path:              v.Path,
			Targets:           targets,
			Libraries:         libs,
		}

		if v.Destination != nil {
			dst.Environments[k].Destination = &EnvironmentDestinationSpec030{
				Server:    v.Destination.Server,
				Namespace: v.Destination.Namespace,
			}
		}

		for lk, lv := range v.Libraries {
			name := libName(lk)
			qualifiedName := qualifyLibName(lv.Registry, name)

			l := &LibraryConfig030{
				Name:     name,
				Registry: lv.Registry,
				Version:  lv.Version,
			}
			libs[qualifiedName] = l
		}
	}

	dst.Libraries = LibraryConfigs030{}
	for k, v := range src.Libraries {
		name := libName(k)
		qualifiedName := qualifyLibName(v.Registry, name)

		l := &LibraryConfig030{
			Name:     name,
			Registry: v.Registry,
			Version:  v.Version,
		}
		dst.Libraries[qualifiedName] = l
	}

	dst.License = src.License

	return dst, nil
}
