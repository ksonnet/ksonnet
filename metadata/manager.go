package metadata

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/afero"
)

func appendToAbsPath(originalPath AbsPath, toAppend ...string) AbsPath {
	paths := append([]string{string(originalPath)}, toAppend...)
	return AbsPath(path.Join(paths...))
}

const (
	defaultEnvName = "dev"

	ksonnetDir    = ".ksonnet"
	libDir        = "lib"
	componentsDir = "components"
	vendorDir     = "vendor"
	schemaDir     = "vendor/schema"
	vendorLibDir  = "vendor/lib"

	schemaFilename = "swagger.json"
)

type manager struct {
	appFS afero.Fs

	rootPath       AbsPath
	ksonnetPath    AbsPath
	libPath        AbsPath
	componentsPath AbsPath
	vendorDir      AbsPath
	schemaDir      AbsPath
	vendorLibDir   AbsPath
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
	data, err := spec.data()
	if err != nil {
		return nil, err
	}

	m := newManager(rootPath, appFS)

	if err = m.createAppDirTree(); err != nil {
		return nil, err
	}

	if err = m.cacheClusterSpecData(defaultEnvName, data); err != nil {
		return nil, err
	}

	return m, nil
}

func newManager(rootPath AbsPath, appFS afero.Fs) *manager {
	return &manager{
		appFS: appFS,

		rootPath:       rootPath,
		ksonnetPath:    appendToAbsPath(rootPath, ksonnetDir),
		libPath:        appendToAbsPath(rootPath, libDir),
		componentsPath: appendToAbsPath(rootPath, componentsDir),
		vendorDir:      appendToAbsPath(rootPath, vendorDir),
		schemaDir:      appendToAbsPath(rootPath, schemaDir),
		vendorLibDir:   appendToAbsPath(rootPath, vendorLibDir),
	}
}

func (m *manager) Root() AbsPath {
	return m.rootPath
}

func (m *manager) cacheClusterSpecData(name string, specData []byte) error {
	envPath := string(appendToAbsPath(m.schemaDir, name))
	err := m.appFS.MkdirAll(envPath, os.ModePerm)
	if err != nil {
		return err
	}

	schemaPath := string(filepath.Join(envPath, schemaFilename))
	err = afero.WriteFile(m.appFS, schemaPath, specData, os.ModePerm)
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
		m.schemaDir,
		m.vendorLibDir,
	}

	for _, p := range paths {
		if err := m.appFS.MkdirAll(string(p), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
