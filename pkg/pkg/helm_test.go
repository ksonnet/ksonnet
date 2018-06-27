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

package pkg

import (
	"testing"

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withHelmChart(t *testing.T, fn func(a *amocks.App, fs afero.Fs)) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		a.On("VendorPath").Return("/app/vendor")

		test.StageDir(t, fs, "redis", "/app/vendor/helm-stable/redis")

		fn(a, fs)
	})
}

func TestHelm_find_latest_chart_when_no_version(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		h, err := NewHelm(a, "redis", "helm-stable", "", nil)
		require.NoError(t, err)

		require.Equal(t, "redis", h.Name())
	})
}

func TestHelm_error_when_version_does_not_exist(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		_, err := NewHelm(a, "redis", "helm-stable", "3.3.1", nil)
		require.Error(t, err)
	})
}

func TestHelm_error_when_chart_has_no_versions(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		test.StageDir(t, fs, "noversion", "/app/vendor/helm-stable/noversion")
		_, err := NewHelm(a, "noversion", "helm-stable", "", nil)
		require.Error(t, err)
	})
}

func TestHelm_Name(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		h, err := NewHelm(a, "redis", "helm-stable", "3.3.6", nil)
		require.NoError(t, err)

		require.Equal(t, "redis", h.Name())
	})
}

func TestHelm_RegistryName(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		h, err := NewHelm(a, "redis", "helm-stable", "3.3.6", nil)
		require.NoError(t, err)

		require.Equal(t, "helm-stable", h.RegistryName())
	})
}

func TestHelm_IsInstalled(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		ic := &fakeInstallChecker{
			isInstalled: true,
		}

		h, err := NewHelm(a, "redis", "helm-stable", "3.3.6", ic)
		require.NoError(t, err)

		i, err := h.IsInstalled()
		assert.NoError(t, err)
		assert.True(t, i)
	})
}

func TestHelm_Description(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		h, err := NewHelm(a, "redis", "helm-stable", "3.3.6", nil)
		require.NoError(t, err)

		require.Equal(t, "redis chart", h.Description())
	})
}

func TestHelm_Prototypes(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		h, err := NewHelm(a, "redis", "helm-stable", "3.3.6", nil)
		require.NoError(t, err)

		prototypes, err := h.Prototypes()
		require.NoError(t, err)

		require.Len(t, prototypes, 1)
		proto := prototypes[0]
		require.Equal(t, "io.ksonnet.pkg.helm-stable-redis", proto.Name)
	})
}

func TestHelm_Path(t *testing.T) {
	withHelmChart(t, func(a *amocks.App, fs afero.Fs) {
		h, err := NewHelm(a, "redis", "helm-stable", "3.3.6", nil)
		require.NoError(t, err)

		got := h.Path()

		expected := "/app/vendor/helm-stable/redis/helm/3.3.6/redis"
		require.Equal(t, expected, got)
	})

}
