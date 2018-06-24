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

package dockerregistry

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	// defaultRegistry is the default Docker registry
	defaultRegistry = "registry-1.docker.io"
)

// ImageName represents the parts of a docker image name
// eg: "myregistryhost:5000/fedora/httpd:version1.0"
type ImageName struct {
	// Registry is the registry api address
	Registry string
	// Repository is the repository name
	Repository string
	// Name is the name of the image
	Name string
	// Tag is the image tag
	Tag string
	// Digest is the image digest
	Digest string
}

// String implements the Stringer interface
func (n ImageName) String() string {
	buf := bytes.Buffer{}
	if n.Registry != "" {
		buf.WriteString(n.Registry)
		buf.WriteString("/")
	}
	if n.Repository != "" {
		buf.WriteString(n.Repository)
		buf.WriteString("/")
	}
	buf.WriteString(n.Name)
	if n.Digest != "" {
		buf.WriteString("@")
		buf.WriteString(n.Digest)
	} else {
		buf.WriteString(":")
		buf.WriteString(n.Tag)
	}
	return buf.String()
}

// RegistryRepoName returns the "repository" as used in the registry URL
func (n ImageName) RegistryRepoName() string {
	repo := n.Repository
	if repo == "" {
		repo = "library"
	}
	return fmt.Sprintf("%s/%s", repo, n.Name)
}

// RegistryURL returns the deduced base URL of the registry for this image
func (n ImageName) RegistryURL() string {
	reg := n.Registry
	if reg == "" {
		reg = defaultRegistry
	}
	return fmt.Sprintf("https://%s", reg)
}

// ParseImageName parses a docker image into an ImageName struct
func ParseImageName(image string) (ImageName, error) {
	ret := ImageName{}

	if parts := strings.Split(image, "/"); len(parts) == 1 {
		ret.Name = parts[0]
	} else if len(parts) == 2 {
		ret.Repository = parts[0]
		ret.Name = parts[1]
	} else if len(parts) == 3 {
		ret.Registry = parts[0]
		ret.Repository = parts[1]
		ret.Name = parts[2]
	} else {
		return ret, fmt.Errorf("Malformed docker image name: %s", image)
	}

	if parts := strings.Split(ret.Name, "@"); len(parts) == 2 {
		ret.Name = parts[0]
		ret.Digest = parts[1]
	} else if parts := strings.Split(ret.Name, ":"); len(parts) == 2 {
		ret.Name = parts[0]
		ret.Tag = parts[1]
	} else if len(parts) == 1 {
		ret.Name = parts[0]
		ret.Tag = "latest"
	} else {
		return ret, fmt.Errorf("Malformed docker image name/tag: %s", image)
	}

	return ret, nil
}
