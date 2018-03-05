package actions

import (
	"os"

	"github.com/ksonnet/ksonnet/metadata"
)

// Upgrade upgrades a ksonnet application.
func Upgrade(dryRun bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	m, err := metadata.Find(cwd)
	if err != nil {
		return err
	}

	a, err := m.App()
	if err != nil {
		return err
	}

	return a.Upgrade(dryRun)
}
