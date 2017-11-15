package registry

import (
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/parts"
)

type ResolveFile func(relPath string, contents []byte) error
type ResolveDirectory func(relPath string) error

type Manager interface {
	RegistrySpecDir() string
	RegistrySpecFilePath() string
	FetchRegistrySpec() (*Spec, error)
	MakeRegistryRefSpec() *app.RegistryRefSpec
	ResolveLibrarySpec(libID, libRefSpec string) (*parts.Spec, error)
	ResolveLibrary(libID, libAlias, version string, onFile ResolveFile, onDir ResolveDirectory) (*parts.Spec, *app.LibraryRefSpec, error)
}
