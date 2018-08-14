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
	"net/http"

	"github.com/ksonnet/ksonnet/pkg/lib"
	"github.com/spf13/afero"
)

// KSLibUpdater generates / updates ksonnet-lib matching a specific k8s version.
type KSLibUpdater interface {
	// Generates a ksonnet-lib matching the specified k8s version.
	// Returns the generated version.
	UpdateKSLib(k8sSpecFlag string, libPath string) (string, error)
}

type ksLibUpdater struct {
	fs         afero.Fs
	httpClient *http.Client
}

// Implements KSLibUpdater
func (k ksLibUpdater) UpdateKSLib(k8sSpecFlag string, libPath string) (string, error) {
	lm, err := lib.NewManager(k8sSpecFlag, k.fs, libPath, k.httpClient)
	if err != nil {
		return "", err
	}

	if err := lm.GenerateLibData(); err != nil {
		return "", err
	}

	return lm.K8sVersion, nil
}
