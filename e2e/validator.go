// Copyright 2018 The ksonnet authors
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	extv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"

	. "github.com/onsi/gomega"
)

type validator struct {
	config    *rest.Config
	namespace string
}

func newValidator(config *rest.Config, namespace string) *validator {
	return &validator{
		config:    config,
		namespace: namespace,
	}
}

func (v *validator) hasService(name string) {
	c, err := corev1.NewForConfig(v.config)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	_, err = c.Services(v.namespace).Get(name, metav1.GetOptions{})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}

func (v *validator) hasDeployment(name string) {
	c, err := extv1beta1.NewForConfig(v.config)
	Expect(err).NotTo(HaveOccurred())

	_, err = c.Deployments(v.namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
}
