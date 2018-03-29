// +build e2e

package e2e

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	extv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ks apply", func() {
	var a app
	var namespace string

	BeforeEach(func() {
		namespace = e.createNamespace()

		io := &initOptions{
			context:   "minikube",
			namespace: namespace,
		}

		a = e.initApp(io)
		a.generateDeployedService()
	})

	AfterEach(func() {
		e.removeNamespace(namespace)
	})

	JustBeforeEach(func() {
		o := a.runKs("apply", namespace)
		assertExitStatus(o, 0)
	})

	It("creates a guestbook-ui service", func() {
		c, err := corev1.NewForConfig(e.restConfig)
		Expect(err).NotTo(HaveOccurred())

		_, err = c.Services(namespace).Get("guestbook-ui", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates a guestbook-ui deployment", func() {
		c, err := extv1beta1.NewForConfig(e.restConfig)
		Expect(err).NotTo(HaveOccurred())

		_, err = c.Deployments(namespace).Get("guestbook-ui", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
})
