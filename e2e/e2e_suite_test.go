// +build e2e

package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	e *e2e
)

var _ = BeforeSuite(func() {
	e = newE2e()
	e.buildKs()
})

var _ = AfterSuite(func() {
	e.close()
})

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ksonnet e2e")
}
