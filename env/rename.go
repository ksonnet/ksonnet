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
	App     app.App
	AppRoot string
	Fs      afero.Fs
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

	pathFrom := envPath(r.AppRoot, from)
	pathTo := envPath(r.AppRoot, to)

	exists, err := afero.DirExists(r.Fs, pathFrom)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("environment directory for %q does not exist", from)
	}

	log.Infof("Setting environment name from %q to %q", from, to)

	current, err := Retrieve(r.App, from)
	if err != nil {
		return err
	}

	update := &app.EnvironmentSpec{
		Destination: &app.EnvironmentDestinationSpec{
			Namespace: current.Destination.Namespace(),
			Server:    current.Destination.Server(),
		},
		KubernetesVersion: current.KubernetesVersion,
		Targets:           current.Targets,
		Path:              to,
	}

	k8sSpecFlag := fmt.Sprintf("version:%s", current.KubernetesVersion)

	if err = r.App.AddEnvironment(from, k8sSpecFlag, update); err != nil {
		return err
	}

	if err = moveDir(r.Fs, pathFrom, pathTo); err != nil {
		return err
	}

	if err = cleanEmptyDirs(r.Fs, r.AppRoot); err != nil {
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

	exists, err := envExists(r.Fs, r.AppRoot, to)
	if err != nil {
		log.Debugf("Failed to check whether environment %q already exists", to)
		return err
	}
	if exists {
		return fmt.Errorf("Failed to update %q; environment %q exists", from, to)
	}

	return nil
}

func envExists(fs afero.Fs, appRoot, name string) (bool, error) {
	path := envPath(appRoot, name, envFileName)
	return afero.Exists(fs, path)
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

func envPath(root, name string, subPath ...string) string {
	return filepath.Join(append([]string{root, envRoot, name}, subPath...)...)
}

func cleanEmptyDirs(fs afero.Fs, root string) error {
	log.Debug("Removing empty environment directories, if any")
	envPath := filepath.Join(root, envRoot)
	return afero.Walk(fs, envPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !fi.IsDir() {
			return nil
		}

		isEmpty, err := afero.IsEmpty(fs, path)
		if err != nil {
			log.Debugf("Failed to check whether directory at path %q is empty", path)
			return err
		}
		if isEmpty {
			return fs.RemoveAll(path)
		}
		return nil
	})
}
