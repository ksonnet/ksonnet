package env

import (
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/app"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// DeleteConfig is a configuration for deleting an environment.
type DeleteConfig struct {
	App     app.App
	AppRoot string
	Name    string
	Fs      afero.Fs
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
	envPath, err := filepath.Abs(filepath.Join(d.AppRoot, envRoot, d.Name))
	if err != nil {
		return err
	}

	log.Infof("Deleting environment %q with metadata at path %q", d.Name, envPath)

	// Remove the directory and all files within the environment path.
	if err = d.Fs.RemoveAll(envPath); err != nil {
		// if err = d.cleanEmptyParentDirs(); err != nil {
		log.Debugf("Failed to remove environment directory at path %q", envPath)
		return err
	}

	if err = d.App.RemoveEnvironment(d.Name); err != nil {
		return err
	}

	if err = cleanEmptyDirs(d.Fs, d.AppRoot); err != nil {
		return err
	}

	log.Infof("Successfully removed environment '%s'", d.Name)
	return nil
}
