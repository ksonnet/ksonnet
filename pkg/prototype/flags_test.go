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
		"name":     "description",
		"module":   "Component module",
		"optional": "optional",
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
