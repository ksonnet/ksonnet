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

package jsonnet

import (
	"testing"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stubVMOpt() VMOpt {
	return func(vm *VM) {
		vm.makeVMFn = func() *jsonnet.VM { return &jsonnet.VM{} }
	}
}

func stubVMEvalOpt(fn func(*jsonnet.VM, string, string) (string, error)) VMOpt {
	return func(vm *VM) {
		vm.evaluateSnippetFn = fn
	}
}

func TestVM_ExtCode(t *testing.T) {
	vm := NewVM(stubVMOpt())
	vm.ExtCode("key", "value")
	require.Equal(t, "value", vm.extCodes["key"])
}

func TestVM_TLACode(t *testing.T) {
	vm := NewVM(stubVMOpt())
	vm.TLACode("key", "value")
	require.Equal(t, "value", vm.tlaCodes["key"])
}

func TestVM_TLAVar(t *testing.T) {
	vm := NewVM(stubVMOpt())
	vm.TLAVar("key", "value")
	require.Equal(t, "value", vm.tlaVars["key"])
}

func TestVM_EvaluateSnippet(t *testing.T) {

	fn := func(vm *jsonnet.VM, name, snippet string) (string, error) {
		assert.Equal(t, "snippet", name)
		assert.Equal(t, "code", snippet)

		return "evaluated", nil
	}

	vm := NewVM(stubVMOpt(), stubVMEvalOpt(fn))
	vm.TLAVar("key", "value")
	vm.TLACode("key", "value")
	vm.ExtCode("key", "value")

	out, err := vm.EvaluateSnippet("snippet", "code")
	require.NoError(t, err)

	require.Equal(t, "evaluated", out)
}

func TestVM_EvaluateSnippet_memory_importer(t *testing.T) {
	fs := afero.NewMemMapFs()
	test.StageFile(t, fs, "set-map.jsonnet", "/lib/set-map.jsonnet")

	fn := func(vm *jsonnet.VM, name, snippet string) (string, error) {
		assert.Equal(t, "snippet", name)
		assert.Equal(t, "code", snippet)

		return "evaluated", nil
	}

	vm := NewVM(stubVMOpt(), stubVMEvalOpt(fn), AferoImporterOpt(fs))
	vm.AddJPath("/lib")

	vm.TLAVar("key", "value")
	vm.TLACode("key", "value")
	vm.ExtCode("key", "value")
	vm.ExtVar("key", "value")

	out, err := vm.EvaluateSnippet("snippet", "code")
	require.NoError(t, err)

	require.Equal(t, "evaluated", out)
}

func Test_regexSubst(t *testing.T) {
	cases := []struct {
		name     string
		in       []interface{}
		expected string
		isErr    bool
	}{
		{
			name: "valid regex",
			in: []interface{}{
				"ee",
				"tree",
				"oll",
			},
			expected: "troll",
		},
		{
			name: "invalid regex",
			in: []interface{}{
				"[",
				"tree",
				"oll",
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := regexSubst(tc.in)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			s, ok := out.(string)
			require.True(t, ok)

			require.Equal(t, tc.expected, s)
		})
	}
}

func Test_regexMatch(t *testing.T) {
	in := []interface{}{"ee", "tree"}
	out, err := regexMatch(in)
	require.NoError(t, err)

	tf, ok := out.(bool)
	require.True(t, ok)
	require.True(t, tf)
}

func Test_escapeStringRegex(t *testing.T) {
	in := []interface{}{"[foo]"}
	out, err := escapeStringRegex(in)
	require.NoError(t, err)

	s, ok := out.(string)
	require.True(t, ok)

	require.Equal(t, `\[foo\]`, s)
}

func Test_parseYAML(t *testing.T) {
	cases := []struct {
		name     string
		in       []interface{}
		expected interface{}
		isErr    bool
	}{
		{
			name: "valid yaml",
			in:   []interface{}{"---\nfoo: bar"},
			expected: []interface{}{
				map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := parseYAML(tc.in)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, out)
		})
	}
}

func Test_parseJSON(t *testing.T) {
	cases := []struct {
		name     string
		in       []interface{}
		expected interface{}
		isErr    bool
	}{
		{
			name: "valid JSON",
			in:   []interface{}{`{ "foo": "bar" }`},
			expected: map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := parseJSON(tc.in)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, out)
		})
	}
}

func TestParseJson(t *testing.T) {
	vm := NewVM()

	_, err := vm.EvaluateSnippet("failtest", `std.native("parseJson")("barf{")`)
	require.Error(t, err)

	x, err := vm.EvaluateSnippet("test", `std.native("parseJson")("null")`)
	require.NoError(t, err)
	assert.Equal(t, "null\n", x)

	x, err = vm.EvaluateSnippet("test", `
    local a = std.native("parseJson")('{"foo": 3, "bar": 4}');
    a.foo + a.bar`)
	require.NoError(t, err)
	assert.Equal(t, "7\n", x)
}

func TestParseYaml(t *testing.T) {
	vm := NewVM()

	_, err := vm.EvaluateSnippet("failtest", `std.native("parseYaml")("[barf")`)
	require.Error(t, err)

	x, err := vm.EvaluateSnippet("test", `std.native("parseYaml")("")`)
	require.NoError(t, err)
	assert.Equal(t, "[ ]\n", x)

	x, err = vm.EvaluateSnippet("test", `
    local a = std.native("parseYaml")("foo:\n- 3\n- 4\n")[0];
    a.foo[0] + a.foo[1]`)
	require.NoError(t, err)
	assert.Equal(t, "7\n", x)

	x, err = vm.EvaluateSnippet("test", `
    local a = std.native("parseYaml")("---\nhello\n---\nworld");
    a[0] + a[1]`)
	require.NoError(t, err)
	assert.Equal(t, "\"helloworld\"\n", x)
}

func Test_regexMatch_fun(t *testing.T) {
	vm := NewVM()

	_, err := vm.EvaluateSnippet("failtest", `std.native("regexMatch")("[f", "foo")`)
	require.Error(t, err)

	x, err := vm.EvaluateSnippet("test", `std.native("regexMatch")("foo.*", "seafood")`)
	require.NoError(t, err)
	assert.Equal(t, "true\n", x)

	x, err = vm.EvaluateSnippet("test", `std.native("regexMatch")("bar.*", "seafood")`)
	require.NoError(t, err)
	assert.Equal(t, "false\n", x)
}

func TestRegexSubst(t *testing.T) {
	vm := NewVM()

	_, err := vm.EvaluateSnippet("failtest", `std.native("regexSubst")("[f",s "foo", "bar")`)
	require.Error(t, err)

	x, err := vm.EvaluateSnippet("test", `std.native("regexSubst")("a(x*)b", "-ab-axxb-", "T")`)
	require.NoError(t, err)
	assert.Equal(t, "\"-T-T-\"\n", x)

	x, err = vm.EvaluateSnippet("test", `std.native("regexSubst")("a(x*)b", "-ab-axxb-", "${1}W")`)
	require.NoError(t, err)
	assert.Equal(t, "\"-W-xxW-\"\n", x)
}
