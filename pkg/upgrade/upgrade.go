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
	"github.com/pkg/errors"
)

// Upgrade upgrades an application to the current version.
func Upgrade(a app.App, out io.Writer, pl PackageLister, dryRun bool) error {
	// TODO new migration framework goes here

	if a == nil {
		return errors.Errorf("nil receiver")
	}

	switch va := a.(type) {
	default:
		return errors.Errorf("Unknown app type: %T", a)
	case *app.App001:
		// First we upgrade 0.0.1 -> 0.1.0, then 0.1.0 -> 0.2.0
		u := newUpgrade001(va)
		err := u.Upgrade(dryRun)
		if err != nil {
			return errors.Wrapf(err, "upgrading from 0.0.1 to 0.1.0")
		}

		// Reload App between upgrades
		app010, err := app.Load(va.Fs(), va.HTTPClient(), va.Root())
		if err != nil {
			return errors.Wrapf(err, "reloading app after 0.1.0 upgrade")
		}

		u2 := newUpgrade010(app010, out, pl)
		err = u2.Upgrade(dryRun)
		if err != nil {
			return errors.Wrapf(err, "upgrading from 0.1.0 to 0.2.0")
		}

		return nil
	case *app.App010:
		u := newUpgrade010(va, out, pl)
		err := u.Upgrade(dryRun)
		if err != nil {
			return errors.Wrapf(err, "upgrading from 0.1.0 to 0.2.0")
		}
		return nil
	}
}
