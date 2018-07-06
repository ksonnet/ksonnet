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

package cluster

import (
	"testing"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type conflictError struct{}

var _ kerrors.APIStatus = (*conflictError)(nil)
var _ error = (*notFoundError)(nil)

func (e *conflictError) Status() metav1.Status {
	return metav1.Status{
		Reason: metav1.StatusReasonConflict,
	}
}

func (e *conflictError) Error() string {
	return "conflict"
}

func Test_Apply(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		applyConfig := ApplyConfig{
			App:          a,
			ClientConfig: &client.Config{},
		}

		setupApp := func(apply *Apply) {
			obj := &unstructured.Unstructured{Object: genObject()}

			apply.clientOpts = &clientOpts{}

			apply.findObjectsFn = func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
				objects := []*unstructured.Unstructured{obj}

				return objects, nil
			}

			apply.ksonnetObjectFactory = func() ksonnetObject {
				return &fakeKsonnetObject{
					obj: obj,
				}
			}

			apply.upserterFactory = func() Upserter {
				return &fakeUpserter{
					upsertID: "12345",
				}
			}
		}

		err := RunApply(applyConfig, setupApp)
		require.NoError(t, err)
	})
}

func Test_Apply_dry_run(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		applyConfig := ApplyConfig{
			App:          a,
			ClientConfig: &client.Config{},
			DryRun:       true,
		}

		setupApp := func(apply *Apply) {
			obj := &unstructured.Unstructured{Object: genObject()}

			apply.clientOpts = &clientOpts{}

			apply.findObjectsFn = func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
				objects := []*unstructured.Unstructured{obj}

				return objects, nil
			}

			apply.ksonnetObjectFactory = func() ksonnetObject {
				return &fakeKsonnetObject{
					obj: obj,
				}
			}

			apply.upserterFactory = func() Upserter {
				return &fakeUpserter{
					upsertErr: errors.New("upsert should not run"),
				}
			}
		}

		err := RunApply(applyConfig, setupApp)
		require.NoError(t, err)
	})
}

func Test_Apply_retry_on_conflict(t *testing.T) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		applyConfig := ApplyConfig{
			App:          a,
			ClientConfig: &client.Config{},
		}

		setupApp := func(apply *Apply) {
			obj := &unstructured.Unstructured{Object: genObject()}

			apply.clientOpts = &clientOpts{}

			apply.findObjectsFn = func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
				objects := []*unstructured.Unstructured{obj}

				return objects, nil
			}

			apply.ksonnetObjectFactory = func() ksonnetObject {
				return &fakeKsonnetObject{
					obj: obj,
				}
			}

			apply.upserterFactory = func() Upserter {
				return &fakeUpserter{
					upsertErr: &conflictError{},
				}
			}

			apply.conflictTimeout = 0
		}

		err := RunApply(applyConfig, setupApp)
		cause := errors.Cause(err)
		require.Equal(t, errApplyConflict, cause)
	})
}

func genObject() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "apps/v1beta1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "guiroot",
			"annotations": map[string]interface{}{
				"ksonnet.io/dummy": "dummy",
			},
		},
		"spec": map[string]interface{}{
			"replicas": 1,
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "guiroot",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "gcr.io/heptio-images/ks-guestbook-demo:0.1",
							"name":  "guiroot",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 80,
								},
							},
						},
					},
				},
			},
		},
	}
}
