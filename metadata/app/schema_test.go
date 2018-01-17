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

package app

import (
	"fmt"
	"testing"

	"github.com/blang/semver"
)

func makeSimpleRefSpec(name, protocol, uri, version string) *RegistryRefSpec {
	return &RegistryRefSpec{
		Name:     name,
		Protocol: protocol,
		URI:      uri,
		GitVersion: &GitVersionSpec{
			RefSpec:   version,
			CommitSHA: version,
		},
	}
}

func TestApiVersionValidate(t *testing.T) {
	type spec struct {
		spec string
		err  bool
	}
	tests := []spec{
		// Versions that we accept.
		{spec: "0.0.1", err: false},
		{spec: "0.0.1+build.1", err: false},

		// Other versions.
		{spec: "0.0.0", err: true},
		{spec: "0.1.0", err: true},
		{spec: "1.0.0", err: true},

		// Builds and pre-releases of current version.
		{spec: "0.0.1-alpha", err: true},
		{spec: "0.0.1-beta+build.2", err: true},

		// Other versions.
		{spec: "0.1.0-alpha", err: true},
		{spec: "0.1.0+build.1", err: true},
		{spec: "0.1.0-beta+build.2", err: true},
		{spec: "1.0.0-alpha", err: true},
		{spec: "1.0.0+build.1", err: true},
		{spec: "1.0.0-beta+build.2", err: true},
	}

	for _, test := range tests {
		_, err := semver.Make(test.spec)
		if err != nil {
			t.Errorf("Failed to parse version '%s':\n%v", test.spec, err)
		}

		spec := &Spec{APIVersion: test.spec}
		err = spec.validate()
		if (test.err && err == nil) || (!test.err && err != nil) {
			t.Errorf("Expected error for version '%s'? %t. Value of error: '%v'", test.spec, test.err, err)
		}
	}
}

func TestGetRegistryRefSuccess(t *testing.T) {
	example1 := Spec{
		Registries: RegistryRefSpecs{
			"simple1": &RegistryRefSpec{
				URI:      "example.com",
				Protocol: "github",
			},
		},
	}

	r1, ok := example1.GetRegistryRef("simple1")
	fmt.Println(r1)
	if r1 == nil || !ok {
		t.Error("Expected registry to contain 'simple1'")
	}

	if r1.URI != "example.com" || r1.Name != "simple1" || r1.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%s", r1)
	}
}

func TestGetRegistryRefFailure(t *testing.T) {
	example1 := Spec{
		Registries: RegistryRefSpecs{
			"simple1": &RegistryRefSpec{
				URI:      "example.com",
				Protocol: "github",
			},
		},
	}

	r1, ok := example1.GetRegistryRef("simple2")
	if r1 != nil || ok {
		t.Error("Expected registry to not contain 'simple2'")
	}
}

func TestAddRegistryRefSuccess(t *testing.T) {
	var example1 = Spec{
		Registries: RegistryRefSpecs{},
	}

	err := example1.AddRegistryRef(makeSimpleRefSpec("simple1", "github", "example.com", "master"))
	if err != nil {
		t.Errorf("Expected registry add to succeed:\n%s", err)
	}

	r1, ok1 := example1.GetRegistryRef("simple1")
	if !ok1 || r1.URI != "example.com" || r1.Name != "simple1" || r1.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%s", r1)
	}

	r2, ok2 := example1.GetRegistryRef("simple1")
	if !ok2 || r2.URI != "example.com" || r2.Name != "simple1" || r2.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%s", r1)
	}
}

func TestAddRegistryRefFailure(t *testing.T) {
	example1 := Spec{
		Registries: RegistryRefSpecs{
			"simple1": &RegistryRefSpec{
				URI:      "example.com",
				Protocol: "github",
			},
		},
	}

	err := example1.AddRegistryRef(makeSimpleRefSpec("", "github", "example.com", "master"))
	if err != ErrRegistryNameInvalid {
		t.Error("Expected registry to fail to add registry with invalid name")
	}

	err = example1.AddRegistryRef(makeSimpleRefSpec("simple1", "fakeProtocol", "example.com", "master"))
	if err != ErrRegistryExists {
		t.Error("Expected registry to fail to add registry with duplicate name and different protocol")
	}

	err = example1.AddRegistryRef(makeSimpleRefSpec("simple1", "github", "fakeUrl", "master"))
	if err != ErrRegistryExists {
		t.Error("Expected registry to fail to add registry with duplicate name and different uri")
	}
}
