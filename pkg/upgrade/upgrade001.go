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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
)

type upgrade001 struct {
	app app.App
}

// newUpgrade001 constructs an Upgrader from version 0.0.1->0.1.0
func newUpgrade001(a app.App) *upgrade001 {
	return &upgrade001{
		app: a,
	}
}

// Upgrade upgrades the app to the latest apiVersion.
func (u *upgrade001) Upgrade(dryRun bool) error {
	if u == nil {
		return errors.Errorf("nil receiver")
	}
	if u.app == nil {
		return errors.Errorf("nil app")
	}
	return u.app.Upgrade(dryRun)
}
