package env

import (
	"reflect"
	"testing"

	"github.com/ksonnet/ksonnet/metadata/params"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestSetParams(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		config := SetParamsConfig{
			AppRoot: "/",
			Fs:      fs,
		}

		p := params.Params{
			"foo": "bar",
		}

		err := SetParams("env1", "component1", p, config)
		require.NoError(t, err)

		compareOutput(t, fs, "updated-params.libsonnet", "/environments/env1/params.libsonnet")
	})
}

func TestGetParams(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		config := GetParamsConfig{
			AppRoot: "/",
			Fs:      fs,
		}

		p, err := GetParams("env1", "", config)
		require.NoError(t, err)

		expected := map[string]params.Params{
			"component1": params.Params{
				"foo": `"bar"`,
			},
		}

		require.Equal(t, expected, p)
	})
}

func TestMergeParamMaps(t *testing.T) {
	tests := []struct {
		base      map[string]params.Params
		overrides map[string]params.Params
		expected  map[string]params.Params
	}{
		{
			map[string]params.Params{
				"bar": params.Params{"replicas": "5"},
			},
			map[string]params.Params{
				"foo": params.Params{"name": `"foo"`, "replicas": "1"},
			},
			map[string]params.Params{
				"bar": params.Params{"replicas": "5"},
				"foo": params.Params{"name": `"foo"`, "replicas": "1"},
			},
		},
		{
			map[string]params.Params{
				"bar": params.Params{"replicas": "5"},
			},
			map[string]params.Params{
				"bar": params.Params{"name": `"foo"`},
			},
			map[string]params.Params{
				"bar": params.Params{"name": `"foo"`, "replicas": "5"},
			},
		},
		{
			map[string]params.Params{
				"bar": params.Params{"name": `"bar"`, "replicas": "5"},
				"foo": params.Params{"name": `"foo"`, "replicas": "4"},
				"baz": params.Params{"name": `"baz"`, "replicas": "3"},
			},
			map[string]params.Params{
				"foo": params.Params{"replicas": "1"},
				"baz": params.Params{"name": `"foobaz"`},
			},
			map[string]params.Params{
				"bar": params.Params{"name": `"bar"`, "replicas": "5"},
				"foo": params.Params{"name": `"foo"`, "replicas": "1"},
				"baz": params.Params{"name": `"foobaz"`, "replicas": "3"},
			},
		},
	}

	for _, s := range tests {
		result := mergeParamMaps(s.base, s.overrides)
		if !reflect.DeepEqual(s.expected, result) {
			t.Errorf("Wrong merge\n  expected:\n%v\n  got:\n%v", s.expected, result)
		}
	}
}
