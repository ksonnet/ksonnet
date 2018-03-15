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
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/ksonnet/ksonnet/metadata/app"
)

// RenameConfig are options for renaming an environment.
type RenameConfig struct {
	App app.App
}

// Rename renames an environment
func Rename(from, to string, config RenameConfig) error {
	r, err := newRenamer(config)
	if err != nil {
		return err
	}
	return r.Rename(from, to)
}

type renamer struct {
	RenameConfig
}

func newRenamer(config RenameConfig) (*renamer, error) {
	return &renamer{
		RenameConfig: config,
	}, nil
}

func (r *renamer) Rename(from, to string) error {
	if from == to || to == "" {
		return nil
	}

	if err := r.preflight(from, to); err != nil {
		return err
	}

	log.Infof("Setting environment name from %q to %q", from, to)

	if err := r.App.RenameEnvironment(from, to); err != nil {
		return err
	}

	if err := cleanEmptyDirs(r.App); err != nil {
		return errors.Wrap(err, "clean empty directories")
	}

	log.Infof("Successfully moved %q to %q", from, to)
	return nil
}

func (r *renamer) preflight(from, to string) error {
	if !isValidName(to) {
		return fmt.Errorf("Environment name %q is not valid; must not contain punctuation, spaces, or begin or end with a slash",
			to)
	}

	exists, err := envExists(r.App, to)
	if err != nil {
		log.Debugf("Failed to check whether environment %q already exists", to)
		return err
	}
	if exists {
		return fmt.Errorf("Failed to update %q; environment %q exists", from, to)
	}

	return nil
}

func envExists(ksApp app.App, name string) (bool, error) {
	path := envPath(ksApp, name, envFileName)
	return afero.Exists(ksApp.Fs(), path)
}

func moveDir(fs afero.Fs, src, dest string) error {
	exists, err := afero.DirExists(fs, dest)
	if err != nil {
		return err
	}

	if !exists {
		if err = fs.MkdirAll(dest, app.DefaultFolderPermissions); err != nil {
			return errors.Wrapf(err, "unable to create destination %q", dest)
		}
	}

	fis, err := afero.ReadDir(fs, src)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != ".metadata" {
			continue
		}

		srcPath := filepath.Join(src, fi.Name())
		destPath := filepath.Join(dest, fi.Name())

		if err = fs.Rename(srcPath, destPath); err != nil {
			return err
		}
	}

	return nil
}

func envPath(ksApp app.App, name string, subPath ...string) string {
	return filepath.Join(append([]string{ksApp.Root(), envRoot, name}, subPath...)...)
}

func cleanEmptyDirs(ksApp app.App) error {
	log.Debug("Removing empty environment directories, if any")
	envPath := filepath.Join(ksApp.Root(), envRoot)
	return afero.Walk(ksApp.Fs(), envPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !fi.IsDir() {
			return nil
		}

		isEmpty, err := afero.IsEmpty(ksApp.Fs(), path)
		if err != nil {
			log.Debugf("Failed to check whether directory at path %q is empty", path)
			return err
		}
		if isEmpty {
			return ksApp.Fs().RemoveAll(path)
		}
		return nil
	})
}
