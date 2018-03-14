package e2e

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/gomega"
)

func assertFileExists(path string) {
	_, err := os.Stat(path)
	if err != nil {
		ExpectWithOffset(1, err).To(Not(HaveOccurred()))
	}
}

func assertOutput(name, output string) {
	path := filepath.Join("testdata", "output", name)
	ExpectWithOffset(1, path).To(BeAnExistingFile())

	b, err := ioutil.ReadFile(path)
	ExpectWithOffset(1, err).To(Not(HaveOccurred()))

	ExpectWithOffset(1, output).To(Equal(string(b)),
		"expected output to be:\n%s\nit was:\n%s\n",
		string(b), output)

}

func assertExitStatus(o *output, status int) {
	ExpectWithOffset(1, o.exitCode).To(Equal(status),
		"expected exit status to be %d but was %d\nstdout:\n%s\nstderr:\n%s\nargs:%s\npath:%s",
		status, o.exitCode, o.stdout, o.stderr, strings.Join(o.args, " "), o.cmdName)
}

func assertContents(name, path string) {
	expectedPath := filepath.Join("testdata", "output", name)
	ExpectWithOffset(1, expectedPath).To(BeAnExistingFile())
	ExpectWithOffset(1, path).To(BeAnExistingFile())

	b, err := ioutil.ReadFile(expectedPath)
	ExpectWithOffset(1, err).To(Not(HaveOccurred()))
	expected := string(b)

	b, err = ioutil.ReadFile(path)
	ExpectWithOffset(1, err).To(Not(HaveOccurred()))
	got := string(b)

	ExpectWithOffset(1, expected).To(Equal(got),
		"expected output to be:\n%s\nit was:\n%s\n",
		expected, got)
}
