package app

import (
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/metadata/lib"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (

	// appYamlName is the name for the app specification.
	appYamlName = "app.yaml"

	// EnvironmentDirName is the directory name for environments.
	EnvironmentDirName = "environments"

	// LibDirName is the directory name for libraries.
	LibDirName = "lib"
)

var (
	// DefaultFilePermissions are the default permissions for a file.
	DefaultFilePermissions = os.FileMode(0644)
	// DefaultFolderPermissions are the default permissions for a folder.
	DefaultFolderPermissions = os.FileMode(0755)

	// LibUpdater updates ksonnet lib versions.
	LibUpdater = updateLibData
)

// App is a ksonnet application.
type App interface {
	AddEnvironment(name, k8sSpecFlag string, spec *EnvironmentSpec) error
	Environment(name string) (*EnvironmentSpec, error)
	Environments() (EnvironmentSpecs, error)
	Libraries() LibraryRefSpecs
	LibPath(envName string) (string, error)
	Init() error
	Registries() RegistryRefSpecs
	RemoveEnvironment(name string) error
	Upgrade(dryRun bool) error
}

// Load loads the application configuration.
func Load(fs afero.Fs, appRoot string) (App, error) {
	spec, err := Read(fs, appRoot)
	if err != nil {
		return nil, err
	}

	switch spec.APIVersion {
	default:
		return nil, errors.Errorf("unknown apiVersion %q in %s", spec.APIVersion, appYamlName)
	case "0.0.1":
		return NewApp001(fs, appRoot)
	case "0.1.0":
		return NewApp010(fs, appRoot)
	}
}

func updateLibData(fs afero.Fs, k8sSpecFlag string, libPath string, useVersionPath bool) (string, error) {
	lm, err := lib.NewManager(k8sSpecFlag, fs, libPath)
	if err != nil {
		return "", err
	}

	if lm.GenerateLibData(useVersionPath); err != nil {
		return "", err
	}

	return lm.K8sVersion, nil
}

func app010LibPath(root string) string {
	return filepath.Join(root, LibDirName)
}

// StubUpdateLibData always returns no error.
func StubUpdateLibData(fs afero.Fs, k8sSpecFlag string, libPath string, useVersionPath bool) (string, error) {
	return "v1.8.7", nil
}
