package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/lib"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// App010 is a ksonnet 0.1.0 application.
type App010 struct {
	spec *Spec
	*baseApp
}

var _ App = (*App010)(nil)

// NewApp010 creates an App010 instance.
func NewApp010(fs afero.Fs, root string) (*App010, error) {
	ba := newBaseApp(fs, root)

	a := &App010{
		baseApp: ba,
	}

	if err := a.load(); err != nil {
		return nil, err
	}

	return a, nil
}

// AddEnvironment adds an environment spec to the app spec. If the spec already exists,
// it is overwritten.
func (a *App010) AddEnvironment(name, k8sSpecFlag string, spec *EnvironmentSpec) error {
	if err := a.load(); err != nil {
		return err
	}

	a.spec.Environments[name] = spec

	if k8sSpecFlag != "" {
		ver, err := LibUpdater(a.fs, k8sSpecFlag, app010LibPath(a.root), true)
		if err != nil {
			return err
		}

		a.spec.Environments[name].KubernetesVersion = ver
	}

	return a.save()
}

// Environment returns the spec for an environment.
func (a *App010) Environment(name string) (*EnvironmentSpec, error) {
	s, ok := a.spec.Environments[name]
	if !ok {
		return nil, errors.Errorf("environment %q was not found", name)
	}

	return s, nil
}

// Environments returns all environment specs.
func (a *App010) Environments() (EnvironmentSpecs, error) {
	return a.spec.Environments, nil
}

// Init initializes the App.
func (a *App010) Init() error {
	// check to see if there are spec.json files.

	legacyEnvs, err := a.findLegacySpec()
	if err != nil {
		return err
	}

	if len(legacyEnvs) == 0 {
		return nil
	}

	msg := "Your application's apiVersion is 0.1.0, but legacy environment declarations " +
		"where found in environments: %s. In order to proceed, you will have to run `ks upgrade` to " +
		"upgrade your application. <see url>"

	return errors.Errorf(msg, strings.Join(legacyEnvs, ", "))
}

// LibPath returns the lib path for an env environment.
func (a *App010) LibPath(envName string) (string, error) {
	env, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	ver := fmt.Sprintf("version:%s", env.KubernetesVersion)
	lm, err := lib.NewManager(ver, a.fs, app010LibPath(a.root))
	if err != nil {
		return "", err
	}

	return lm.GetLibPath(true)
}

// Libraries returns application libraries.
func (a *App010) Libraries() LibraryRefSpecs {
	return a.spec.Libraries
}

// Registries returns application registries.
func (a *App010) Registries() RegistryRefSpecs {
	return a.spec.Registries
}

// RemoveEnvironment removes an environment.
func (a *App010) RemoveEnvironment(envName string) error {
	if err := a.load(); err != nil {
		return err
	}
	delete(a.spec.Environments, envName)
	return a.save()
}

// RenameEnvironment renames environments.
func (a *App010) RenameEnvironment(from, to string) error {
	if err := moveEnvironment(a.fs, a.root, from, to); err != nil {
		return err
	}

	if err := a.load(); err != nil {
		return err
	}

	a.spec.Environments[to] = a.spec.Environments[from]
	delete(a.spec.Environments, from)

	a.spec.Environments[to].Path = to

	return a.save()
}

// UpdateTargets updates the list of targets for a 0.1.0 application.
func (a *App010) UpdateTargets(envName string, targets []string) error {
	spec, err := a.Environment(envName)
	if err != nil {
		return err
	}

	spec.Targets = targets

	return errors.Wrap(a.AddEnvironment(envName, "", spec), "update targets")
}

// Upgrade upgrades the app to the latest apiVersion.
func (a *App010) Upgrade(dryRun bool) error {
	return nil
}

func (a *App010) findLegacySpec() ([]string, error) {
	var found []string

	envPath := filepath.Join(a.root, EnvironmentDirName)
	err := afero.Walk(a.fs, envPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		if fi.Name() == app001specJSON {
			envName := strings.TrimPrefix(path, envPath+"/")
			envName = strings.TrimSuffix(envName, "/"+app001specJSON)
			found = append(found, envName)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return found, nil
}

func (a *App010) save() error {
	return Write(a.fs, a.root, a.spec)
}

func (a *App010) load() error {
	spec, err := Read(a.fs, a.root)
	if err != nil {
		return err
	}

	a.spec = spec
	return nil
}
