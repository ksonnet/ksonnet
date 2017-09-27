package kubecfg

import "github.com/ksonnet/kubecfg/metadata"

type InitCmd struct {
	rootPath  metadata.AbsPath
	spec      metadata.ClusterSpec
	serverURI *string
	namespace *string
}

func NewInitCmd(rootPath metadata.AbsPath, specFlag string, serverURI, namespace *string) (*InitCmd, error) {
	// NOTE: We're taking `rootPath` here as an absolute path (rather than a partial path we expand to an absolute path)
	// to make it more testable.

	spec, err := metadata.ParseClusterSpec(specFlag)
	if err != nil {
		return nil, err
	}

	return &InitCmd{rootPath: rootPath, spec: spec, serverURI: serverURI, namespace: namespace}, nil
}

func (c *InitCmd) Run() error {
	_, err := metadata.Init(c.rootPath, c.spec, c.serverURI, c.namespace)
	return err
}
