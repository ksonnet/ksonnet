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

// +build e2e

package e2e

import (
	"bytes"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/util/table"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ks import", func() {
	var (
		a          app
		importPath string
		isErr      bool
		o          *output
	)

	BeforeEach(func() {
		a = e.initApp(nil)
		a.generateDeployedService()
		isErr = false
	})

	JustBeforeEach(func() {
		o = a.runKs("import", "-f", importPath)
		if isErr {
			assertExitStatus(o, 1)
			return
		}

		assertExitStatus(o, 0)
	})

	Context("directory", func() {
		BeforeEach(func() {
			importPath = filepath.Join(e.wd(), "testdata", "input", "import")
		})

		It("imports the files in the directory", func() {
			names := a.componentNames()

			var buf bytes.Buffer
			t := table.New(&buf)
			t.SetHeader([]string{"component"})
			for _, name := range names {
				t.Append([]string{name})
			}
			Expect(t.Render()).NotTo(HaveOccurred())

			o = a.componentList()
			Expect(o.stdout).To(Equal(buf.String()))
		})
	})

	Context("file", func() {
		BeforeEach(func() {
			importPath = filepath.Join(e.wd(), "testdata", "input", "import", "deployment.yaml")
		})

		It("imports the file", func() {
			names := a.componentNames()

			var buf bytes.Buffer
			t := table.New(&buf)
			t.SetHeader([]string{"component"})
			for _, name := range names {
				t.Append([]string{name})
			}
			Expect(t.Render()).NotTo(HaveOccurred())

			o = a.componentList()
			Expect(o.stdout).To(Equal(buf.String()))
		})
	})

	Context("invalid path", func() {
		BeforeEach(func() {
			importPath = filepath.Join(e.wd(), "testdata", "input", "import", "invalid.yaml")
			isErr = true
		})

		It("returns an error", func() {
			assertOutputContains("import/invalid.txt", o.stderr)
		})
	})
})
