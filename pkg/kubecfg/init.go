// Copyright 2018 The kubecfg authors
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

package kubecfg

import (
	"github.com/ksonnet/ksonnet/metadata"
	log "github.com/sirupsen/logrus"
)

type InitCmd struct {
	name        string
	rootPath    string
	k8sSpecFlag *string
	serverURI   *string
	namespace   *string
}

func NewInitCmd(name, rootPath string, k8sSpecFlag, serverURI, namespace *string) (*InitCmd, error) {
	return &InitCmd{name: name, rootPath: rootPath, k8sSpecFlag: k8sSpecFlag, serverURI: serverURI, namespace: namespace}, nil
}

func (c *InitCmd) Run() error {
	_, err := metadata.Init(c.name, c.rootPath, c.k8sSpecFlag, c.serverURI, c.namespace)
	if err == nil {
		log.Info("ksonnet app successfully created! Next, try creating a component with `ks generate`.")
	}
	return err
}
