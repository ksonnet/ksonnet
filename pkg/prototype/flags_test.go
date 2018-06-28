// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package prototype

import (
	"sort"
	"testing"

	"github.com/spf13/afero"

	"github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagDefinitionError(t *testing.T) {
	err := FlagDefinitionError{name: "name"}
	got := err.Error()

	expected := `unable to define flag "name"`
	require.Equal(t, expected, got)
}

func TestBindFlags(t *testing.T) {
	p := &Prototype{
		Params: ParamSchemas{
			{
				Name:        "name",
				Description: "description",
				Type:        String,
			},
			{
				Name:        "optional",
				Description: "optional",
				Type:        String,
				Default:     strings.Ptr("value"),
			},
		},
	}

	flags, err := BindFlags(p)
	require.NoError(t, err)

	expectedFlags := map[string]string{
		"name":        "description",
		"module":      "Component module",
		"optional":    "optional",
		"values-file": "Prototype values file (file returns a Jsonnet object)",
	}

	var seenFlags []string

	flags.VisitAll(func(f *pflag.Flag) {
		desc, ok := expectedFlags[f.Name]
		if assert.True(t, ok, "unexpected flag %q", f.Name) {
			assert.Equal(t, desc, f.Usage, "flag %q usage was not expected")
			seenFlags = append(seenFlags, f.Name)
		}
	})

	var expectedKeys []string
	for k := range expectedFlags {
		expectedKeys = append(expectedKeys, k)
	}
	sort.Strings(expectedKeys)
	sort.Strings(seenFlags)

	assert.Equal(t, expectedKeys, seenFlags, "did not see all expected flags")
}

func TestBindFlags_duplicate_required_param(t *testing.T) {
	p := &Prototype{
		Params: ParamSchemas{
			{
				Name:        "module",
				Description: "module",
				Type:        String,
			},
		},
	}

	_, err := BindFlags(p)
	require.Error(t, err)
}

func TestBindFlags_duplicate_optional_param(t *testing.T) {
	p := &Prototype{
		Params: ParamSchemas{
			{
				Name:        "module",
				Description: "module",
				Default:     strings.Ptr("value"),
				Type:        String,
			},
		},
	}

	_, err := BindFlags(p)
	require.Error(t, err)
}

func TestExtractParameters(t *testing.T) {
	validPrototype := &Prototype{
		APIVersion: "0.1",
		Name:       "io.ksonnet.pkg.configMap",
		Params: ParamSchemas{
			{
				Name:        "name",
				Description: "Name to give the configMap",
				Type:        String,
			},
			{
				Name:        "data",
				Description: "Data for the configMap",
				Default:     strings.Ptr("{}"),
				Type:        Object,
			},
			{
				Name:        "val",
				Description: "Value",
				Default:     strings.Ptr("9"),
				Type:        Number,
			},
		},
	}

	cases := []struct {
		name      string
		p         *Prototype
		initFlags func(*testing.T, *Prototype, []string) *pflag.FlagSet
		initFs    func(t *testing.T, fs afero.Fs)
		args      []string
		expected  map[string]string
		isErr     bool
	}{
		{
			name: "with valid flags",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags, err := BindFlags(p)
				require.NoError(t, err)

				err = flags.Parse(args)
				require.NoError(t, err)

				return flags
			},
			args: []string{
				"--name=name",
			},
			expected: map[string]string{
				"data": `{}`,
				"name": `"name"`,
				"val":  `9`,
			},
		},
		{
			name: "values from file",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags, err := BindFlags(p)
				require.NoError(t, err)

				err = flags.Parse(args)
				require.NoError(t, err)

				return flags
			},
			initFs: func(t *testing.T, fs afero.Fs) {
				data := []byte(`{name: "name"}`)
				afero.WriteFile(fs, "/values-file", data, 0644)
			},
			args: []string{
				"--values-file=/values-file",
			},
			expected: map[string]string{
				"data": `{}`,
				"name": `"name"`,
				"val":  `9`,
			},
		},
		{
			name: "missing a required flag",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags, err := BindFlags(p)
				require.NoError(t, err)

				err = flags.Parse(args)
				require.NoError(t, err)

				return flags
			},
			args:  []string{},
			isErr: true,
		},
		{
			name: "flag not defined",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags := pflag.NewFlagSet("prototype-flags", pflag.ContinueOnError)
				return flags
			},
			args: []string{
				"--name=name",
				"--undefined=a",
			},
			isErr: true,
		},
		{
			name: "missing values file",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags, err := BindFlags(p)
				require.NoError(t, err)

				err = flags.Parse(args)
				require.NoError(t, err)

				return flags
			},
			args: []string{
				"--values-file=/missing-values-file",
			},
			isErr: true,
		},
		{
			name: "valid file does not parse as valid jsonnet",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags, err := BindFlags(p)
				require.NoError(t, err)

				err = flags.Parse(args)
				require.NoError(t, err)

				return flags
			},
			initFs: func(t *testing.T, fs afero.Fs) {
				data := []byte(`{`)
				afero.WriteFile(fs, "/values-file", data, 0644)
			},
			args: []string{
				"--values-file=/values-file",
			},
			isErr: true,
		},
		{
			name: "valid file does not contain object",
			p:    validPrototype,
			initFlags: func(t *testing.T, p *Prototype, args []string) *pflag.FlagSet {
				flags, err := BindFlags(p)
				require.NoError(t, err)

				err = flags.Parse(args)
				require.NoError(t, err)

				return flags
			},
			initFs: func(t *testing.T, fs afero.Fs) {
				data := []byte(`[]`)
				afero.WriteFile(fs, "/values-file", data, 0644)
			},
			args: []string{
				"--values-file=/values-file",
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flags := tc.initFlags(t, tc.p, tc.args)
			fs := afero.NewMemMapFs()

			if tc.initFs != nil {
				tc.initFs(t, fs)
			}

			m, err := ExtractParameters(fs, tc.p, flags)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expected, m)
		})
	}
}
