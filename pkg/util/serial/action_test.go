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

package serial

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RunActions(t *testing.T) {
	fn1 := func() error { return nil }
	fn2 := func() error { return nil }

	err := RunActions(fn1, fn2)
	require.NoError(t, err)
}

func Test_RunActions_failure(t *testing.T) {
	fn1 := func() error { return errors.New("failed") }
	fn2 := func() error { t.Fatal("should not have run"); return nil }

	err := RunActions(fn1, fn2)
	require.Error(t, err)
}
