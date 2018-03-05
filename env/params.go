package env

import (
	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/metadata/app"
	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// SetParamsConfig is config items for setting environment params.
type SetParamsConfig struct {
	AppRoot string
	Fs      afero.Fs
}

// SetParams sets params for an environment.
func SetParams(envName, component string, params param.Params, config SetParamsConfig) error {
	exists, err := envExists(config.Fs, config.AppRoot, envName)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf("Environment %q does not exist", envName)
	}

	path := envPath(config.AppRoot, envName, paramsFileName)

	text, err := afero.ReadFile(config.Fs, path)
	if err != nil {
		return err
	}

	appended, err := param.SetEnvironmentParams(component, string(text), params)
	if err != nil {
		return err
	}

	err = afero.WriteFile(config.Fs, path, []byte(appended), app.DefaultFilePermissions)
	if err != nil {
		return err
	}

	log.Debugf("Successfully set parameters for component %q at environment %q", component, envName)
	return nil
}

// GetParamsConfig is config items for getting environment params.
type GetParamsConfig struct {
	AppRoot string
	Fs      afero.Fs
}

// GetParams gets all parameters for an environment.
func GetParams(envName, nsName string, config GetParamsConfig) (map[string]param.Params, error) {
	exists, err := envExists(config.Fs, config.AppRoot, envName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("Environment %q does not exist", envName)
	}

	// Get the environment specific params
	envParamsPath := envPath(config.AppRoot, envName, paramsFileName)
	envParamsText, err := afero.ReadFile(config.Fs, envParamsPath)
	if err != nil {
		return nil, err
	}
	envParams, err := param.GetAllEnvironmentParams(string(envParamsText))
	if err != nil {
		return nil, err
	}

	// figure out what component we need
	ns := component.NewNamespace(config.Fs, config.AppRoot, nsName)
	componentParamsFile, err := afero.ReadFile(config.Fs, ns.ParamsPath())
	if err != nil {
		return nil, err
	}

	componentParams, err := param.GetAllComponentParams(string(componentParamsFile))
	if err != nil {
		return nil, err
	}

	return mergeParamMaps(componentParams, envParams), nil
}

// TODO: move this to the consolidated params support namespace.
func mergeParamMaps(base, overrides map[string]param.Params) map[string]param.Params {
	for component, params := range overrides {
		if _, contains := base[component]; !contains {
			base[component] = params
		} else {
			for k, v := range params {
				base[component][k] = v
			}
		}
	}
	return base
}
