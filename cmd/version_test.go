package cmd

import (
	"regexp"
	"testing"
)

func TestVersion(t *testing.T) {
	output := cmdOutput(t, []string{"version"})

	// Also a good smoke-test that libjsonnet linked successfully
	if !regexp.MustCompile(`jsonnet version: v[\d.]+`).MatchString(output) {
		t.Error("Failed to find jsonnet version in:", output)
	}
}
