package generator

import (
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
)

var (
	// ksonnetEmitter is the function which emits the ksonnet standard library.
	ksonnetEmitter = ksonnet.GenerateLib
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
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}

	defer os.Remove(f.Name())

	_, err = f.Write(swaggerData)
	if err != nil {
		return nil, err
	}

	if err = f.Close(); err != nil {
		return nil, err
	}

	spew.Dump("---", f.Name(), ksonnetEmitter)

	lib, err := ksonnetEmitter(f.Name())
	if err != nil {
		return nil, err
	}

	kl := &KsonnetLib{
		K:       lib.Extensions,
		K8s:     lib.K8s,
		Swagger: swaggerData,
		Version: lib.Version,
	}

	return kl, nil
}
