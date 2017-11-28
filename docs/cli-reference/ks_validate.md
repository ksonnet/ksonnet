## ks validate

Compare generated manifest against server OpenAPI spec

### Synopsis


Validate that an application or file is compliant with the Kubernetes
specification.

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.

```
ks validate [env-name] [-f <file-or-dir>]
```

### Examples

```
# Validate all resources described in a ksonnet application, expanding
# ksonnet code with 'dev' environment where necessary (i.e., not YAML, JSON,
# or non-ksonnet Jsonnet code).
ksonnet validate dev

# Validate resources described in a YAML file.
ksonnet validate -f ./pod.yaml

# Validate resources described in the JSON file against existing resources
# in the cluster the 'dev' environment is pointing at.
ksonnet validate dev -f ./pod.yaml

# Validate resources described in a Jsonnet file. Does not expand using
# environment bindings.
ksonnet validate -f ./pod.jsonnet
```

### Options

```
      --as string                      Username to impersonate for the operation
      --certificate-authority string   Path to a cert. file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
  -c, --component stringArray          Name of a specific component (multiple -c flags accepted, allows YAML, JSON, and Jsonnet)
      --context string                 The name of the kubeconfig context to use
  -V, --ext-str stringSlice            Values of external variables
      --ext-str-file stringSlice       Read external variable from a file
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -J, --jpath stringSlice              Additional jsonnet library search path
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --resolve-images string          Change implementation of resolveImage native function. One of: noop, registry (default "noop")
      --resolve-images-error string    Action when resolveImage fails. One of ignore,warn,error (default "warn")
      --server string                  The address and port of the Kubernetes API server
  -A, --tla-str stringSlice            Values of top level arguments
      --tla-str-file stringSlice       Read top level argument from a file
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files

