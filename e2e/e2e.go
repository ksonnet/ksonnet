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

package e2e

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	// test helpers
	. "github.com/onsi/gomega"
)

type e2e struct {
	root string
}

func newE2e() *e2e {
	dir, err := ioutil.TempDir("", "")
	Expect(err).ToNot(HaveOccurred())

	e := &e2e{
		root: dir,
	}

	return e
}

func (e *e2e) close() {
	err := os.RemoveAll(e.root)
	Expect(err).ToNot(HaveOccurred())
}

func (e *e2e) ksBin() string {
	return filepath.Join(e.root, "ks")
}

func (e *e2e) ks(args ...string) *output {
	cmd := exec.Command(e.ksBin(), args...)
	o, err := runWithOutput(cmd)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return o
}

func (e *e2e) ksInApp(appDir string, args ...string) *output {
	ExpectWithOffset(1, appDir).To(BeADirectory())
	cmd := exec.Command(e.ksBin(), args...)
	cmd.Dir = appDir
	o, err := runWithOutput(cmd)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	return o
}

func (e *e2e) buildKs() {
	args := []string{
		"build",
		"-o",
		e.ksBin(),
		`github.com/ksonnet/ksonnet`,
	}

	cmd := exec.Command("go", args...)

	o, err := runWithOutput(cmd)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	assertExitStatus(o, 0)
}

func (e *e2e) initApp(server string) app {
	if server == "" {
		server = "http://example.com"
	}

	appID := randString(6)
	appDir := filepath.Join(e.root, appID)
	options := []string{
		"init",
		appID,
		"--dir",
		appDir,
		"--server",
		server,
	}

	o := e.ks(options...)
	assertExitStatus(o, 0)
	return app{dir: appDir, e2e: e}
}
