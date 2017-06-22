package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func cmdOutput(t *testing.T, args []string) string {
	var buf bytes.Buffer
	RootCmd.SetOutput(&buf)
	defer RootCmd.SetOutput(nil)

	t.Log("Running args", args)
	RootCmd.SetArgs(args)
	if err := RootCmd.Execute(); err != nil {
		t.Fatal("command failed:", err)
	}

	return buf.String()
}

func TestShow(t *testing.T) {
	formats := map[string]func(string) (interface{}, error){
		"json": func(text string) (ret interface{}, err error) {
			err = json.Unmarshal([]byte(text), &ret)
			return
		},
		"yaml": func(text string) (ret interface{}, err error) {
			err = yaml.Unmarshal([]byte(text), &ret)
			return
		},
	}

	// Use the fact that JSON is also valid YAML ..
	expected := `
{
  "apiVersion": "v0alpha1",
  "kind": "TestObject",
  "nil": null,
  "bool": true,
  "number": 42,
  "string": "bar",
  "notAVal": "aVal",
  "array": ["one", 2, [3]],
  "object": {"foo": "bar"}
}
`

	for format, parser := range formats {
		expected, err := parser(expected)
		if err != nil {
			t.Errorf("error parsing *expected* value: %s", err)
		}

		output := cmdOutput(t, []string{"show",
			"-J", filepath.FromSlash("../testdata/lib"),
			"-o", format,
			filepath.FromSlash("../testdata/test.jsonnet"),
			"-V", "aVar=aVal",
		})

		t.Log("output is", output)
		actual, err := parser(output)
		if err != nil {
			t.Errorf("error parsing output of format %s: %s", format, err)
		} else if !reflect.DeepEqual(expected, actual) {
			t.Errorf("format %s expected != actual: %s != %s", format, expected, actual)
		}
	}
}
