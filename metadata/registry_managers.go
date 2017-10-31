package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-github/github"
	"github.com/ksonnet/ksonnet/metadata/app"
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
	RefSpec              string `json:"refSpec"`
	ResolvedSHA          string `json:"resolvedSHA"`
	org                  string
	repo                 string
	registrySpecRepoPath string
}

func makeGitHubRegistryManager(registryRef *app.RegistryRefSpec) (*gitHubRegistryManager, error) {
	gh := gitHubRegistryManager{RegistryRefSpec: registryRef}

	var err error

	// Set registry path.
	// NOTE: Resolve this to a specific commit.
	gh.registryDir = gh.Name

	rawURI, uriExists := gh.Spec[uriField]
	uri, isString := rawURI.(string)
	if !uriExists || !isString {
		return nil, fmt.Errorf("GitHub app registry '%s' is missing a 'uri' in field 'spec'", gh.Name)
	}

	gh.org, gh.repo, gh.RefSpec, gh.registrySpecRepoPath, err = parseGitHubURI(uri)
	if err != nil {
		return nil, err
	}

	return &gh, nil
}

func (gh *gitHubRegistryManager) VersionsDir() string {
	return gh.registryDir
}

func (gh *gitHubRegistryManager) SpecPath() string {
	if gh.ResolvedSHA != "" {
		return path.Join(gh.registryDir, gh.ResolvedSHA+".yaml")
	}
	return path.Join(gh.registryDir, gh.RefSpec+".yaml")
}

func (gh *gitHubRegistryManager) FindSpec() (*registry.Spec, error) {
	// Fetch app spec at specific commit.
	client := github.NewClient(nil)
	ctx := context.Background()

	sha, _, err := client.Repositories.GetCommitSHA1(ctx, gh.org, gh.repo, gh.RefSpec, "")
	if err != nil {
		return nil, err
	}
	gh.ResolvedSHA = sha

	// Get contents.
	getOpts := github.RepositoryContentGetOptions{Ref: gh.ResolvedSHA}
	file, _, _, err := client.Repositories.GetContents(ctx, gh.org, gh.repo, gh.registrySpecRepoPath, &getOpts)
	if file == nil {
		return nil, fmt.Errorf("Could not find valid registry at uri '%s/%s/%s' and refspec '%s' (resolves to sha '%s')", gh.org, gh.repo, gh.registrySpecRepoPath, gh.RefSpec, gh.ResolvedSHA)
	} else if err != nil {
		return nil, err
	}

	registrySpecText, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	// Deserialize, return.
	registrySpec := registry.Spec{}
	err = json.Unmarshal([]byte(registrySpecText), &registrySpec)
	if err != nil {
		return nil, err
	}

	return &registrySpec, nil
}

func (gh *gitHubRegistryManager) registrySpecRawURL() string {
	return strings.Join([]string{rawGitHubRoot, gh.org, gh.repo, gh.RefSpec, gh.registrySpecRepoPath}, "/")
}

func parseGitHubURI(uri string) (org, repo, refSpec, regSpecRepoPath string, err error) {
	// Normalize URI.
	uri = strings.TrimSpace(uri)
	if strings.HasPrefix(uri, "http://github.com") || strings.HasPrefix(uri, "https://github.com") || strings.HasPrefix(uri, "http://www.github.com") || strings.HasPrefix(uri, "https://www.github.com") {
		// Do nothing.
	} else if strings.HasPrefix(uri, "github.com") || strings.HasPrefix(uri, "www.github.com") {
		uri = "http://" + uri
	} else {
		return "", "", "", "", fmt.Errorf("Registries using protocol 'github' must provide URIs beginning with 'github.com' (optionally prefaced with 'http', 'https', 'www', and so on")
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return "", "", "", "", err
	}

	if len(parsed.Query()) != 0 {
		return "", "", "", "", fmt.Errorf("No query strings allowed in registry URI:\n%s", uri)
	}

	components := strings.Split(parsed.Path, "/")
	if len(components) < 3 {
		return "", "", "", "", fmt.Errorf("GitHub URI must point at a respository:\n%s", uri)
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
			// Path to the `registry.yaml` (may or may not exist).
			regSpecRepoPath = strings.Join(components[5:], "/")
			return
		} else {
			return "", "", "", "", fmt.Errorf("Invalid GitHub URI: try navigating in GitHub to the URI of the folder containing the 'app.yaml', and using that URI instead. Generally, this URI should be of the form 'github.com/{organization}/{repository}/tree/{branch}/[path-to-directory]'")
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

		regSpecRepoPath = registryYAMLFile
		return
	}
}

//
// Mock registry manager.
//

type mockRegistryManager struct {
	registryDir string
}

func newMockRegistryManager(name string) *mockRegistryManager {
	return &mockRegistryManager{
		registryDir: name,
	}
}

func (m *mockRegistryManager) VersionsDir() string {
	return m.registryDir
}

func (m *mockRegistryManager) SpecPath() string {
	return path.Join(m.registryDir, "master.yaml")
}

func (m *mockRegistryManager) FindSpec() (*registry.Spec, error) {
	registrySpec := registry.Spec{
		APIVersion: registry.DefaultApiVersion,
		Kind:       registry.DefaultKind,
	}

	return &registrySpec, nil
}
