package generator

import (
	"errors"
	"testing"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
)

func TestKsonnet(t *testing.T) {
	ogEmitter := ksonnetEmitter
	defer func() {
		ksonnetEmitter = ogEmitter
	}()

	var (
		ext            = []byte("k")
		lib            = []byte("k8s")
		successfulEmit = func(*kubespec.APISpec, *string, *string) ([]byte, []byte, error) {
			return ext, lib, nil
		}
		failureEmit = func(*kubespec.APISpec, *string, *string) ([]byte, []byte, error) {
			return nil, nil, errors.New("failure")
		}
		v170swagger = []byte(`{"info":{"version":"v1.7.0"}}`)
		v180swagger = []byte(`{"info":{"version":"v1.8.0"}}`)
	)

	cases := []struct {
		name        string
		emitter     func(*kubespec.APISpec, *string, *string) ([]byte, []byte, error)
		swaggerData []byte
		version     string
		isErr       bool
	}{
		{
			name:        "valid swagger",
			emitter:     successfulEmit,
			swaggerData: v170swagger,
			version:     "v1.7.0",
		},
		{
			name:        "invalid swagger",
			swaggerData: []byte(`{`),
			isErr:       true,
		},
		{
			name:        "emitter error",
			emitter:     failureEmit,
			swaggerData: v170swagger,
			isErr:       true,
		},
		{
			name:        "valid beta swagger",
			emitter:     successfulEmit,
			swaggerData: v180swagger,
			version:     "v1.8.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ksonnetEmitter = tc.emitter

			kl, err := Ksonnet(tc.swaggerData)

			if tc.isErr {
				if err == nil {
					t.Fatal("Ksonnet() should have returned an error")
				}
			} else {
				if err != nil {
					t.Fatal("Ksonnet() returned unexpected error")
				}

				if got, expected := string(kl.K), string(ext); got != expected {
					t.Errorf("Ksonnet() K = %s; expected = %s", got, expected)
				}
				if got, expected := string(kl.K8s), string(lib); got != expected {
					t.Errorf("Ksonnet() K8s = %s; expected = %s", got, expected)
				}
				if got, expected := string(kl.Swagger), string(tc.swaggerData); got != expected {
					t.Errorf("Ksonnet() Swagger = %s; expected = %s", got, expected)
				}
				if got, expected := string(kl.Version), tc.version; got != expected {
					t.Errorf("Ksonnet() Version = %s; expected = %s", got, expected)
				}
			}
		})
	}

}
