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

package diff

import (
	"fmt"
	gostrings "strings"

	"github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
)

var (
	diffDestinationNames = []string{"local", "remote"}

	errInvalidLocation = errors.New("invalid location. format is destination:environment or environment")
)

// Location is a diff location.
type Location struct {
	// destination is either `local` or `remote`
	destination string
	// envName is the environment name.
	envName string

	err error
}

// NewLocation creates a Location.
func NewLocation(src string) *Location {
	if src == "" {
		return &Location{err: errInvalidLocation}
	}

	parts := gostrings.Split(src, ":")

	l := &Location{}

	switch len(parts) {
	default:
		l.err = errInvalidLocation
	case 1:
		l.destination = "local"
		l.envName = parts[0]
	case 2:
		if !strings.InSlice(parts[0], diffDestinationNames) {
			l.err = errors.Errorf("%q is not a valid destination name", parts[0])
			break
		}
		l.destination = parts[0]
		l.envName = parts[1]
	}

	return l
}

// Err returns an error if this location is invalid.
func (l *Location) Err() error {
	return l.err
}

// Destination is the location destination.
func (l *Location) Destination() string {
	return l.destination
}

// EnvName is the environment name for the destination.
func (l *Location) EnvName() string {
	return l.envName
}

func (l *Location) String() string {
	return fmt.Sprintf("%s:%s", l.destination, l.envName)
}
