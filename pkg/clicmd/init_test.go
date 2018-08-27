// Copyright 2018 The kubecfg authors
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

package clicmd

import (
	"testing"

	"github.com/ksonnet/ksonnet/pkg/actions"
)

func Test_initCmd(t *testing.T) {
	cases := []cmdTestCase{
		{
			name: "in general",
			args: []string{"init", "app",
				"--namespace", "new-namespace",
				"--server", "http://127.0.0.1",
				"--env", "env-name",
				"--api-spec", "version:v1.8.0",
			},
			action: actionInit,
			expected: map[string]interface{}{
				actions.OptionFs:                    nil,
				actions.OptionName:                  "app",
				actions.OptionEnvName:               "env-name",
				actions.OptionNewRoot:               "/app",
				actions.OptionServer:                "http://127.0.0.1",
				actions.OptionSpecFlag:              "version:v1.8.0",
				actions.OptionNamespace:             "new-namespace",
				actions.OptionSkipDefaultRegistries: false,
				actions.OptionTLSSkipVerify:         false,
				actions.OptionSkipCheckUpgrade:      true,
			},
		},
		{
			name: "verbose flag before command",
			args: []string{"--verbose=3", "init", "app",
				"--namespace", "new-namespace",
				"--server", "http://127.0.0.1",
				"--env", "env-name",
				"--api-spec", "version:v1.8.0",
			},
			action: actionInit,
			expected: map[string]interface{}{
				actions.OptionFs:                    nil,
				actions.OptionName:                  "app",
				actions.OptionEnvName:               "env-name",
				actions.OptionNewRoot:               "/app",
				actions.OptionServer:                "http://127.0.0.1",
				actions.OptionSpecFlag:              "version:v1.8.0",
				actions.OptionNamespace:             "new-namespace",
				actions.OptionSkipDefaultRegistries: false,
				actions.OptionTLSSkipVerify:         false,
				actions.OptionSkipCheckUpgrade:      true,
			},
		},
		{
			name: "global dir flag",
			args: []string{"init", "app",
				"--namespace", "new-namespace",
				"--server", "http://127.0.0.1",
				"--env", "env-name",
				"--api-spec", "version:v1.8.0",
				"--dir", "/app/custom",
			},
			action: actionInit,
			expected: map[string]interface{}{
				actions.OptionFs:                    nil,
				actions.OptionName:                  "app",
				actions.OptionEnvName:               "env-name",
				actions.OptionNewRoot:               "/app/custom",
				actions.OptionServer:                "http://127.0.0.1",
				actions.OptionSpecFlag:              "version:v1.8.0",
				actions.OptionNamespace:             "new-namespace",
				actions.OptionSkipDefaultRegistries: false,
				actions.OptionTLSSkipVerify:         false,
				actions.OptionSkipCheckUpgrade:      true,
			},
		},
		{
			name:  "no args",
			args:  []string{"init"},
			isErr: true,
		},
	}

	runTestCmd(t, cases)
}

func Test_genKsRoot(t *testing.T) {
	cases := []struct {
		name      string
		appName   string
		ksExecDir string
		newAppDir string
		expected  string
		isErr     bool
	}{
		{name: "no newAppDir", appName: "app", ksExecDir: "/root", expected: "/root/app"},
		{name: "with abs newAppDir", appName: "app", ksExecDir: "/root", newAppDir: "/custom", expected: "/custom"},
		{name: "with rel newAppDir #1", appName: "app", ksExecDir: "/root", newAppDir: "./custom", expected: "/root/custom"},
		{name: "with rel newAppDir #2", appName: "app", ksExecDir: "/root", newAppDir: "custom", expected: "/root/custom"},
		{name: "with rel newAppDir #2", appName: "app", ksExecDir: "/root", newAppDir: "../custom", expected: "/custom"},
		{name: "missing ksExecDir with rel newAppDir", appName: "app", newAppDir: "./custom", isErr: true},
		{name: "missing ksExecDir with abs newAppDir", appName: "app", newAppDir: "/custom", expected: "/custom"},
		{name: "missing appname and ksExecDir", newAppDir: "/custom", expected: "/custom"},
		{name: "missing appname and ksExecDir rel newAppDir", newAppDir: "./custom", isErr: true},
		{name: "missing appName and newAppDir", ksExecDir: "/root", isErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := genKsRoot(tc.appName, tc.ksExecDir, tc.newAppDir)
			if tc.isErr {
				if err == nil {
					t.Errorf("genKsRoot expected error, but none was received")
				}
			} else {
				if got != tc.expected {
					t.Errorf("genKsRoot got %q; expected %q", got, tc.expected)
				}
			}
		})
	}
}
