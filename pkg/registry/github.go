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
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/parts"
	"github.com/ksonnet/ksonnet/pkg/util/github"
	ksio "github.com/ksonnet/ksonnet/pkg/util/io"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	rawGitHubRoot       = "https://raw.githubusercontent.com"
	defaultGitHubBranch = "master"
)

var (
	// errInvalidURI is an invalid github uri error.
	errInvalidURI = fmt.Errorf("Invalid GitHub URI: try navigating in GitHub to the URI of the folder containing the 'yaml', and using that URI instead. Generally, this URI should be of the form 'github.com/{organization}/{repository}/tree/{branch}/[path-to-directory]'")

	githubFactory = func(a app.App, spec *app.RegistryRefSpec) (*GitHub, error) {
		return NewGitHub(a, spec)
	}
)

type ghFactoryFn func(a app.App, spec *app.RegistryRefSpec) (*GitHub, error)

// GitHubClient is an option for the setting a github client.
func GitHubClient(c github.GitHub) GitHubOpt {
	return func(gh *GitHub) {
		gh.ghClient = c
	}
}

// GitHubOpt is an option for configuring GitHub.
type GitHubOpt func(*GitHub)

// GitHub is a Github Registry
type GitHub struct {
	app      app.App
	name     string
	hd       *hubDescriptor
	ghClient github.GitHub
	spec     *app.RegistryRefSpec
}

// NewGitHub creates an instance of GitHub.
func NewGitHub(a app.App, registryRef *app.RegistryRefSpec, opts ...GitHubOpt) (*GitHub, error) {
	if registryRef == nil {
		return nil, errors.New("registry ref is nil")
	}

	gh := &GitHub{
		app:      a,
		name:     registryRef.Name,
		spec:     registryRef,
		ghClient: github.DefaultClient,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(gh)
	}

	hd, err := parseGitHubURI(gh.URI())
	if err != nil {
		return nil, err
	}
	gh.hd = hd

	if gh.spec.GitVersion == nil || gh.spec.GitVersion.CommitSHA == "" {
		ctx := context.Background()
		sha, err := gh.ghClient.CommitSHA1(ctx, hd.Repo(), hd.refSpec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to find SHA1 for repo")
		}

		gh.spec.GitVersion = &app.GitVersionSpec{
			RefSpec:   hd.refSpec,
			CommitSHA: sha,
		}
	}

	if err = gh.ghClient.ValidateURL(registryRef.URI); err != nil {
		return nil, errors.Wrap(err, "validating GitHub registry URL")
	}

	return gh, nil
}

// IsOverride is true if this registry an an override.
func (gh *GitHub) IsOverride() bool {
	return gh.spec.IsOverride()
}

// Name is the registry name.
func (gh *GitHub) Name() string {
	return gh.name
}

// Protocol is the registry protocol.
func (gh *GitHub) Protocol() Protocol {
	return Protocol(gh.spec.Protocol)
}

// URI is the registry URI.
func (gh *GitHub) URI() string {
	return gh.spec.URI
}

// RegistrySpecDir is the registry directory.
func (gh *GitHub) RegistrySpecDir() string {
	return gh.Name()
}

// RegistrySpecFilePath is the path for the registry.yaml
func (gh *GitHub) RegistrySpecFilePath() string {
	if gh.spec.GitVersion.CommitSHA != "" {
		return path.Join(gh.Name(), gh.spec.GitVersion.CommitSHA+".yaml")
	}
	return path.Join(gh.Name(), gh.spec.GitVersion.RefSpec+".yaml")
}

// FetchRegistrySpec fetches the registry spec (registry.yaml, inventory of packages)
// This inventory may have been previously cached on disk, otherwise it is
// fetched from the remote GitHub repo.
func (gh *GitHub) FetchRegistrySpec() (*Spec, error) {
	log := log.WithField("action", "GitHub.FetchRegistrySpec")

	// Check local disk cache.
	registrySpecFile := registrySpecFilePath(gh.app, gh)

	log.Debugf("checking for registry cache: %v", registrySpecFile)
	registrySpec, exists, err := load(gh.app, registrySpecFile)
	if err != nil {
		log.Debugf("error loading cache for %v (%v), trying to refresh instead", gh.spec.Name, err)
	}

	if !exists {
		log.Debugf("cache not found, fetching remote for %v", gh.spec.Name)
		// If failed, use the protocol to try to retrieve app specification.
		cs := github.ContentSpec{
			Repo:    gh.hd.Repo(),
			Path:    gh.hd.regSpecRepoPath,
			RefSpec: gh.spec.GitVersion.CommitSHA,
		}

		registrySpec, err = gh.fetchRemoteSpec(cs)
		if err != nil {
			return nil, err
		}

		var registrySpecBytes []byte
		registrySpecBytes, err = registrySpec.Marshal()
		if err != nil {
			return nil, err
		}

		// NOTE: We call mkdir after getting the registry spec, since a
		// network call might fail and leave this half-initialized empty
		// directory.
		registrySpecDir := filepath.Join(registryCacheRoot(gh.app), gh.RegistrySpecDir())
		err = gh.app.Fs().MkdirAll(registrySpecDir, app.DefaultFolderPermissions)
		if err != nil {
			return nil, err
		}

		err = afero.WriteFile(gh.app.Fs(), registrySpecFile, registrySpecBytes, app.DefaultFilePermissions)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return registrySpec, nil
}

// fetchRemoteSpec fetche a ksonnet registry spec (registry.yaml) from a remote GitHub repository.
// repo describes the remote repo (org/repo)
// path is the file path within the repo (represents the registry.yaml file)
// sha1 is the commit to pull the contents from
func (gh *GitHub) fetchRemoteSpec(cs github.ContentSpec) (*Spec, error) {
	ctx := context.Background()

	log.Debugf("[fetchRemoteSpec] fetching %v", cs)
	file, _, err := gh.ghClient.Contents(ctx, cs.Repo, cs.Path,
		cs.RefSpec)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("Could not find valid registry with coordinates: %v", cs)
	}

	registrySpecText, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	// Deserialize, return.
	registrySpec, err := Unmarshal([]byte(registrySpecText))
	if err != nil {
		return nil, err
	}

	registrySpec.GitVersion = &app.GitVersionSpec{
		RefSpec:   gh.spec.GitVersion.RefSpec,
		CommitSHA: gh.spec.GitVersion.CommitSHA,
	}

	return registrySpec, nil
}

// MakeRegistryRefSpec returns an app registry ref spec.
func (gh *GitHub) MakeRegistryRefSpec() *app.RegistryRefSpec {
	return gh.spec
}

// ResolveLibrarySpec returns a resolved spec for a part.
func (gh *GitHub) ResolveLibrarySpec(partName, libRefSpec string) (*parts.Spec, error) {
	ctx := context.Background()
	resolvedSHA, err := gh.ghClient.CommitSHA1(ctx, gh.hd.Repo(), libRefSpec)
	if err != nil {
		return nil, err
	}

	// Resolve app spec.
	appSpecPath := strings.Join([]string{gh.hd.regRepoPath, partName, partsYAMLFile}, "/")

	file, directory, err := gh.ghClient.Contents(ctx, gh.hd.Repo(), appSpecPath, resolvedSHA)
	if err != nil {
		return nil, err
	} else if directory != nil {
		return nil, fmt.Errorf("Can't download library specification; resource '%s' points at a file", gh.registrySpecRawURL())
	}

	partsSpecText, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	parts, err := parts.Unmarshal([]byte(partsSpecText))
	if err != nil {
		return nil, err
	}

	return parts, nil
}

// ResolveLibrary fetches the part and creates a parts spec and library ref spec.
func (gh *GitHub) ResolveLibrary(partName, partAlias, libRefSpec string, onFile ResolveFile, onDir ResolveDirectory) (*parts.Spec, *app.LibraryRefSpec, error) {
	defaultRefSpec := "master"

	// Resolve `version` (a git refspec) to a specific SHA.
	ctx := context.Background()
	if len(libRefSpec) == 0 {
		libRefSpec = defaultRefSpec
	}

	resolvedSHA, err := gh.ghClient.CommitSHA1(ctx, gh.hd.Repo(), libRefSpec)
	if err != nil {
		return nil, nil, err
	}

	// Resolve directories and files.
	path := strings.Join([]string{gh.hd.regRepoPath, partName}, "/")
	err = gh.resolveDir(partName, path, resolvedSHA, onFile, onDir)
	if err != nil {
		return nil, nil, err
	}

	// Resolve app spec.
	appSpecPath := strings.Join([]string{path, partsYAMLFile}, "/")
	ctx = context.Background()
	file, directory, err := gh.ghClient.Contents(ctx, gh.hd.Repo(), appSpecPath, resolvedSHA)

	if err != nil {
		return nil, nil, err
	} else if directory != nil {
		return nil, nil, fmt.Errorf("Can't download library specification; resource '%s' points at a file", gh.registrySpecRawURL())
	}

	partsSpecText, err := file.GetContent()
	if err != nil {
		return nil, nil, err
	}

	parts, err := parts.Unmarshal([]byte(partsSpecText))
	if err != nil {
		return nil, nil, err
	}

	if partAlias == "" {
		partAlias = partName
	}

	refSpec := app.LibraryRefSpec{
		Name:     partAlias,
		Registry: gh.Name(),
		GitVersion: &app.GitVersionSpec{
			RefSpec:   libRefSpec,
			CommitSHA: resolvedSHA,
		},
	}

	return parts, &refSpec, nil
}

func (gh *GitHub) resolveDir(libID, path, version string, onFile ResolveFile, onDir ResolveDirectory) error {
	ctx := context.Background()

	file, directory, err := gh.ghClient.Contents(ctx, gh.hd.Repo(), path, version)
	if err != nil {
		return err
	} else if file != nil {
		return fmt.Errorf("Lib ID %q resolves to a file in registry %q", libID, gh.Name())
	}

	for _, item := range directory {
		switch item.GetType() {
		case "file":
			itemPath := item.GetPath()
			file, directory, err := gh.ghClient.Contents(ctx, gh.hd.Repo(), itemPath, version)
			if err != nil {
				return err
			} else if directory != nil {
				return fmt.Errorf("INTERNAL ERROR: GitHub API reported resource %q of type file, but returned type dir", itemPath)
			}
			contents, err := file.GetContent()
			if err != nil {
				return err
			}
			if err := onFile(itemPath, []byte(contents)); err != nil {
				return err
			}
		case "dir":
			itemPath := item.GetPath()
			if err := onDir(itemPath); err != nil {
				return err
			}
			if err := gh.resolveDir(libID, itemPath, version, onFile, onDir); err != nil {
				return err
			}
		case "symlink":
		case "submodule":
			return fmt.Errorf("Invalid library %q; ksonnet doesn't support libraries with symlinks or submodules", libID)
		}
	}

	return nil
}

func (gh *GitHub) registrySpecRawURL() string {
	return strings.Join([]string{
		rawGitHubRoot,
		gh.hd.org,
		gh.hd.repo,
		gh.spec.GitVersion.RefSpec,
		gh.hd.regSpecRepoPath}, "/")
}

type hubDescriptor struct {
	org             string
	repo            string
	refSpec         string
	regRepoPath     string
	regSpecRepoPath string
}

func (hd *hubDescriptor) Repo() github.Repo {
	return github.Repo{Org: hd.org, Repo: hd.repo}
}

// func parseGitHubURI(uri string) (org, repo, refSpec, regRepoPath, regSpecRepoPath string, err error) {
func parseGitHubURI(uri string) (hd *hubDescriptor, err error) {
	// Normalize URI.
	uri = strings.TrimSpace(uri)
	if strings.HasPrefix(uri, "http://github.com") || strings.HasPrefix(uri, "https://github.com") || strings.HasPrefix(uri, "http://www.github.com") || strings.HasPrefix(uri, "https://www.github.com") {
		// Do nothing.
	} else if strings.HasPrefix(uri, "github.com") || strings.HasPrefix(uri, "www.github.com") {
		uri = "http://" + uri
	} else {
		return nil, errors.Errorf("Registries using protocol 'github' must provide URIs beginning with 'github.com' (optionally prefaced with 'http', 'https', 'www', and so on")
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if len(parsed.Query()) != 0 {
		return nil, errors.Errorf("No query strings allowed in registry URI:\n%s", uri)
	}

	components := strings.Split(parsed.Path, "/")
	if len(components) < 3 {
		return nil, errors.Errorf("GitHub URI must point at a repository:\n%s", uri)
	}

	hd = &hubDescriptor{}

	// NOTE: The first component is always blank, because the path
	// begins like: '/whatever'.
	hd.org = components[1]
	hd.repo = components[2]

	//
	// Parse out `regSpecRepoPath`. There are a few cases:
	//   * URI points at a directory inside the respoitory, e.g.,
	//     'http://github.com/ksonnet/parts/tree/master/incubator'
	//   * URI points at an 'app.yaml', e.g.,
	//     'http://github.com/ksonnet/parts/blob/master/yaml'
	//   * URI points at a repository root, e.g.,
	//     'http://github.com/ksonnet/parts'
	//
	if len := len(components); len > 4 {
		hd.refSpec = components[4]

		//
		// Case where we're pointing at either a directory inside a GitHub
		// URL, or an 'app.yaml' inside a GitHub URL.
		//

		// See note above about first component being blank.
		if components[3] == "tree" {
			// If we have a trailing '/' character, last component will be blank. Make
			// sure that `regRepoPath` does not contain a trailing `/`.
			if components[len-1] == "" {
				hd.regRepoPath = strings.Join(components[5:len-1], "/")
				components[len-1] = registryYAMLFile
			} else {
				hd.regRepoPath = strings.Join(components[5:], "/")
				components = append(components, registryYAMLFile)
			}
			hd.regSpecRepoPath = strings.Join(components[5:], "/")
			return
		} else if components[3] == "blob" && components[len-1] == registryYAMLFile {
			hd.regRepoPath = strings.Join(components[5:len-1], "/")
			// Path to the `yaml` (may or may not exist).
			hd.regSpecRepoPath = strings.Join(components[5:], "/")
			return
		} else {
			return nil, errInvalidURI
		}
	} else {
		hd.refSpec = defaultGitHubBranch

		// Else, URI should point at repository root.
		if components[len-1] == "" {
			components[len-1] = defaultGitHubBranch
			components = append(components, registryYAMLFile)
		} else {
			components = append(components, defaultGitHubBranch, registryYAMLFile)
		}

		hd.regRepoPath = ""
		hd.regSpecRepoPath = registryYAMLFile
		return
	}
}

// CacheRoot returns the root for caching.
func (gh *GitHub) CacheRoot(name, path string) (string, error) {
	u, err := url.Parse("https://" + gh.URI())
	if err != nil {
		return "", errors.Errorf("unknown URL: %q", gh.URI())
	}

	var root string

	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	switch {
	case len(parts) == 2:
		// org/repo - use empty root prefix
	case len(parts) > 3:
		// org/repo/tree/branch/path/subpath
		root = strings.Join(parts[4:], "/")
	default:
		return "", errors.Errorf("unknown path %q", u.Path)
	}

	relPath := strings.TrimPrefix(path, "/")
	return filepath.Join(name, strings.TrimPrefix(relPath, root)), nil
}

func (gh *GitHub) fetchRemoteAndSave(cs github.ContentSpec, w io.Writer) error {
	if gh == nil {
		return errors.Errorf("nil receiver")
	}

	if gh.app == nil {
		return errors.Errorf("application is required")
	}

	if w == nil {
		return errors.Errorf("writer is required")
	}

	// If failed, use the protocol to try to retrieve app specification.
	registrySpec, err := gh.fetchRemoteSpec(cs)
	if err != nil {
		return err
	}

	var registrySpecBytes []byte
	registrySpecBytes, err = registrySpec.Marshal()
	if err != nil {
		return err
	}

	r := bytes.NewReader(registrySpecBytes)
	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrap(err, "failed writing registry spec")

	}

	return nil
}

// Update implements registry.Updater
func (gh *GitHub) Update(version string) (string, error) {
	if gh == nil {
		return "", errors.Errorf("nil receiver")
	}

	if version != "" {
		// TODO, see NewGitHub
		return "", errors.Errorf("TODO not implemented")
	}

	// Should we be using gh.spec.GitVersion.RefSpec instead??
	log.Debugf("[github.Update] trying to update registry %v@%v", gh.hd.Repo(), gh.hd.refSpec)
	if gh.spec.GitVersion != nil {
		log.Debugf("[github.Update] current version: %v", gh.spec.GitVersion.CommitSHA)
	}

	// Resolve the latest SHA for the desired git RefSpec
	ctx := context.Background()
	sha, err := gh.ghClient.CommitSHA1(ctx, gh.hd.Repo(), gh.hd.refSpec)
	if err != nil {
		return "", errors.Wrap(err, "unable to find SHA1 for repo")
	}

	// Use the new sha we resolved using the branch refspec
	cs := github.ContentSpec{
		Repo:    gh.hd.Repo(),
		Path:    gh.hd.regSpecRepoPath,
		RefSpec: sha,
	}

	fs := gh.app.Fs()
	registrySpecFile := registrySpecFilePath(gh.app, gh)
	tw, err := ksio.NewTransactionWriter(fs, registrySpecFile)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create writer for file: %v", registrySpecFile)
	}
	if err := gh.fetchRemoteAndSave(cs, tw); err != nil {
		tw.Abort()
		return "", errors.Wrapf(err, "failed to update registry %v", gh.name)
	}

	tw.Commit()
	return sha, nil
}
