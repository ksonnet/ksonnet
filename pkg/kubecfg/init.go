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
