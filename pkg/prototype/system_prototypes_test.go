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

package prototype

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_defaultPrototypes(t *testing.T) {
	cases := []struct {
		name      string
		container func(*testing.T) *fakeContainer
		pb        func(string) (*Prototype, error)
		isErr     bool
	}{
		{
			name: "with no errors",
			container: func(t *testing.T) *fakeContainer {
				return &fakeContainer{t: t}
			},
			pb: func(source string) (*Prototype, error) {
				return &Prototype{}, nil
			},
		},
		{
			name: "prototype build error",
			container: func(t *testing.T) *fakeContainer {
				return &fakeContainer{t: t}
			},
			pb: func(source string) (*Prototype, error) {
				return nil, errors.New("fail")
			},
			isErr: true,
		},
		{
			name: "walk error",
			container: func(t *testing.T) *fakeContainer {
				return &fakeContainer{
					t:       t,
					walkErr: true,
				}
			},
			isErr: true,
		},
		{
			name: "string error",
			container: func(t *testing.T) *fakeContainer {
				return &fakeContainer{
					t:         t,
					stringErr: true,
				}
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prototypes, err := systemPrototypes(tc.container(t), tc.pb)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Len(t, prototypes, 1)
		})
	}

}

type fakeContainer struct {
	t         *testing.T
	walkErr   bool
	stringErr bool
}

func (fc *fakeContainer) String(path string) (string, error) {
	if fc.stringErr {
		return "", errors.New("fail")
	}

	name := filepath.FromSlash("testdata/prototype1.jsonnet")

	b, err := ioutil.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (fc *fakeContainer) Walk(path string, fn filepath.WalkFunc) error {
	if fc.walkErr {
		return fn("", nil, errors.New("fail"))
	}

	name := filepath.FromSlash("testdata/prototype1.jsonnet")

	dirFi, err := os.Stat("testdata")
	require.NoError(fc.t, err)

	if err = fn("testdata", dirFi, nil); err != nil {
		return err
	}

	fi, err := os.Stat(name)
	require.NoError(fc.t, err)

	if err = fn(name, fi, nil); err != nil {
		return err
	}

	return nil
}
