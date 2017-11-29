package cmd

import "testing"

func Test_genKsRoot(t *testing.T) {
	cases := []struct {
		name     string
		appName  string
		ksDir    string
		wd       string
		expected string
		isErr    bool
	}{
		{name: "no wd", appName: "app", ksDir: "/root", expected: "/root/app"},
		{name: "with abs wd", appName: "app", ksDir: "/root", wd: "/custom", expected: "/custom"},
		{name: "with rel wd #1", appName: "app", ksDir: "/root", wd: "./custom", expected: "/root/custom"},
		{name: "with rel wd #2", appName: "app", ksDir: "/root", wd: "custom", expected: "/root/custom"},
		{name: "with rel wd #2", appName: "app", ksDir: "/root", wd: "../custom", expected: "/custom"},
		{name: "missing ksDir", appName: "app", wd: "./custom", isErr: true},
		{name: "missing appName and wd", ksDir: "/root", isErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := genKsRoot(tc.appName, tc.ksDir, tc.wd)
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
