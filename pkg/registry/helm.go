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

package registry

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/helm"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/util/archive"
	ksstrings "github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
)

var (
	// helmFactory creates Helm registry instances.
	helmFactory = func(a app.App, registryConfig *app.RegistryConfig, httpClient *helm.HTTPClient) (*Helm, error) {
		return NewHelm(a, registryConfig, httpClient, nil)
	}
)

type helmFactoryFn func(a app.App, registryConfig *app.RegistryConfig, httpClient *helm.HTTPClient) (*Helm, error)

// Helm is a Helm repository.
type Helm struct {
	app              app.App
	spec             *app.RegistryConfig
	repositoryClient helm.RepositoryClient
	unarchiver       archive.Unarchiver
}

// NewHelm creates an instance of Helm.
func NewHelm(a app.App, registryRef *app.RegistryConfig, rc helm.RepositoryClient, ua archive.Unarchiver) (*Helm, error) {
	if ua == nil {
		ua = &archive.Tgz{}
	}

	if rc == nil {
		return nil, errors.New("helm repository client is nil")
	}

	h := &Helm{
		app:              a,
		spec:             registryRef,
		repositoryClient: rc,
		unarchiver:       ua,
	}

	return h, nil
}

// RegistrySpecDir is the registry directory.
func (h *Helm) RegistrySpecDir() string {
	return h.Name()
}

// RegistrySpecFilePath is the path for the registry.yaml
// NOTE: this function appears to be github registry specific and may not
// need to be a part of the interface.
func (h *Helm) RegistrySpecFilePath() string {
	return ""
}

// FetchRegistrySpec fetches the registry spec. This method returns an unmarshalled version
// of registry.yaml
func (h *Helm) FetchRegistrySpec() (*Spec, error) {
	spec := &Spec{
		APIVersion: DefaultAPIVersion,
		Kind:       DefaultKind,
		Libraries:  LibraryConfigs{},
	}

	repository, err := h.repositoryClient.Repository()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving repository")
	}

	for _, chart := range repository.Latest() {
		name := chart.Name
		if name == "" {
			return nil, errors.Errorf("entries are invalid")
		}

		spec.Libraries[name] = &LibaryConfig{
			Path:    name,
			Version: chart.Version,
		}
	}

	return spec, nil
}

// MakeRegistryConfig returns app registry ref spec.
func (h *Helm) MakeRegistryConfig() *app.RegistryConfig {
	return h.spec
}

// ResolveLibrarySpec returns a resolved spec for a part.
func (h *Helm) ResolveLibrarySpec(partName, version string) (*parts.Spec, error) {
	chart, err := h.repositoryClient.Chart(partName, version)
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving chart %s-%s", partName, version)
	}

	part := &parts.Spec{
		APIVersion: parts.DefaultAPIVersion,
		Kind:       parts.DefaultKind,

		Name:        chart.Name,
		Version:     chart.Version,
		Description: chart.Description,
	}

	return part, nil
}

// ResolveLibrary fetches the part and creates a parts spec and library ref spec.
func (h *Helm) ResolveLibrary(partName string, partAlias string, version string, onFile ResolveFile, onDir ResolveDirectory) (*parts.Spec, *app.LibraryConfig, error) {
	part, err := h.ResolveLibrarySpec(partName, version)
	if err != nil {
		return nil, nil, err
	}

	chart, err := h.repositoryClient.Chart(partName, version)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "retrieving chart %s-%s", partName, version)
	}

	for _, u := range chart.URLs {
		r, err := h.repositoryClient.Fetch(u)
		if r != nil {
			defer r.Close()
		}

		if err != nil {
			return nil, nil, err
		}

		handler := func(f *archive.File) error {
			var b []byte
			b, err = ioutil.ReadAll(f.Reader)
			if err != nil {
				return err
			}

			name := path.Join(chart.Name, "helm", chart.Version, f.Name)

			if containsHelmHook(name, b) {
				// skip this file because it has a helm hook in it
				return nil
			}

			return onFile(name, b)
		}

		if err = h.unarchiver.Unarchive(r, handler); err != nil {
			return nil, nil, err
		}
	}

	refSpec := &app.LibraryConfig{
		Name:     partAlias,
		Registry: h.Name(),
	}

	return part, refSpec, nil

}

// Name is the registry name.
func (h *Helm) Name() string {
	return h.spec.Name
}

// Protocol is the registry protocol.
func (h *Helm) Protocol() Protocol {
	return ProtocolHelm
}

// URI is the registry URI.
func (h *Helm) URI() string {
	return h.spec.URI
}

// IsOverride is true if this registry is an override.
func (h *Helm) IsOverride() bool {
	return h.spec.IsOverride()
}

// CacheRoot returns the root for caching by combining the path with the registry
// name.
func (h *Helm) CacheRoot(name string, relPath string) (string, error) {
	return filepath.Join(name, relPath), nil
}

// Update implements registry.Updater
func (h *Helm) Update(version string) (string, error) {
	return "", errors.Errorf("TODO not implemented")
}

// ValidateURI implements registry.Validator. A URI is valid if:
//   * It is a valid URI (RFC 3986)
//   * It is an absolute URI or an absolute path
func (h *Helm) ValidateURI(uri string) (bool, error) {
	if h == nil {
		return false, errors.Errorf("nil receiver")
	}

	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return false, err
	}

	return true, nil
}

// SetURI implements registry.Setter. It sets the URI for the registry.
func (h *Helm) SetURI(uri string) error {
	if h == nil {
		return errors.Errorf("nil receiver")
	}
	if h.spec == nil {
		return errors.Errorf("nil spec")
	}

	// Validate
	if ok, err := h.ValidateURI(uri); err != nil || !ok {
		return errors.Wrap(err, "validating uri")
	}

	h.spec.URI = uri
	return nil
}

// containsHelmHook checks file contents for helm hooks. Helm hooks are denoted by
// `helm.sh/hook`.
func containsHelmHook(name string, b []byte) bool {
	dir := filepath.Dir(name)
	ext := filepath.Ext(name)
	if strings.Contains(dir, "templates") && ksstrings.InSlice(ext, []string{".yaml", ".yml"}) {
		return bytes.Contains(b, []byte("helm.sh/hook"))
	}

	return false
}
