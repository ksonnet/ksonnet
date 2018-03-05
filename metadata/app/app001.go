package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// app001specJSON is the name for environment schema
	app001specJSON = "spec.json"
)

// App001 is a ksonnet 0.0.1 application.
type App001 struct {
	spec *Spec
	root string
	fs   afero.Fs
	out  io.Writer
}

var _ App = (*App001)(nil)

// NewApp001 creates an App001 instance.
func NewApp001(fs afero.Fs, root string) (*App001, error) {
	spec, err := Read(fs, root)
	if err != nil {
		return nil, err
	}

	return &App001{
		spec: spec,
		fs:   fs,
		root: root,
		out:  os.Stdout,
	}, nil
}

// Init initializes the App.
func (a *App001) Init() error {
	msg := "Your application's apiVersion is below 0.1.0. In order to use all ks features, you " +
		"can upgrade your application using `ks upgrade`."
	log.Warn(msg)

	return nil
}

// AddEnvironment adds an environment spec to the app spec. If the spec already exists,
// it is overwritten.
func (a *App001) AddEnvironment(name, k8sSpecFlag string, spec *EnvironmentSpec) error {
	envPath := filepath.Join(a.root, EnvironmentDirName, name)
	if err := a.fs.MkdirAll(envPath, DefaultFolderPermissions); err != nil {
		return err
	}

	specPath := filepath.Join(envPath, app001specJSON)

	b, err := json.Marshal(spec.Destination)
	if err != nil {
		return err
	}

	if err = afero.WriteFile(a.fs, specPath, b, DefaultFilePermissions); err != nil {
		return err
	}

	_, err = LibUpdater(a.fs, k8sSpecFlag, a.appLibPath(name), false)
	return err
}

// Registries returns application registries.
func (a *App001) Registries() RegistryRefSpecs {
	return a.spec.Registries
}

// Libraries returns application libraries.
func (a *App001) Libraries() LibraryRefSpecs {
	return a.spec.Libraries
}

// Environment returns the spec for an environment. In 0.1.0, the file lives in
// /environments/name/spec.json.
func (a *App001) Environment(name string) (*EnvironmentSpec, error) {
	path := filepath.Join(a.root, EnvironmentDirName, name, app001specJSON)
	return read001EnvSpec(a.fs, name, path)
}

// Environments returns specs for all environments. In 0.1.0, the environment spec
// lives in spec.json files.
func (a *App001) Environments() (EnvironmentSpecs, error) {
	specs := EnvironmentSpecs{}

	root := filepath.Join(a.root, EnvironmentDirName)

	err := afero.Walk(a.fs, root, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		if fi.Name() == app001specJSON {
			dir := filepath.Dir(path)
			envName := strings.TrimPrefix(dir, root+"/")
			spec, err := read001EnvSpec(a.fs, envName, path)
			if err != nil {
				return err
			}

			specs[envName] = spec
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return specs, nil
}

type k8sSchema struct {
	Info struct {
		Version string `json:"version,omitempty"`
	} `json:"info,omitempty"`
}

func read001EnvSpec(fs afero.Fs, name, path string) (*EnvironmentSpec, error) {
	b, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}

	var s EnvironmentDestinationSpec
	if err = json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	if s.Namespace == "" {
		s.Namespace = "default"
	}

	envPath := filepath.Dir(path)
	swaggerPath := filepath.Join(envPath, ".metadata", "swagger.json")

	b, err = afero.ReadFile(fs, swaggerPath)
	if err != nil {
		return nil, err
	}

	var ks k8sSchema
	if err = json.Unmarshal(b, &ks); err != nil {
		return nil, err
	}

	if ks.Info.Version == "" {
		return nil, errors.New("unable to determine environment Kubernetes version")
	}

	spec := EnvironmentSpec{
		Path:              name,
		Destination:       &s,
		KubernetesVersion: ks.Info.Version,
	}

	return &spec, nil
}

// RemoveEnvironment removes an environment.
func (a *App001) RemoveEnvironment(envName string) error {
	return nil
}

// Upgrade upgrades the app to the latest apiVersion.
func (a *App001) Upgrade(dryRun bool) error {
	if err := a.load(); err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(a.out, "\n[dry run] Upgrading application settings from version 0.0.1 to to 0.1.0.\n")
	}

	envs, err := a.Environments()
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(a.out, "[dry run] Converting 0.0.1 environments to 0.1.0a:\n")
	}
	for _, env := range envs {
		a.convertEnvironment(env.Path, dryRun)
	}

	a.spec.APIVersion = "0.1.0"

	if dryRun {
		data, err := a.spec.Marshal()
		if err != nil {
			return err
		}

		fmt.Fprintf(a.out, "\n[dry run] Upgraded app.yaml:\n%s\n", string(data))
		fmt.Fprintf(a.out, "[dry run] You can preform the migration by running `ks upgrade`.\n")
		return nil
	}

	return a.save()
}

func (a *App001) convertEnvironment(envName string, dryRun bool) error {
	path := filepath.Join(a.root, EnvironmentDirName, envName, "spec.json")
	env, err := read001EnvSpec(a.fs, envName, path)
	if err != nil {
		return err
	}

	a.spec.Environments[envName] = env

	if dryRun {
		fmt.Fprintf(a.out, "[dry run]\t* adding the environment description in environment `%s to `app.yaml`.\n",
			envName)
		return nil
	}

	if err = a.fs.Remove(path); err != nil {
		return err
	}

	k8sSpecFlag := fmt.Sprintf("version:%s", env.KubernetesVersion)
	_, err = LibUpdater(a.fs, k8sSpecFlag, app010LibPath(a.root), true)
	return err
}

func (a *App001) appLibPath(envName string) string {
	return filepath.Join(a.root, EnvironmentDirName, envName, ".metadata")
}

func (a *App001) save() error {
	return Write(a.fs, a.root, a.spec)
}

func (a *App001) load() error {
	spec, err := Read(a.fs, a.root)
	if err != nil {
		return err
	}

	a.spec = spec
	return nil
}

// LibPath returns the lib path for an env environment.
func (a *App001) LibPath(envName string) (string, error) {
	env, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	ver := fmt.Sprintf("version:%s", env.KubernetesVersion)
	lm, err := lib.NewManager(ver, a.fs, a.appLibPath(envName))
	if err != nil {
		return "", err
	}

	return lm.GetLibPath(false)
}
