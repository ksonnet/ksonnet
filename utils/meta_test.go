package utils

import (
	"testing"

	"k8s.io/apimachinery/pkg/version"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    version.Info
		expected ServerVersion
		error    bool
	}{
		{
			input:    version.Info{Major: "1", Minor: "6"},
			expected: ServerVersion{Major: 1, Minor: 6},
		},
		{
			input:    version.Info{Major: "1", Minor: "70"},
			expected: ServerVersion{Major: 1, Minor: 70},
		},
		{
			input: version.Info{Major: "1", Minor: "6x"},
			error: true,
		},
	}

	for _, test := range tests {
		v, err := ParseVersion(&test.input)
		if test.error {
			if err == nil {
				t.Errorf("test %s should have failed and did not", test.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("test %v failed: %v", test.input, err)
			continue
		}
		if v != test.expected {
			t.Errorf("Expected %v, got %v", test.expected, v)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	v := ServerVersion{Major: 2, Minor: 3}
	tests := []struct {
		major, minor, result int
	}{
		{major: 1, minor: 0, result: 1},
		{major: 2, minor: 0, result: 1},
		{major: 2, minor: 2, result: 1},
		{major: 2, minor: 3, result: 0},
		{major: 2, minor: 4, result: -1},
		{major: 3, minor: 0, result: -1},
	}
	for _, test := range tests {
		res := v.Compare(test.major, test.minor)
		if res != test.result {
			t.Errorf("%d.%d => Expected %d, got %d", test.major, test.minor, test.result, res)
		}
	}
}
