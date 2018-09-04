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
	"io"

	"github.com/ksonnet/ksonnet/pkg/app"
)

type upgrader interface {
	Upgrade(bool) error
}

type checker interface {
	CheckUpgrade() (bool, error)
}

// Upgrade upgrades an application to the current version.
func Upgrade(a app.App, out io.Writer, pl PackageLister, dryRun bool) error {
	var upgrades = []upgrader{
		newUpgrade010(a, out, pl),
	}

	for _, u := range upgrades {
		err := u.Upgrade(dryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckUpgrade checks whether an app should be upgraded.
func CheckUpgrade(a app.App, out io.Writer, pl PackageLister, dryRun bool) (bool, error) {
	var checks = []checker{
		newUpgrade010(a, out, pl),
	}

	var needsUpgrade bool
	for _, u := range checks {
		var err error
		needsUpgrade, err = u.CheckUpgrade()
		if err != nil {
			return false, err
		}
		if needsUpgrade {
			break
		}
	}
	return needsUpgrade, nil
}
