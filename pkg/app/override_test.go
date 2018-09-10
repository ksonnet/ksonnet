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

package app

import (
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestOverride_Validate(t *testing.T) {
	cases := []struct {
		name  string
		o     Override
		isErr bool
	}{
		{
			name: "valid override",
			o:    Override{Kind: overrideKind, APIVersion: overrideVersion},
		},
		{
			name:  "missing kind",
			o:     Override{APIVersion: overrideVersion},
			isErr: true,
		},
		{
			name:  "invalid kind",
			o:     Override{Kind: "invalid", APIVersion: overrideVersion},
			isErr: true,
		},
		{
			name:  "missing version",
			o:     Override{Kind: overrideKind},
			isErr: true,
		},
		{
			name:  "invalid version",
			o:     Override{APIVersion: "invalid", Kind: overrideKind},
			isErr: true,
		},
		{
			name:  "missing kind and version",
			o:     Override{},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.o.Validate()
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSaveOverride(t *testing.T) {
	cases := []struct {
		name    string
		o       *Override
		encoder Encoder
		isErr   bool
	}{
		{
			name:    "save override",
			o:       &Override{},
			encoder: defaultYAMLEncoder,
		},
		{
			name:    "encode error",
			o:       &Override{},
			encoder: &failEncoder{},
			isErr:   true,
		},
		{
			name:  "override is nil",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			err := saveOverride(tc.encoder, fs, "/", tc.o)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

type failEncoder struct{}

func (e *failEncoder) Encode(interface{}, io.Writer) error {
	return errors.Errorf("fail")
}
