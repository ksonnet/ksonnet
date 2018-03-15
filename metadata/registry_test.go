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

package metadata

import (
	"path"
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/parts"
	"github.com/ksonnet/ksonnet/metadata/registry"
)

func TestParseGiHubRegistryURITest(t *testing.T) {
	tests := []struct {
		// Specification to parse.
		uri string

		// Optional error to check.
		targetErr error

		// Optional results to verify.
		targetOrg                  string
		targetRepo                 string
		targetRefSpec              string
		targetRegistryRepoPath     string
		targetRegistrySpecRepoPath string
	}{
		//
		// `parseGitHubURI` should correctly parse org, repo, and refspec. Does not
		// test path parsing.
		//
		{
			uri: "github.com/exampleOrg1/exampleRepo1",

			targetOrg:                  "exampleOrg1",
			targetRepo:                 "exampleRepo1",
			targetRefSpec:              "master",
			targetRegistryRepoPath:     "",
			targetRegistrySpecRepoPath: "registry.yaml",
		},
		{
			uri: "github.com/exampleOrg2/exampleRepo2/tree/master",

			targetOrg:                  "exampleOrg2",
			targetRepo:                 "exampleRepo2",
			targetRefSpec:              "master",
			targetRegistryRepoPath:     "",
			targetRegistrySpecRepoPath: "registry.yaml",
		},
		{
			uri: "github.com/exampleOrg3/exampleRepo3/tree/exampleBranch1",

			targetOrg:                  "exampleOrg3",
			targetRepo:                 "exampleRepo3",
			targetRefSpec:              "exampleBranch1",
			targetRegistryRepoPath:     "",
			targetRegistrySpecRepoPath: "registry.yaml",
		},
		{
			// Fails because `blob` refers to a file, but this refers to a directory.
			uri:       "github.com/exampleOrg4/exampleRepo4/blob/master",
			targetErr: errInvalidURI,
		},
		{
			uri: "github.com/exampleOrg4/exampleRepo4/tree/exampleBranch2",

			targetOrg:                  "exampleOrg4",
			targetRepo:                 "exampleRepo4",
			targetRefSpec:              "exampleBranch2",
			targetRegistryRepoPath:     "",
			targetRegistrySpecRepoPath: "registry.yaml",
		},

		//
		// Parsing URIs with paths.
		//
		{
			// Fails because referring to a directory requires a URI with
			// `tree/{branchName}` prepending the path.
			uri:       "github.com/exampleOrg6/exampleRepo6/path/to/some/registry",
			targetErr: errInvalidURI,
		},
		{
			uri: "github.com/exampleOrg5/exampleRepo5/tree/master/path/to/some/registry",

			targetOrg:                  "exampleOrg5",
			targetRepo:                 "exampleRepo5",
			targetRefSpec:              "master",
			targetRegistryRepoPath:     "path/to/some/registry",
			targetRegistrySpecRepoPath: "path/to/some/registry/registry.yaml",
		},
		{
			uri: "github.com/exampleOrg6/exampleRepo6/tree/exampleBranch3/path/to/some/registry",

			targetOrg:                  "exampleOrg6",
			targetRepo:                 "exampleRepo6",
			targetRefSpec:              "exampleBranch3",
			targetRegistryRepoPath:     "path/to/some/registry",
			targetRegistrySpecRepoPath: "path/to/some/registry/registry.yaml",
		},
		{
			// Fails because `blob` refers to a file, but this refers to a directory.
			uri:       "github.com/exampleOrg7/exampleRepo7/blob/master",
			targetErr: errInvalidURI,
		},
		{
			// Fails because `blob` refers to a file, but this refers to a directory.
			uri:       "github.com/exampleOrg5/exampleRepo5/blob/exampleBranch2",
			targetErr: errInvalidURI,
		},
	}

	for _, test := range tests {
		// Make sure we correctly parse each URN as a bare-domain URI, as well as
		// with 'http://' and 'https://' as prefixes.
		for _, prefix := range []string{"http://", "https://", "http://www.", "https://www.", "www.", ""} {
			// Make sure we correctly parse each URI even if it has the optional
			// trailing `/` character.
			for _, suffix := range []string{"/", ""} {
				uri := prefix + test.uri + suffix

				t.Run(uri, func(t *testing.T) {
					org, repo, refspec, registryRepoPath, registrySpecRepoPath, err := parseGitHubURI(uri)
					if test.targetErr != nil {
						if err != test.targetErr {
							t.Fatalf("Expected URI '%s' parse to fail with err '%v', got: '%v'", uri, test.targetErr, err)
						}
						return
					}

					if err != nil {
						t.Fatalf("Expected parse to succeed, but failed with error '%v'", err)
					}

					if org != test.targetOrg {
						t.Errorf("Expected org '%s', got '%s'", test.targetOrg, org)
					}

					if repo != test.targetRepo {
						t.Errorf("Expected repo '%s', got '%s'", test.targetRepo, repo)
					}

					if refspec != test.targetRefSpec {
						t.Errorf("Expected refspec '%s', got '%s'", test.targetRefSpec, refspec)
					}

					if registryRepoPath != test.targetRegistryRepoPath {
						t.Errorf("Expected registryRepoPath '%s', got '%s'", test.targetRegistryRepoPath, registryRepoPath)
					}

					if registrySpecRepoPath != test.targetRegistrySpecRepoPath {
						t.Errorf("Expected targetRegistrySpecRepoPath '%s', got '%s'", test.targetRegistrySpecRepoPath, registrySpecRepoPath)
					}
				})
			}
		}
	}
}

//
// Mock registry manager for end-to-end tests.
//

type mockRegistryManager struct {
	*app.RegistryRefSpec
	registryDir string
}

func newMockRegistryManager(name string) *mockRegistryManager {
	return &mockRegistryManager{
		registryDir: name,
		RegistryRefSpec: &app.RegistryRefSpec{
			Name: name,
		},
	}
}

func (m *mockRegistryManager) ResolveLibrarySpec(libID, libRefSpec string) (*parts.Spec, error) {
	return nil, nil
}

func (m *mockRegistryManager) RegistrySpecDir() string {
	return m.registryDir
}

func (m *mockRegistryManager) RegistrySpecFilePath() string {
	return path.Join(m.registryDir, "master.yaml")
}

func (m *mockRegistryManager) FetchRegistrySpec() (*registry.Spec, error) {
	registrySpec := registry.Spec{
		APIVersion: registry.DefaultAPIVersion,
		Kind:       registry.DefaultKind,
	}

	return &registrySpec, nil
}

func (m *mockRegistryManager) MakeRegistryRefSpec() *app.RegistryRefSpec {
	return m.RegistryRefSpec
}

func (m *mockRegistryManager) ResolveLibrary(libID, libAlias, version string, onFile registry.ResolveFile, onDir registry.ResolveDirectory) (*parts.Spec, *app.LibraryRefSpec, error) {
	return nil, nil, nil
}
