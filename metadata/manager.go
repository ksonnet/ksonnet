package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
	"github.com/spf13/afero"
)

func appendToAbsPath(originalPath AbsPath, toAppend ...string) AbsPath {
	paths := append([]string{string(originalPath)}, toAppend...)
	return AbsPath(path.Join(paths...))
}

const (
	ksonnetDir      = ".ksonnet"
	libDir          = "lib"
	componentsDir   = "components"
	environmentsDir = "environments"
	vendorDir       = "vendor"

	defaultEnvName = "default"

	// Environment-specific files
	schemaFilename         = "swagger.json"
	ksonnetLibCoreFilename = "k8s.libsonnet"
	specFilename           = "spec.json"
)

type manager struct {
	appFS afero.Fs

	rootPath        AbsPath
	ksonnetPath     AbsPath
	libPath         AbsPath
	componentsPath  AbsPath
	environmentsDir AbsPath
	vendorDir       AbsPath
}

func findManager(abs AbsPath, appFS afero.Fs) (*manager, error) {
	var lastBase string
	currBase := string(abs)

	for {
		currPath := path.Join(currBase, ksonnetDir)
		exists, err := afero.Exists(appFS, currPath)
		if err != nil {
			return nil, err
		}
		if exists {
			return newManager(AbsPath(currBase), appFS), nil
		}

		lastBase = currBase
		currBase = filepath.Dir(currBase)
		if lastBase == currBase {
			return nil, fmt.Errorf("No ksonnet application found")
		}
	}
}

func initManager(rootPath AbsPath, spec ClusterSpec, appFS afero.Fs) (*manager, error) {
	// Get cluster specification data, possibly from the network.
	specData, err := spec.data()
	if err != nil {
		return nil, err
	}

	m := newManager(rootPath, appFS)

	// Generate the program text for ksonnet-lib.
	//
	// IMPLEMENTATION NOTE: We get the cluster specification and generate
	// ksonnet-lib before initializing the directory structure so that failure of
	// either (e.g., GET'ing the spec from a live cluster returns 404) does not
	// result in a partially-initialized directory structure.
	//
	ksonnetLibDir := appendToAbsPath(m.environmentsDir, defaultEnvName)
	ksonnetLibData, err := generateKsonnetLibData(ksonnetLibDir, specData)
	if err != nil {
		return nil, err
	}

	// Initialize directory structure.
	if err = m.createAppDirTree(); err != nil {
		return nil, err
	}

	// Cache specification data.
	if err = m.createEnvironment(defaultEnvName, specData, ksonnetLibData); err != nil {
		return nil, err
	}

	return m, nil
}

func newManager(rootPath AbsPath, appFS afero.Fs) *manager {
	return &manager{
		appFS: appFS,

		rootPath:        rootPath,
		ksonnetPath:     appendToAbsPath(rootPath, ksonnetDir),
		libPath:         appendToAbsPath(rootPath, libDir),
		componentsPath:  appendToAbsPath(rootPath, componentsDir),
		environmentsDir: appendToAbsPath(rootPath, environmentsDir),
		vendorDir:       appendToAbsPath(rootPath, vendorDir),
	}
}

func (m *manager) Root() AbsPath {
	return m.rootPath
}

func (m *manager) ComponentPaths() (AbsPaths, error) {
	paths := AbsPaths{}
	err := afero.Walk(m.appFS, string(m.componentsPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (m *manager) LibPaths(envName string) (libPath, envLibPath AbsPath) {
	return m.libPath, appendToAbsPath(m.environmentsDir, envName)
}

func (m *manager) createEnvironment(name string, specData, ksonnetLibData []byte) error {
	envPath := appendToAbsPath(m.environmentsDir, name)
	err := m.appFS.MkdirAll(string(envPath), os.ModePerm)
	if err != nil {
		return err
	}

	// Generate the schema file.
	schemaPath := appendToAbsPath(envPath, schemaFilename)
	err = afero.WriteFile(m.appFS, string(schemaPath), specData, os.ModePerm)
	if err != nil {
		return err
	}

	ksonnetLibPath := appendToAbsPath(envPath, ksonnetLibCoreFilename)
	err = afero.WriteFile(m.appFS, string(ksonnetLibPath), ksonnetLibData, 0644)
	return err
}

func (m *manager) createAppDirTree() error {
	exists, err := afero.DirExists(m.appFS, string(m.rootPath))
	if err != nil {
		return fmt.Errorf("Could not check existance of directory '%s':\n%v", m.rootPath, err)
	} else if exists {
		return fmt.Errorf("Could not create app; directory '%s' already exists", m.rootPath)
	}

	paths := []AbsPath{
		m.rootPath,
		m.ksonnetPath,
		m.libPath,
		m.componentsPath,
		m.vendorDir,
	}

	for _, p := range paths {
		if err := m.appFS.MkdirAll(string(p), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func generateKsonnetLibData(ksonnetLibDir AbsPath, text []byte) ([]byte, error) {
	// Deserialize the API object.
	s := kubespec.APISpec{}
	err := json.Unmarshal(text, &s)
	if err != nil {
		return nil, err
	}

	s.Text = text
	s.FilePath = filepath.Dir(string(ksonnetLibDir))

	// Emit Jsonnet code.
	return ksonnet.Emit(&s, nil, nil)
}
