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

package registry

import (
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/stretchr/testify/assert"
)

func TestGithub(t *testing.T) {
	spec := &app.RegistryRefSpec{
		Protocol: "github",
		URI:      "github.com/foo/bar",
	}

	g := NewGitHub("incubator", spec)

	assert.Equal(t, "incubator", g.Name())
	assert.Equal(t, "github", g.Protocol())
	assert.Equal(t, "github.com/foo/bar", g.URI())
}
