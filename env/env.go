package env

import (
	"github.com/ksonnet/ksonnet/metadata/app"
)

const (
	// primary environment files.
	envFileName    = "main.jsonnet"
	paramsFileName = "params.libsonnet"

	// envRoot is the name for the environment root.
	envRoot = "environments"
)

// Env represents a ksonnet environment.
type Env struct {
	// Name is the environment name.
	Name string
	// KubernetesVersion is the version of Kubernetes for this environment.
	KubernetesVersion string
	// Destination is the cluster destination for this environment.
	Destination Destination
	// Targets are the component namespaces that will be installed.
	Targets []string
}

func envFromSpec(name string, envSpec *app.EnvironmentSpec) *Env {
	return &Env{
		Name:              name,
		KubernetesVersion: envSpec.KubernetesVersion,
		Destination:       NewDestination(envSpec.Destination.Server, envSpec.Destination.Namespace),
		Targets:           envSpec.Targets,
	}
}

// List lists all environments for the current ksonnet application.
func List(ksApp app.App) (map[string]Env, error) {
	envs := make(map[string]Env)

	specs, err := ksApp.Environments()
	if err != nil {
		return nil, err
	}

	for name, spec := range specs {
		env := envFromSpec(name, spec)
		envs[name] = *env
	}

	return envs, nil
}

// Retrieve retrieves an environment by name.
func Retrieve(ksApp app.App, name string) (*Env, error) {
	envSpec, err := ksApp.Environment(name)
	if err != nil {
		return nil, err
	}

	return envFromSpec(name, envSpec), nil
}
