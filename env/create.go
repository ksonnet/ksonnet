package env

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// CreateConfig is configuration for creating an environment.
type CreateConfig struct {
	App         app.App
	Destination Destination
	Fs          afero.Fs
	K8sSpecFlag string
	Name        string
	RootPath    string

	OverrideData []byte
	ParamsData   []byte
}

// Create creates a new environment for the project.
func Create(config CreateConfig) error {
	c, err := newCreator(config)
	if err != nil {
		return err
	}
	return c.Create()
}

type creator struct {
	CreateConfig
}

func newCreator(config CreateConfig) (*creator, error) {
	return &creator{
		CreateConfig: config,
	}, nil
}

func (c *creator) Create() error {
	if c.environmentExists() {
		return errors.Errorf("Could not create %q", c.Name)
	}

	// ensure environment name does not contain punctuation
	if !isValidName(c.Name) {
		return fmt.Errorf("Environment name %q is not valid; must not contain punctuation, spaces, or begin or end with a slash", c.Name)
	}

	log.Infof("Creating environment %q with namespace %q, pointing to cluster at address %q",
		c.Name, c.Destination.Namespace(), c.Destination.Server())

	envPath := filepath.Join(c.RootPath, app.EnvironmentDirName, c.Name)
	err := c.Fs.MkdirAll(envPath, app.DefaultFolderPermissions)
	if err != nil {
		return err
	}

	metadata := []struct {
		path string
		data []byte
	}{
		{
			// environment base override file
			filepath.Join(envPath, envFileName),
			c.OverrideData,
		},
		{
			// params file
			filepath.Join(envPath, paramsFileName),
			c.ParamsData,
		},
	}

	for _, a := range metadata {
		fileName := path.Base(a.path)
		log.Debugf("Generating '%s', length: %d", fileName, len(a.data))
		if err = afero.WriteFile(c.Fs, a.path, a.data, app.DefaultFilePermissions); err != nil {
			log.Debugf("Failed to write '%s'", fileName)
			return err
		}
	}

	// update app.yaml
	err = c.App.AddEnvironment(c.Name, c.K8sSpecFlag, &app.EnvironmentSpec{
		Path: c.Name,
		Destination: &app.EnvironmentDestinationSpec{
			Server:    c.Destination.Server(),
			Namespace: c.Destination.Namespace(),
		},
	})

	return err
}

func (c *creator) environmentExists() bool {
	_, err := c.App.Environment(c.Name)
	return err == nil
}

// isValidName returns true if a name (e.g., for an environment) is valid.
// Broadly, this means it does not contain punctuation, whitespace, leading or
// trailing slashes.
func isValidName(name string) bool {
	// No unicode whitespace is allowed. `Fields` doesn't handle trailing or
	// leading whitespace.
	fields := strings.Fields(name)
	if len(fields) > 1 || len(strings.TrimSpace(name)) != len(name) {
		return false
	}

	hasPunctuation := regexp.MustCompile(`[\\,;':!()?"{}\[\]*&%@$]+`).MatchString
	hasTrailingSlashes := regexp.MustCompile(`/+$`).MatchString
	hasLeadingSlashes := regexp.MustCompile(`^/+`).MatchString
	return len(name) != 0 && !hasPunctuation(name) && !hasTrailingSlashes(name) && !hasLeadingSlashes(name)
}
