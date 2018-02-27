package generator

import (
	"errors"
	"testing"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/ksonnet"
)

func TestKsonnet(t *testing.T) {
	ogEmitter := ksonnetEmitter
	defer func() {
		ksonnetEmitter = ogEmitter
	}()

	var (
		ext            = []byte("k")
		lib            = []byte("k8s")
		successfulEmit = func(string) (*ksonnet.Lib, error) {
			return &ksonnet.Lib{
				Version:    "v1.7.0",
				K8s:        lib,
				Extensions: ext,
			}, nil
		}
		failureEmit = func(string) (*ksonnet.Lib, error) {
			return nil, errors.New("failure")
		}
		v170swagger = []byte(`{"info":{"version":"v1.7.0"}}`)
	)

	cases := []struct {
		name        string
		emitter     func(string) (*ksonnet.Lib, error)
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
			emitter:     failureEmit,
			swaggerData: []byte(`{`),
			isErr:       true,
		},
		{
			name:        "emitter error",
			emitter:     failureEmit,
			swaggerData: v170swagger,
			isErr:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ogKSEmitter := ksonnetEmitter
			defer func() { ksonnetEmitter = ogKSEmitter }()
			ksonnetEmitter = tc.emitter

			kl, err := Ksonnet(tc.swaggerData)

			if tc.isErr {
				if err == nil {
					t.Fatal("Ksonnet() should have returned an error")
				}
			} else {
				if err != nil {
					t.Fatalf("Ksonnet() returned unexpected error: %#v", err)
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
