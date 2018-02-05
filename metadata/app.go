package metadata

import (
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/spf13/afero"
)

// AppSpec will return the specification for a ksonnet application
// (typically stored in `app.yaml`)
func (m *manager) AppSpec() (*app.Spec, error) {
	bytes, err := afero.ReadFile(m.appFS, string(m.appYAMLPath))
	if err != nil {
		return nil, err
	}

	schema, err := app.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}

	if schema.Contributors == nil {
		schema.Contributors = app.ContributorSpecs{}
	}

	if schema.Registries == nil {
		schema.Registries = app.RegistryRefSpecs{}
	}

	if schema.Libraries == nil {
		schema.Libraries = app.LibraryRefSpecs{}
	}

	if schema.Environments == nil {
		schema.Environments = app.EnvironmentSpecs{}
	}

	return schema, nil
}

// WriteAppSpec writes the provided spec to the app.yaml file.
func (m *manager) WriteAppSpec(appSpec *app.Spec) error {
	appSpecData, err := appSpec.Marshal()
	if err != nil {
		return err
	}

	return afero.WriteFile(m.appFS, string(m.appYAMLPath), appSpecData, defaultFilePermissions)
}
