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

package version

import (
	"strings"

	"github.com/blang/semver"
)

// Version represents a version.
type Version struct {
	raw string
	v   semver.Version
}

// Make takes a string and converts it to a version. It supports the following:
// * 1.2.3
// * 1.2
// * 1
// * v1.2.3
// * 1.2.3-build-1
func Make(s string) (Version, error) {
	versionStr := strings.TrimPrefix(s, "v")

	parts := strings.SplitN(versionStr, ".", 3)

	switch len(parts) {
	case 3:
		// nothing to do because we have three parts of a version
	case 2:
		// assume we have major and minor
		versionStr = strings.Join(append(parts, "0"), ".")
	case 1:
		// assume we have major
		versionStr = strings.Join(append(parts, "0", "0"), ".")
	}

	v, err := semver.Make(versionStr)
	if err != nil {
		return Version{}, err
	}

	return Version{
		raw: s,
		v:   v,
	}, nil
}

func (v Version) String() string {
	return v.raw
}

// LT checks if v is less than o.
func (v Version) LT(o Version) bool {
	return v.v.LT(o.v)
}
