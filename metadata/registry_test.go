package metadata

import (
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
)

type ghRegistryGetSuccess struct {
	target string
	source string
}

func (r *ghRegistryGetSuccess) Test(t *testing.T) {
	gh, err := makeGitHubRegistryManager(&app.RegistryRefSpec{
		Protocol: "github",
		Spec: map[string]interface{}{
			uriField:     r.source,
			refSpecField: "master",
		},
		Name: "incubator",
	})
	if err != nil {
		t.Error(err)
	}

	rawURI := gh.registrySpecRawURL()
	if rawURI != r.target {
		t.Errorf("Expected URI '%s', got '%s'", r.target, rawURI)
	}
}

func TestGetRegistryRefSuccess(t *testing.T) {
	successes := []*ghRegistryGetSuccess{
		&ghRegistryGetSuccess{
			source: "http://github.com/ksonnet/parts",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "http://github.com/ksonnet/parts/",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "http://www.github.com/ksonnet/parts",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "https://www.github.com/ksonnet/parts",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "github.com/ksonnet/parts",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "www.github.com/ksonnet/parts",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},

		&ghRegistryGetSuccess{
			source: "http://github.com/ksonnet/parts/tree/master/incubator",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/incubator/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "http://github.com/ksonnet/parts/tree/master/incubator/",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/incubator/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "http://www.github.com/ksonnet/parts/tree/master/incubator",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/incubator/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "https://github.com/ksonnet/parts/tree/master/incubator",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/incubator/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "github.com/ksonnet/parts/tree/master/incubator",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/incubator/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "www.github.com/ksonnet/parts/tree/master/incubator",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/incubator/registry.yaml",
		},

		&ghRegistryGetSuccess{
			source: "http://github.com/ksonnet/parts/blob/master/registry.yaml",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "http://www.github.com/ksonnet/parts/blob/master/registry.yaml",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "https://github.com/ksonnet/parts/blob/master/registry.yaml",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "github.com/ksonnet/parts/blob/master/registry.yaml",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
		&ghRegistryGetSuccess{
			source: "www.github.com/ksonnet/parts/blob/master/registry.yaml",
			target: "https://raw.githubusercontent.com/ksonnet/parts/master/registry.yaml",
		},
	}

	for _, success := range successes {
		success.Test(t)
	}
}

//
// TODO: Add failure tests, like:
//
// &ghRegistryGetSuccess{
// 	source: "http://github.com/ksonnet/parts/tree/master/incubator?foo=bar",
// },
//

func TestCacheGitHubRegistry(t *testing.T) {
	// registrySpec, err := cacheGitHubRegistry("")
	// if err != nil {
	// 	t.Error(err)
	// }

	// panic(registrySpec)
}
