package kubecfg

import (
	"github.com/ksonnet/ksonnet/metadata"
	log "github.com/sirupsen/logrus"
)

type InitCmd struct {
	name      string
	rootPath  metadata.AbsPath
	spec      metadata.ClusterSpec
	serverURI *string
	namespace *string
}

func NewInitCmd(name string, rootPath metadata.AbsPath, specFlag string, serverURI, namespace *string) (*InitCmd, error) {
	// NOTE: We're taking `rootPath` here as an absolute path (rather than a partial path we expand to an absolute path)
	// to make it more testable.

	spec, err := metadata.ParseClusterSpec(specFlag)
	if err != nil {
		return nil, err
	}

	return &InitCmd{name: name, rootPath: rootPath, spec: spec, serverURI: serverURI, namespace: namespace}, nil
}

func (c *InitCmd) Run() error {
	_, err := metadata.Init(c.name, c.rootPath, c.spec, c.serverURI, c.namespace)
	if err == nil {
		log.Info("ksonnet app successfully created! Next, try creating a component with `ks generate`.")
	}
	return err
}
