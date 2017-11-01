// Copyright 2017 The kubecfg authors
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

package params

import (
	"reflect"
	"testing"
)

func TestAppendComponentParams(t *testing.T) {
	tests := []struct {
		componentName string
		jsonnet       string
		params        Params
		expected      string
	}{
		// Test case with existing components
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
  components: {
    // Component-level parameters, defined initially from 'ks prototype use ...'
    // Each object below should correspond to a component in the components/ directory
    foo: {
      name: "foo",
      replicas: 1,
    },
    bar: {
      name: "bar",
    },
  },
}`,
			Params{"replicas": "5", "name": `"baz"`},
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
  components: {
    // Component-level parameters, defined initially from 'ks prototype use ...'
    // Each object below should correspond to a component in the components/ directory
    foo: {
      name: "foo",
      replicas: 1,
    },
    bar: {
      name: "bar",
    },
    baz: {
      name: "baz",
      replicas: 5,
    },
  },
}`,
		},
		// Test case with no existing components
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
  components: {
    // Component-level parameters, defined initially from 'ks prototype use ...'
    // Each object below should correspond to a component in the components/ directory
  },
}`,
			Params{"replicas": "5", "name": `"baz"`},
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
  components: {
    // Component-level parameters, defined initially from 'ks prototype use ...'
    // Each object below should correspond to a component in the components/ directory
    baz: {
      name: "baz",
      replicas: 5,
    },
  },
}`,
		},
	}

	errors := []struct {
		componentName string
		jsonnet       string
		params        Params
	}{
		// Test case where there isn't a components object
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
  },
}`,
			Params{"replicas": "5", "name": `"baz"`},
		},
		// Test case where components isn't a top level object
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
		components: {},
  },
}`,
			Params{"replicas": "5", "name": `"baz"`},
		},
		// Test case where component already exists
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
	},
  components: {
    // Component-level parameters, defined initially from 'ks prototype use ...'
    // Each object below should correspond to a component in the components/ directory
    baz: {
      name: "baz",
      replicas: 5,
    },
  },	
}`,
			Params{"replicas": "5", "name": `"baz"`},
		},
	}

	for _, s := range tests {
		parsed, err := AppendComponent(s.componentName, s.jsonnet, s.params)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if parsed != s.expected {
			t.Errorf("Wrong conversion\n  expected: %v\n  got: %v", s.expected, parsed)
		}
	}

	for _, e := range errors {
		parsed, err := AppendComponent(e.componentName, e.jsonnet, e.params)
		if err == nil {
			t.Errorf("Expected error but not found\n  input: %v  got: %v", e, parsed)
		}
	}
}

func TestGetComponentParams(t *testing.T) {
	tests := []struct {
		componentName string
		jsonnet       string
		expected      Params
	}{
		// Test getting the parameters where there is a single component
		{
			"foo",
			`
{
  global: {},
  components: {
    foo: {
      name: "foo",
      replicas: 1,
    },
  },
}`,
			Params{"name": `"foo"`, "replicas": "1"},
		},
		// Test getting the parameters where there are multiple components
		{
			"foo",
			`
{
  global: {},
  components: {
    bar: {
      replicas: 5
    },
    foo: {
      name: "foo",
      replicas: 1,
    },
  },
}`,
			Params{"name": `"foo"`, "replicas": "1"},
		},
	}

	errors := []struct {
		componentName string
		jsonnet       string
	}{
		// Test case where component doesn't exist
		{
			"baz",
			`
{
  components: {
    foo: {
      name: "foo",
    },
  },
}`,
		},
		// Test case where components isn't a top level object
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
		components: {},
  },
}`,
		},
	}

	for _, s := range tests {
		params, err := GetComponentParams(s.componentName, s.jsonnet)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if !reflect.DeepEqual(params, s.expected) {
			t.Errorf("Wrong conversion\n  expected:%v\n  got:%v", s.expected, params)
		}
	}

	for _, e := range errors {
		params, err := GetComponentParams(e.componentName, e.jsonnet)
		if err == nil {
			t.Errorf("Expected error but not found\n  input: %v  got: %v", e, params)
		}
	}
}

func TestGetAllComponentParams(t *testing.T) {
	tests := []struct {
		jsonnet  string
		expected map[string]Params
	}{
		// Test getting the parameters where there are zero components
		{
			`
{
  global: {},
  components: {
  },
}`,
			map[string]Params{},
		},
		// Test getting the parameters where there is a single component
		{
			`
{
  global: {},
  components: {
    bar: {
      replicas: 5
    },
  },
}`,
			map[string]Params{
				"bar": Params{"replicas": "5"},
			},
		},
		// Test getting the parameters where there are multiple components
		{
			`
{
  global: {},
  components: {
    bar: {
      replicas: 5
    },
    foo: {
      name: "foo",
      replicas: 1,
    },
  },
}`,
			map[string]Params{
				"bar": Params{"replicas": "5"},
				"foo": Params{"name": `"foo"`, "replicas": "1"},
			},
		},
	}

	for _, s := range tests {
		params, err := GetAllComponentParams(s.jsonnet)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if !reflect.DeepEqual(params, s.expected) {
			t.Errorf("Wrong conversion\n  expected:%v\n  got:%v", s.expected, params)
		}
	}
}

func TestSetComponentParams(t *testing.T) {
	tests := []struct {
		componentName string
		jsonnet       string
		params        Params
		expected      string
	}{
		// Test setting one parameter
		{
			"foo",
			`
{
  global: {},
  components: {
    foo: {
      name: "foo",
      replicas: 1,
    },
    bar: {
      name: "bar",
    },
  },
}`,
			Params{"replicas": "5"},
			`
{
  global: {},
  components: {
    foo: {
      name: "foo",
      replicas: 5,
    },
    bar: {
      name: "bar",
    },
  },
}`,
		},
		// Test setting multiple parameters
		{
			"foo",
			`
{
  components: {
    foo: {
      name: "foo",
      replicas: 1,
    },
  },
}`,
			Params{"replicas": "5", "name": `"foobar"`},
			`
{
  components: {
    foo: {
      name: "foobar",
      replicas: 5,
    },
  },
}`,
		},
		// Test setting parameter that does not exist -- this should add the param
		{
			"foo",
			`
{
  components: {
    foo: {
      name: "foo",
    },
  },
}`,
			Params{"replicas": "5"},
			`
{
  components: {
    foo: {
      name: "foo",
      replicas: 5,
    },
  },
}`,
		},
	}

	errors := []struct {
		componentName string
		jsonnet       string
		params        Params
	}{
		// Test case where component doesn't exist
		{
			"baz",
			`
{
  components: {
    foo: {
      name: "foo",
    },
  },
}`,
			Params{"name": `"baz"`},
		},
		// Test case where components isn't a top level object
		{
			"baz",
			`
{
  global: {
    // User-defined global parameters; accessible to all component and environments, Ex:
    // replicas: 4,
		components: {},
  },
}`,
			Params{"replicas": "5", "name": `"baz"`},
		},
	}

	for _, s := range tests {
		parsed, err := SetComponentParams(s.componentName, s.jsonnet, s.params)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if parsed != s.expected {
			t.Errorf("Wrong conversion\n  expected:%v\n  got:%v", s.expected, parsed)
		}
	}

	for _, e := range errors {
		parsed, err := SetComponentParams(e.componentName, e.jsonnet, e.params)
		if err == nil {
			t.Errorf("Expected error but not found\n  input: %v  got: %v", e, parsed)
		}
	}
}

func TestGetAllEnvironmentParams(t *testing.T) {
	tests := []struct {
		jsonnet  string
		expected map[string]Params
	}{
		// Test getting the parameters where there are zero components
		{
			`
local params = import "/fake/path";
params + {
  components +: {
  },
}`,
			map[string]Params{},
		},
		// Test getting the parameters where there is a single component
		{
			`
local params = import "/fake/path";
params + {
  components +: {
    bar +: {
      name: "bar",
      replicas: 1,
    },
  },
}`,
			map[string]Params{
				"bar": Params{"name": `"bar"`, "replicas": "1"},
			},
		},
		// Test getting the parameters where there are multiple components
		{
			`
local params = import "/fake/path";
params + {
  components +: {
    bar +: {
      name: "bar",
      replicas: 1,
    },
    foo +: {
      name: "foo",
      replicas: 5,
    },
  },
}`,
			map[string]Params{
				"bar": Params{"name": `"bar"`, "replicas": "1"},
				"foo": Params{"name": `"foo"`, "replicas": "5"},
			},
		},
	}

	for _, s := range tests {
		params, err := GetAllEnvironmentParams(s.jsonnet)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if !reflect.DeepEqual(params, s.expected) {
			t.Errorf("Wrong conversion\n  expected:%v\n  got:%v", s.expected, params)
		}
	}
}

func TestSetEnvironmentParams(t *testing.T) {
	tests := []struct {
		componentName string
		jsonnet       string
		params        Params
		expected      string
	}{
		// Test environment param case
		{
			"foo",
			`
local params = import "/fake/path";
params + {
  components +: {
    foo +: {
      name: "foo",
      replicas: 1,
    },
  },
}`,
			Params{"replicas": "5"},
			`
local params = import "/fake/path";
params + {
  components +: {
    foo +: {
      name: "foo",
      replicas: 5,
    },
  },
}`,
		},
		// Test environment param case with multiple components
		{
			"foo",
			`
local params = import "/fake/path";
params + {
  components +: {
    bar +: {
      name: "bar",
      replicas: 1,
    },
    foo +: {
      name: "foo",
      replicas: 1,
    },
  },
}`,
			Params{"name": `"foobar"`, "replicas": "5"},
			`
local params = import "/fake/path";
params + {
  components +: {
    bar +: {
      name: "bar",
      replicas: 1,
    },
    foo +: {
      name: "foobar",
      replicas: 5,
    },
  },
}`,
		},
		// Test setting environment param case where component isn't in the snippet
		{
			"foo",
			`
local params = import "/fake/path";
params + {
  components +: {
  },
}`,
			Params{"replicas": "5"},
			`
local params = import "/fake/path";
params + {
  components +: {
    foo +: {
      replicas: 5,
    },
  },
}`,
		},
	}

	errors := []struct {
		componentName string
		jsonnet       string
		params        Params
	}{
		// Test bad schema
		{
			"foo",
			`
local params = import "/fake/path";
params + {
  badobj +: {
  },
}`,
			Params{"replicas": "5"},
		},
	}

	for _, s := range tests {
		parsed, err := SetEnvironmentParams(s.componentName, s.jsonnet, s.params)
		if err != nil {
			t.Errorf("Unexpected error\n  input: %v\n  error: %v", s.jsonnet, err)
		}

		if parsed != s.expected {
			t.Errorf("Wrong conversion\n  expected:%v\n  got:%v", s.expected, parsed)
		}
	}

	for _, e := range errors {
		parsed, err := SetEnvironmentParams(e.componentName, e.jsonnet, e.params)
		if err == nil {
			t.Errorf("Expected error but not found\n  input: %v  got: %v", e, parsed)
		}
	}
}
