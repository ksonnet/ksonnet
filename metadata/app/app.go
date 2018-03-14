package app

import (
	"os"
	"path/filepath"
	"sort"

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
	Fs() afero.Fs
	Init() error
	LibPath(envName string) (string, error)
	Libraries() LibraryRefSpecs
	Registries() RegistryRefSpecs
	RemoveEnvironment(name string) error
	RenameEnvironment(from, to string) error
	Root() string
	UpdateTargets(envName string, targets []string) error
	Upgrade(dryRun bool) error
}

type baseApp struct {
	root string
	fs   afero.Fs
}

func (ba *baseApp) Fs() afero.Fs {
	return ba.fs
}

func (ba *baseApp) Root() string {
	return ba.root
}

func newBaseApp(fs afero.Fs, root string) *baseApp {
	return &baseApp{
		fs:   fs,
		root: root,
	}
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

func updateLibData(fs afero.Fs, k8sSpecFlag, libPath string, useVersionPath bool) (string, error) {
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
func StubUpdateLibData(fs afero.Fs, k8sSpecFlag, libPath string, useVersionPath bool) (string, error) {
	return "v1.8.7", nil
}

func moveEnvironment(fs afero.Fs, root, from, to string) error {
	toPath := filepath.Join(root, EnvironmentDirName, to)

	exists, err := afero.Exists(fs, filepath.Join(toPath, "main.jsonnet"))
	if err != nil {
		return err
	}

	if exists {
		return errors.Errorf("unable to rename %q because %q exists", from, to)
	}

	fromPath := filepath.Join(root, EnvironmentDirName, from)
	exists, err = afero.Exists(fs, fromPath)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("environment %q does not exist", from)
	}

	// create to directory
	if err = fs.MkdirAll(toPath, DefaultFolderPermissions); err != nil {
		return err
	}

	fis, err := afero.ReadDir(fs, fromPath)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != ".metadata" {
			continue
		}

		oldPath := filepath.Join(fromPath, fi.Name())
		newPath := filepath.Join(toPath, fi.Name())
		if err := fs.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return cleanEnv(fs, root)
}

func cleanEnv(fs afero.Fs, root string) error {
	var dirs []string

	envDir := filepath.Join(root, EnvironmentDirName)
	err := afero.Walk(fs, envDir, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			return nil
		}

		dirs = append(dirs, path)
		return nil
	})

	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	for _, dir := range dirs {
		fis, err := afero.ReadDir(fs, dir)
		if err != nil {
			return err
		}

		if len(fis) == 0 {
			if err := fs.RemoveAll(dir); err != nil {
				return err
			}
		}
	}

	return nil
}
