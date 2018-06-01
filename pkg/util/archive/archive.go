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

package archive

import "io"

// File is an archived file.
type File struct {
	Name   string
	Reader io.Reader
}

// FileHandler is a functional that can process a File. It
// returns an error if there is one.
type FileHandler func(*File) error

// Unarchiver unarchives a reader and processes files using the
// FileHandler.
type Unarchiver interface {
	Unarchive(io.Reader, FileHandler) error
}
