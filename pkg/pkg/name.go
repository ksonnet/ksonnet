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

package pkg

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var errInvalidSpec = fmt.Errorf("package name should be in the form `<registry>/<library>@<version>`")

// Descriptor describes a package.
type Descriptor struct {
	Registry string
	Part     string
	Version  string
}

// ParseName parses a package name into its components
func ParseName(name string) (Descriptor, error) {
	split := strings.SplitN(name, "/", 2)
	if len(split) < 2 {
		return Descriptor{}, errInvalidSpec
	}

	registryName := split[0]
	partName := strings.SplitN(split[1], "@", 2)[0]

	split = strings.Split(name, "@")
	if len(split) > 2 {
		return Descriptor{}, errors.Errorf("symbol '@' is only allowed once, at the end of the argument of the form <registry>/<library>@<version>")
	}
	var version string
	if len(split) == 2 {
		version = split[1]
	}

	return Descriptor{
		Registry: registryName,
		Part:     partName,
		Version:  version,
	}, nil
}
