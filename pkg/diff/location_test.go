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
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocation(t *testing.T) {
	cases := []struct {
		name        string
		src         string
		destination string
		envName     string
		isErr       bool
	}{
		{
			name:        "default",
			src:         "default",
			destination: "local",
			envName:     "default",
		},
		{
			name:        "local:default",
			src:         "local:default",
			destination: "local",
			envName:     "default",
		},
		{
			name:  "blank",
			isErr: true,
		},
		{
			name:  "invalid destination",
			src:   "other:default",
			isErr: true,
		},
		{
			name:  "wrong format",
			src:   "other:default:something",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLocation(tc.src)
			spew.Dump(l)
			if tc.isErr {
				require.Error(t, l.Err())
				return
			}

			assert.Equal(t, tc.destination, l.Destination())
			assert.Equal(t, tc.envName, l.EnvName())
			assert.Equal(t, fmt.Sprintf("%s:%s", l.Destination(), l.EnvName()), l.String())
		})
	}
}
