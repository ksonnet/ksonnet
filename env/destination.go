package env

import "encoding/json"

const (
	// destDefaultNamespace is the default namespace name.
	destDefaultNamespace = "default"
)

// Destination contains destination information for a cluster.
type Destination struct {
	server    string
	namespace string
}

// NewDestination creates an instance of Destination.
func NewDestination(server, namespace string) Destination {
	return Destination{
		server:    server,
		namespace: namespace,
	}
}

// MarshalJSON marshals a Destination to JSON.
func (d *Destination) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Server    string `json:"server"`
		Namespace string `json:"namespace"`
	}{
		Server:    d.Server(),
		Namespace: d.Namespace(),
	})
}

// Server is URL to the Kubernetes server that the cluster is running on.
func (d *Destination) Server() string {
	return d.server
}

// Namespace is the namespace of the Kubernetes server that targets should
// be deployed.
func (d *Destination) Namespace() string {
	if d.namespace == "" {
		return destDefaultNamespace
	}

	return d.namespace
}
