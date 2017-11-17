## ks apply

Apply local configuration to remote cluster

### Synopsis


Update (or optionally create) Kubernetes resources on the cluster using the
local configuration. Use the `--create` flag to control whether we create them
if they do not exist (default: true).

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.

```
ks apply [env-name] [-f <file-or-dir>]
```

### Examples

```
# Create or update all resources described in a ksonnet application, and
# running in the 'dev' environment. Can be used in any subdirectory of the
# application.
ks apply dev

# Create or update resources described in a YAML file. Automatically picks up
# the cluster's location from '$KUBECONFIG'.
ks appy -f ./pod.yaml

# Create or update resources described in the JSON file. Changes are deployed
# to the cluster pointed at the 'dev' environment.
ks apply dev -f ./pod.json

# Update resources described in a YAML file, and running in cluster referred
# to by './kubeconfig'.
ks apply --kubeconfig=./kubeconfig -f ./pod.yaml

# Display set of actions we will execute when we run 'apply'.
ks apply dev --dry-run
```

### Options

```
      --as string                      Username to impersonate for the operation
      --certificate-authority string   Path to a cert. file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --create                         Create missing resources (default true)
      --dry-run                        Perform only read-only operations
  -V, --ext-str stringSlice            Values of external variables
      --ext-str-file stringSlice       Read external variable from a file
  -f, --file stringArray               Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)
      --gc-tag string                  Add this tag to updated objects, and garbage collect existing objects with this tag and not in config
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -J, --jpath stringSlice              Additional jsonnet library search path
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --resolve-images string          Change implementation of resolveImage native function. One of: noop, registry (default "noop")
      --resolve-images-error string    Action when resolveImage fails. One of ignore,warn,error (default "warn")
      --server string                  The address and port of the Kubernetes API server
      --skip-gc                        Don't perform garbage collection, even with --gc-tag
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

