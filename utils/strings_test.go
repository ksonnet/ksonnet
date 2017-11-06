// Copyright 2017 The kubecfg authors
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

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsASCIIIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "HelloWorld",
			expected: true,
		},
		{
			input:    "Hello World",
			expected: false,
		},
		{
			input:    "helloworld",
			expected: true,
		},
		{
			input:    "hello-world",
			expected: false,
		},
		{
			input:    "hello世界",
			expected: false,
		},
	}
	for _, test := range tests {
		require.EqualValues(t, test.expected, IsASCIIIdentifier(test.input))
	}
}
