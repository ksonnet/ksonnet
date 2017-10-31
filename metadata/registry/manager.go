package registry

type Manager interface {
	VersionsDir() string
	SpecPath() string
	FindSpec() (*Spec, error)
}
