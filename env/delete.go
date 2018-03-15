// Copyright 2018 The kubecfg authors
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

package env

import (
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	log "github.com/sirupsen/logrus"
)

// DeleteConfig is a configuration for deleting an environment.
type DeleteConfig struct {
	App  app.App
	Name string
}

// Delete deletes an environment.
func Delete(config DeleteConfig) error {
	d, err := newDeleter(config)
	if err != nil {
		return err
	}
	return d.Delete()
}

type deleter struct {
	DeleteConfig
}

func newDeleter(config DeleteConfig) (*deleter, error) {
	return &deleter{
		DeleteConfig: config,
	}, nil
}

func (d *deleter) Delete() error {
	envPath, err := filepath.Abs(filepath.Join(d.App.Root(), envRoot, d.Name))
	if err != nil {
		return err
	}

	log.Infof("Deleting environment %q with metadata at path %q", d.Name, envPath)

	// Remove the directory and all files within the environment path.
	if err = d.App.Fs().RemoveAll(envPath); err != nil {
		// if err = d.cleanEmptyParentDirs(); err != nil {
		log.Debugf("Failed to remove environment directory at path %q", envPath)
		return err
	}

	if err = d.App.RemoveEnvironment(d.Name); err != nil {
		return err
	}

	if err = cleanEmptyDirs(d.App); err != nil {
		return err
	}

	log.Infof("Successfully removed environment '%s'", d.Name)
	return nil
}
