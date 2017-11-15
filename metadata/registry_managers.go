package metadata

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/parts"
	"github.com/ksonnet/ksonnet/metadata/registry"
)

const (
	rawGitHubRoot       = "https://raw.githubusercontent.com"
	defaultGitHubBranch = "master"

	uriField         = "uri"
	refSpecField     = "refSpec"
	resolvedSHAField = "resolvedSHA"
)

//
// GitHub registry manager.
//

type gitHubRegistryManager struct {
	*app.RegistryRefSpec
	registryDir          string
	org                  string
	repo                 string
	registryRepoPath     string
	registrySpecRepoPath string
}

func makeGitHubRegistryManager(registryRef *app.RegistryRefSpec) (*gitHubRegistryManager, error) {
	gh := gitHubRegistryManager{RegistryRefSpec: registryRef}

	var err error

	// Set registry path.
	gh.registryDir = gh.Name

	// Parse GitHub URI.
	var refspec string
	gh.org, gh.repo, refspec, gh.registryRepoPath, gh.registrySpecRepoPath, err = parseGitHubURI(gh.URI)
	if err != nil {
		return nil, err
	}

	// Resolve the refspec to a commit SHA.
	client := github.NewClient(nil)
	ctx := context.Background()

	sha, _, err := client.Repositories.GetCommitSHA1(ctx, gh.org, gh.repo, refspec, "")
	if err != nil {
		return nil, err
	}
	gh.GitVersion = &app.GitVersionSpec{
		RefSpec:   refspec,
		CommitSHA: sha,
	}

	return &gh, nil
}

func (gh *gitHubRegistryManager) RegistrySpecDir() string {
	return gh.registryDir
}

func (gh *gitHubRegistryManager) RegistrySpecFilePath() string {
	if gh.GitVersion.CommitSHA != "" {
		return path.Join(gh.registryDir, gh.GitVersion.CommitSHA+".yaml")
	}
	return path.Join(gh.registryDir, gh.GitVersion.RefSpec+".yaml")
}

func (gh *gitHubRegistryManager) FetchRegistrySpec() (*registry.Spec, error) {
	// Fetch app spec at specific commit.
	client := github.NewClient(nil)
	ctx := context.Background()

	// Get contents.
	getOpts := github.RepositoryContentGetOptions{Ref: gh.GitVersion.CommitSHA}
	file, _, _, err := client.Repositories.GetContents(ctx, gh.org, gh.repo, gh.registrySpecRepoPath, &getOpts)
	if file == nil {
		return nil, fmt.Errorf("Could not find valid registry at uri '%s/%s/%s' and refspec '%s' (resolves to sha '%s')", gh.org, gh.repo, gh.registrySpecRepoPath, gh.GitVersion.RefSpec, gh.GitVersion.CommitSHA)
	} else if err != nil {
		return nil, err
	}

	registrySpecText, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	// Deserialize, return.
	registrySpec := registry.Spec{}
	err = yaml.Unmarshal([]byte(registrySpecText), &registrySpec)
	if err != nil {
		return nil, err
	}

	registrySpec.GitVersion = &app.GitVersionSpec{
		RefSpec:   gh.GitVersion.RefSpec,
		CommitSHA: gh.GitVersion.CommitSHA,
	}

	return &registrySpec, nil
}

func (gh *gitHubRegistryManager) MakeRegistryRefSpec() *app.RegistryRefSpec {
	return gh.RegistryRefSpec
}

func (gh *gitHubRegistryManager) ResolveLibrarySpec(libID, libRefSpec string) (*parts.Spec, error) {
	client := github.NewClient(nil)

	// Resolve `version` (a git refspec) to a specific SHA.
	ctx := context.Background()
	resolvedSHA, _, err := client.Repositories.GetCommitSHA1(ctx, gh.org, gh.repo, libRefSpec, "")
	if err != nil {
		return nil, err
	}

	// Resolve app spec.
	appSpecPath := strings.Join([]string{gh.registryRepoPath, libID, partsYAMLFile}, "/")
	ctx = context.Background()
	getOpts := &github.RepositoryContentGetOptions{Ref: resolvedSHA}
	file, directory, _, err := client.Repositories.GetContents(ctx, gh.org, gh.repo, appSpecPath, getOpts)
	if err != nil {
		return nil, err
	} else if directory != nil {
		return nil, fmt.Errorf("Can't download library specification; resource '%s' points at a file", gh.registrySpecRawURL())
	}

	partsSpecText, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	parts := parts.Spec{}
	err = yaml.Unmarshal([]byte(partsSpecText), &parts)
	if err != nil {
		return nil, err
	}

	return &parts, nil
}

func (gh *gitHubRegistryManager) ResolveLibrary(libID, libAlias, libRefSpec string, onFile registry.ResolveFile, onDir registry.ResolveDirectory) (*parts.Spec, *app.LibraryRefSpec, error) {
	client := github.NewClient(nil)

	// Resolve `version` (a git refspec) to a specific SHA.
	ctx := context.Background()
	resolvedSHA, _, err := client.Repositories.GetCommitSHA1(ctx, gh.org, gh.repo, libRefSpec, "")
	if err != nil {
		return nil, nil, err
	}

	// Resolve directories and files.
	path := strings.Join([]string{gh.registryRepoPath, libID}, "/")
	err = gh.resolveDir(client, libID, path, resolvedSHA, onFile, onDir)
	if err != nil {
		return nil, nil, err
	}

	// Resolve app spec.
	appSpecPath := strings.Join([]string{path, partsYAMLFile}, "/")
	ctx = context.Background()
	getOpts := &github.RepositoryContentGetOptions{Ref: resolvedSHA}
	file, directory, _, err := client.Repositories.GetContents(ctx, gh.org, gh.repo, appSpecPath, getOpts)
	if err != nil {
		return nil, nil, err
	} else if directory != nil {
		return nil, nil, fmt.Errorf("Can't download library specification; resource '%s' points at a file", gh.registrySpecRawURL())
	}

	partsSpecText, err := file.GetContent()
	if err != nil {
		return nil, nil, err
	}

	parts := parts.Spec{}
	err = yaml.Unmarshal([]byte(partsSpecText), &parts)
	if err != nil {
		return nil, nil, err
	}

	refSpec := app.LibraryRefSpec{
		Name:     libAlias,
		Registry: gh.Name,
		GitVersion: &app.GitVersionSpec{
			RefSpec:   libRefSpec,
			CommitSHA: resolvedSHA,
		},
	}

	return &parts, &refSpec, nil
}

func (gh *gitHubRegistryManager) resolveDir(client *github.Client, libID, path, version string, onFile registry.ResolveFile, onDir registry.ResolveDirectory) error {
	ctx := context.Background()
	getOpts := &github.RepositoryContentGetOptions{Ref: version}
	file, directory, _, err := client.Repositories.GetContents(ctx, gh.org, gh.repo, path, getOpts)
	if err != nil {
		return err
	} else if file != nil {
		return fmt.Errorf("Lib ID '%s' resolves to a file in registry '%s'", libID, gh.Name)
	}

	for _, item := range directory {
		switch item.GetType() {
		case "file":
			itemPath := item.GetPath()
			file, directory, _, err := client.Repositories.GetContents(ctx, gh.org, gh.repo, itemPath, getOpts)
			if err != nil {
				return err
			} else if directory != nil {
				return fmt.Errorf("INTERNAL ERROR: GitHub API reported resource '%s' of type file, but returned type dir", itemPath)
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
			if err := gh.resolveDir(client, libID, itemPath, version, onFile, onDir); err != nil {
				return err
			}
		case "symlink":
		case "submodule":
			return fmt.Errorf("Invalid library '%s'; ksonnet doesn't support libraries with symlinks or submodules", libID)
		}
	}

	return nil
}

func (gh *gitHubRegistryManager) registrySpecRawURL() string {
	return strings.Join([]string{rawGitHubRoot, gh.org, gh.repo, gh.GitVersion.RefSpec, gh.registrySpecRepoPath}, "/")
}

func parseGitHubURI(uri string) (org, repo, refSpec, regRepoPath, regSpecRepoPath string, err error) {
	// Normalize URI.
	uri = strings.TrimSpace(uri)
	if strings.HasPrefix(uri, "http://github.com") || strings.HasPrefix(uri, "https://github.com") || strings.HasPrefix(uri, "http://www.github.com") || strings.HasPrefix(uri, "https://www.github.com") {
		// Do nothing.
	} else if strings.HasPrefix(uri, "github.com") || strings.HasPrefix(uri, "www.github.com") {
		uri = "http://" + uri
	} else {
		return "", "", "", "", "", fmt.Errorf("Registries using protocol 'github' must provide URIs beginning with 'github.com' (optionally prefaced with 'http', 'https', 'www', and so on")
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return "", "", "", "", "", err
	}

	if len(parsed.Query()) != 0 {
		return "", "", "", "", "", fmt.Errorf("No query strings allowed in registry URI:\n%s", uri)
	}

	components := strings.Split(parsed.Path, "/")
	if len(components) < 3 {
		return "", "", "", "", "", fmt.Errorf("GitHub URI must point at a respository:\n%s", uri)
	}

	// NOTE: The first component is always blank, because the path
	// begins like: '/whatever'.
	org = components[1]
	repo = components[2]

	//
	// Parse out `regSpecRepoPath`. There are a few cases:
	//   * URI points at a directory inside the respoitory, e.g.,
	//     'http://github.com/ksonnet/parts/tree/master/incubator'
	//   * URI points at an 'app.yaml', e.g.,
	//     'http://github.com/ksonnet/parts/blob/master/app.yaml'
	//   * URI points at a repository root, e.g.,
	//     'http://github.com/ksonnet/parts'
	//
	if len := len(components); len > 4 {
		refSpec = components[4]

		//
		// Case where we're pointing at either a directory inside a GitHub
		// URL, or an 'app.yaml' inside a GitHub URL.
		//

		// See note above about first component being blank.
		if components[3] == "tree" {
			regRepoPath = strings.Join(components[5:], "/")

			// If we have a trailing '/' character, last component will be
			// blank.
			if components[len-1] == "" {
				components[len-1] = registryYAMLFile
			} else {
				components = append(components, registryYAMLFile)
			}
			regSpecRepoPath = strings.Join(components[5:], "/")
			return
		} else if components[3] == "blob" && components[len-1] == registryYAMLFile {
			regRepoPath = strings.Join(components[5:len-1], "/")
			// Path to the `registry.yaml` (may or may not exist).
			regSpecRepoPath = strings.Join(components[5:], "/")
			return
		} else {
			return "", "", "", "", "", fmt.Errorf("Invalid GitHub URI: try navigating in GitHub to the URI of the folder containing the 'app.yaml', and using that URI instead. Generally, this URI should be of the form 'github.com/{organization}/{repository}/tree/{branch}/[path-to-directory]'")
		}
	} else {
		refSpec = defaultGitHubBranch

		// Else, URI should point at repository root.
		if components[len-1] == "" {
			components[len-1] = defaultGitHubBranch
			components = append(components, registryYAMLFile)
		} else {
			components = append(components, defaultGitHubBranch, registryYAMLFile)
		}

		regRepoPath = ""
		regSpecRepoPath = registryYAMLFile
		return
	}
}
