package generator

import (
	"encoding/json"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubeversion"
	log "github.com/sirupsen/logrus"
)

var (
	// ksonnetEmitter is the function which emits the ksonnet standard library.
	ksonnetEmitter = ksonnet.Emit
)

// KsonnetLib is the ksonnet standard library for a version of swagger.
type KsonnetLib struct {
	// K is ksonnet extensions.
	K []byte
	// K is the generated ksonnet library.
	K8s []byte
	// Swagger is the swagger JSON used to generate the library.
	Swagger []byte
	// Version is the API version of the swagger.
	Version string
}

// Ksonnet generates the ksonnet standard library or returns an error if there was
// a problem.
func Ksonnet(swaggerData []byte) (*KsonnetLib, error) {
	// Deserialize the API object.
	s := kubespec.APISpec{}
	if err := json.Unmarshal(swaggerData, &s); err != nil {
		return nil, err
	}

	s.Text = swaggerData

	// Emit Jsonnet code.
	extensionsLibData, k8sLibData, err := ksonnetEmitter(&s, nil, nil)
	if err != nil {
		return nil, err
	}

	// Warn where the Kubernetes version is currently only supported as Beta.
	if kubeversion.Beta(s.Info.Version) {
		log.Warnf(`!
============================================================================================
Kubernetes version %s is currently supported as Beta; you may encounter unexpected behavior
============================================================================================`, s.Info.Version)
	}

	kl := &KsonnetLib{
		K:       extensionsLibData,
		K8s:     k8sLibData,
		Swagger: swaggerData,
		Version: s.Info.Version,
	}

	return kl, nil
}
