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
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// overrideKind is the override resource type.
	overrideKind = "ksonnet.io/app-override"
	// overrideVersion is the version of the override resource.
	overrideVersion = "0.3.0"
)

// Override defines overrides to ksonnet project configurations.
type Override = Override030

// overridePath constructs a path for app.override.yaml
func overridePath(appRoot string) string {
	return filepath.Join(appRoot, overrideYamlName)
}

func newOverride() *Override {
	o := &Override{
		APIVersion:   overrideVersion,
		Kind:         overrideKind,
		Environments: EnvironmentConfigs{},
		Registries:   RegistryConfigs{},
	}
	return o
}

// readOverrides returns optional override configuration
// Returns nil if no overrides were defined.
func readOverrides(fs afero.Fs, root string) (*Override, error) {
	log.Debugf("loading overrides from %s", root)

	exists, err := afero.Exists(fs, overridePath(root))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	var o Override

	overrideConfig, err := afero.ReadFile(fs, overridePath(root))
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(overrideConfig, &o)
	if err != nil {
		return nil, err
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	return &o, nil
}

// saveOverride saves the override to the filesystem.
func saveOverride(encoder Encoder, fs afero.Fs, root string, o *Override) error {
	if o == nil {
		return errors.New("override was nil")
	}

	o.APIVersion = overrideVersion
	o.Kind = overrideKind

	f, err := fs.OpenFile(overridePath(root), os.O_WRONLY|os.O_CREATE, DefaultFilePermissions)
	if err != nil {
		return err
	}

	if err := encoder.Encode(o, f); err != nil {
		return errors.Wrap(err, "encoding override")
	}

	return nil
}

func removeOverride(fs afero.Fs, appRoot string) error {
	exists, err := afero.Exists(fs, overridePath(appRoot))
	if err != nil {
		return err
	}

	if exists {
		return fs.Remove(overridePath(appRoot))
	}

	return nil
}
