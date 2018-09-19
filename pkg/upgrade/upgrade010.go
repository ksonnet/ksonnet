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

package upgrade

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/lib"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// PackageLister lists installed packages.
type PackageLister interface {
	Packages() ([]pkg.Package, error)
}

type upgrade010 struct {
	out           io.Writer
	app           app.App
	packageLister PackageLister
}

// NewUpgrade010 constructs an Upgrader from version 0.1.0->0.2.0
func newUpgrade010(a app.App, out io.Writer, pl PackageLister) *upgrade010 {
	return &upgrade010{
		app:           a,
		out:           out,
		packageLister: pl,
	}
}

// Upgrade upgrades the app to the latest apiVersion.
func (u *upgrade010) Upgrade(dryRun bool) error {
	if u == nil {
		return errors.New("nil receiver")
	}
	if u.app == nil {
		return errors.New("nil app")
	}

	if err := u.checkForOldKSLibLocation(dryRun); err != nil {
		return errors.Wrapf(err, "upgrading kslib location")
	}

	if u.packageLister == nil {
		return errors.Errorf("nil packageLister")
	}
	if err := u.upgradeOldVendoredPackages(u.packageLister, dryRun); err != nil {
		return errors.Wrapf(err, "upgrading vendored packages")
	}
	if err := u.upgradeEnvTargets(u.app, dryRun); err != nil {
		return errors.Wrapf(err, "upgrading environment targets")
	}

	// Upgrade app.yaml
	return u.app.Upgrade(dryRun)
}

var (
	// reKSLibName matches a ksonnet library directory e.g. v1.10.3.
	reKSLibName = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
)

func (u *upgrade010) checkForOldKSLibLocation(dryRun bool) error {
	fs := u.app.Fs()
	libRoot := filepath.Join(u.app.Root(), app.LibDirName)
	fis, err := afero.ReadDir(fs, libRoot)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintf(u.out, "[dry run] Updating ksonnet-lib paths\n")
	} else {
		if err = fs.MkdirAll(filepath.Join(libRoot, lib.KsonnetLibHome), app.DefaultFolderPermissions); err != nil {
			return err
		}
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		if reKSLibName.MatchString(fi.Name()) {
			p := filepath.Join(libRoot, fi.Name())
			new := filepath.Join(libRoot, lib.KsonnetLibHome, fi.Name())

			if dryRun {
				fmt.Fprintf(u.out, "[dry run] Moving %q from %s to %s\n", fi.Name(), p, new)
				continue
			}

			fmt.Fprintf(u.out, "Moving %q from %s to %s\n", fi.Name(), p, new)
			err = fs.Rename(p, new)
			if err != nil {
				return errors.Wrapf(err, "renaming %s to %s", p, new)
			}
		}
	}

	return nil
}

var removeVersionPattern = regexp.MustCompile("(.*)@.*$")

// In ksonnet 0.12.0 (app version 0.2.0) - the vendor cache began storing
// packages using versioned paths. This upgrade translates old paths into new, versioned paths.
// Example:
// `vendor/incubator/mysql` -> `vendor/incubator/mysql@1.2.3`
func (u *upgrade010) upgradeOldVendoredPackages(pl PackageLister, dryRun bool) error {
	if u == nil {
		return errors.Errorf("nil receiver")
	}
	if u.app == nil {
		return errors.Errorf("nil app")
	}
	fs := u.app.Fs()
	if fs == nil {
		return errors.Errorf("nil filesystem interface")
	}
	if pl == nil {
		return errors.Errorf("nil packageLister")
	}

	if dryRun {
		fmt.Fprintf(u.out, "[dry run] Updating vendored packages\n")
	}

	pkgs, err := pl.Packages()
	if err != nil {
		return errors.Wrapf(err, "resolving packages")
	}

	for _, p := range pkgs {
		if p.Version() == "" {
			// Skip unversioned packages
			continue
		}

		versioned := p.Path()

		ok, err := afero.Exists(fs, versioned)
		if err != nil {
			return errors.Wrapf(err, "checking package: %v", p)
		}
		if ok {
			// Already upgraded
			continue
		}

		// Check for unversioned path
		unversioned := string(removeVersionPattern.ReplaceAll([]byte(versioned), []byte("$1")))
		ok, err = afero.Exists(fs, unversioned)
		if err != nil {
			return errors.Wrapf(err, "checking path: %v", unversioned)
		}
		if !ok {
			// Nothing to upgrade, the package is simply missing
			continue
		}

		// Ok, time to upgrade -> unversioned -> versioned
		if dryRun {
			fmt.Fprintf(u.out, "[dry run] Moving %q from %s to %s\n", p, unversioned, versioned)
			continue
		}
		if err := fs.Rename(unversioned, versioned); err != nil {
			return errors.Wrapf(err, "renaming %v to %v", unversioned, versioned)
		}
	}

	return nil
}

type envListerUpdater interface {
	// Environments returns all environments.
	Environments() (app.EnvironmentConfigs, error)
	// UpdateTargets sets the targets for an environment.
	UpdateTargets(envName string, targets []string, isOverride bool) error
}

// In ksonnet 0.12.0 (app version 0.2.0) - environment targets began using the dot (.) separator for modules.
// Example:
// `module_a/module_b/module_c` -> `module_a.module_b.component_a`
func (*upgrade010) upgradeEnvTargets(listerUpdater envListerUpdater, dryRun bool) error {
	if listerUpdater == nil {
		return errors.New("nil envListerUpdater")
	}

	targets := make(map[string][]string)

	envs, err := listerUpdater.Environments()
	if err != nil {
		return errors.Wrap(err, "fetching environments")
	}
	for _, e := range envs {
		if e == nil {
			continue
		}
		var dirty bool
		upgraded := make([]string, len(e.Targets))
		for i, t := range e.Targets {
			if t == "/" {
				// Root module is an exception - leave as-is
				upgraded[i] = t
				continue
			}
			if strings.Contains(t, "/") {
				dirty = true
				upgraded[i] = strings.Replace(t, "/", ".", -1)
			}
		}
		if !dirty || dryRun {
			continue
		}
		targets[e.Name] = upgraded
	}

	for name, upgraded := range targets {
		if err := listerUpdater.UpdateTargets(name, upgraded, false); err != nil {
			return errors.Wrapf(err, "updating targets for environment: %s", name)
		}
	}

	return nil
}

// Returns library configurations for the app, whether they are global or environment-scoped.
func allLibraries(a app.App) (app.LibraryConfigs, error) {
	if a == nil {
		return nil, errors.Errorf("nil receiver")
	}

	combined := app.LibraryConfigs{}

	libs, err := a.Libraries()
	if err != nil {
		return nil, errors.Wrapf(err, "checking libraries")
	}
	for _, lib := range libs {
		combined[lib.Name] = lib
	}

	envs, err := a.Environments()
	if err != nil {
		return nil, errors.Wrapf(err, "checking environments")
	}
	for _, env := range envs {
		for _, lib := range env.Libraries {
			// NOTE We do not check for collisions at this time
			combined[lib.Name] = lib
		}
	}

	return combined, nil
}

// CheckUpgrade checks for any configuration that needs to be upgraded.
// If any is found, the user will be informed and we will return true.
func (u *upgrade010) CheckUpgrade() (bool, error) {
	if u == nil {
		return false, errors.Errorf("nil reciever")
	}

	var needUpgrade bool

	legacyPkgs, err := u.checkUpgradeVendoredPackages()
	if err != nil {
		return false, err
	}

	if len(legacyPkgs) > 0 {
		logrus.Warnf("Versioned packages stored in unversioned paths.")
		needUpgrade = true
	}

	// Check overrides
	if app.CheckOverrideUpgrade(u.app.Fs(), u.app.Root()) {
		logrus.Warn("Application override should be upgraded.")
		needUpgrade = true

	}

	if needUpgrade {
		logrus.Warnf("Your application must be upgraded to work with this version of ksonnet. Please run `ks upgrade` proceed.")
	}

	return needUpgrade, nil
}

// checkUpgradeVendoredPackages checks whether vendored packages need to be upgraded.
// Upgrades are necessary if a versioned package is stored in pre-ksonnet 0.12.0, unversioned directory.
func (u *upgrade010) checkUpgradeVendoredPackages() ([]pkg.Package, error) {
	if u == nil {
		return nil, errors.Errorf("nil receiver")
	}
	a := u.app
	if a == nil {
		return nil, errors.Errorf("nil app")
	}
	fs := a.Fs()
	if fs == nil {
		return nil, errors.Errorf("nil filesystem interface")
	}

	pkgs, err := u.packageLister.Packages()
	if err != nil {
		return nil, errors.Wrap(err, "listing packages")
	}

	results := make([]pkg.Package, 0)
	for _, p := range pkgs {
		if p.Version() == "" {
			continue
		}

		path := p.Path()
		ok, err := afero.DirExists(fs, path)
		if err != nil {
			return nil, err
		}
		if !ok {
			results = append(results, p)
		}
	}

	return results, nil
}
