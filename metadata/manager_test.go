package metadata

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

const (
	blankSwagger     = "/blankSwagger.json"
	blankSwaggerData = `{}`
)

var appFS = afero.NewMemMapFs()

func init() {
	afero.WriteFile(appFS, blankSwagger, []byte(blankSwaggerData), os.ModePerm)
}

func TestInitSuccess(t *testing.T) {
	spec, err := ParseClusterSpec(fmt.Sprintf("file:%s", blankSwagger))
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath("/fromEmptySwagger")
	_, err = initManager(appPath, spec, appFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	defaultEnvDir := appendToAbsPath(schemaDir, defaultEnvName)
	paths := []AbsPath{
		ksonnetDir,
		libDir,
		componentsDir,
		vendorDir,
		schemaDir,
		vendorLibDir,
		defaultEnvDir,
	}

	for _, p := range paths {
		path := appendToAbsPath(appPath, string(p))
		exists, err := afero.DirExists(appFS, string(path))
		if err != nil {
			t.Fatalf("Expected to create directory '%s', but failed:\n%v", p, err)
		} else if !exists {
			t.Fatalf("Expected to create directory '%s', but failed", path)
		}
	}

	envPath := appendToAbsPath(appPath, string(defaultEnvDir))
	schemaPath := appendToAbsPath(envPath, schemaFilename)
	bytes, err := afero.ReadFile(appFS, string(schemaPath))
	if err != nil {
		t.Fatalf("Failed to read swagger file at '%s':\n%v", schemaPath, err)
	} else if actualSwagger := string(bytes); actualSwagger != blankSwaggerData {
		t.Fatalf("Expected swagger file at '%s' to have value: '%s', got: '%s'", schemaPath, blankSwaggerData, actualSwagger)
	}
}

func TestFindSuccess(t *testing.T) {
	findSuccess := func(t *testing.T, appDir, currDir AbsPath) {
		m, err := findManager(currDir, appFS)
		if err != nil {
			t.Fatalf("Failed to find manager at path '%s':\n%v", currDir, err)
		} else if m.rootPath != appDir {
			t.Fatalf("Found manager at incorrect path '%s', expected '%s'", m.rootPath, appDir)
		}
	}

	spec, err := ParseClusterSpec(fmt.Sprintf("file:%s", blankSwagger))
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath("/findSuccess")
	_, err = initManager(appPath, spec, appFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	findSuccess(t, appPath, appPath)

	components := appendToAbsPath(appPath, componentsDir)
	findSuccess(t, appPath, components)

	// Create empty app file.
	appFile := appendToAbsPath(components, "app.jsonnet")
	f, err := appFS.OpenFile(string(appFile), os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		t.Fatalf("Failed to touch app file '%s'\n%v", appFile, err)
	}
	f.Close()

	findSuccess(t, appPath, appFile)
}

func TestFindFailure(t *testing.T) {
	findFailure := func(t *testing.T, currDir AbsPath) {
		_, err := findManager(currDir, appFS)
		if err == nil {
			t.Fatalf("Expected to fail to find ksonnet app in '%s', but succeeded", currDir)
		}
	}

	findFailure(t, "/")
	findFailure(t, "/fakePath")
	findFailure(t, "")
}

func TestDoubleNewFailure(t *testing.T) {
	spec, err := ParseClusterSpec(fmt.Sprintf("file:%s", blankSwagger))
	if err != nil {
		t.Fatalf("Failed to parse cluster spec: %v", err)
	}

	appPath := AbsPath("/doubleNew")

	_, err = initManager(appPath, spec, appFS)
	if err != nil {
		t.Fatalf("Failed to init cluster spec: %v", err)
	}

	targetErr := fmt.Sprintf("Could not create app; directory '%s' already exists", appPath)
	_, err = initManager(appPath, spec, appFS)
	if err == nil || err.Error() != targetErr {
		t.Fatalf("Expected to fail to create app with message '%s', got '%s'", targetErr, err.Error())
	}
}
