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

package strings

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRand(t *testing.T) {
	for i := 1; i < 45; i++ {
		t.Run(fmt.Sprintf("Rand(%d)", i), func(t *testing.T) {
			got := Rand(i)

			assert.Len(t, got, i)

			re := regexp.MustCompile(fmt.Sprintf(`^[A-Za-z]{%d}$`, i))
			assert.True(t, re.MatchString(got))
		})
	}
}

func TestRand_negative(t *testing.T) {
	got := Rand(-1)
	assert.Len(t, got, 0)
}

func TestLowerRand(t *testing.T) {
	for i := 1; i < 45; i++ {
		t.Run(fmt.Sprintf("LowerRand(%d)", i), func(t *testing.T) {
			got := LowerRand(i)

			assert.Len(t, got, i)

			re := regexp.MustCompile(fmt.Sprintf(`^[a-z]{%d}$`, i))
			assert.True(t, re.MatchString(got))
		})
	}
}

func TestLowerRand_negative(t *testing.T) {
	got := LowerRand(-1)
	assert.Len(t, got, 0)
}
