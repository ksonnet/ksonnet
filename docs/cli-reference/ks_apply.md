## ks apply

Apply local manifests to a remote cluster

### Synopsis


Use local manifest(s) to update (and optionally create) Kubernetes resources on the cluster.
This cluster is determined by the mandatory environment argument.
To control whether resources are created if they do not exist, use the `--create` flag (defaults to `true`).

An entire ksonnet application can be applied, or individual JSON, YAML, or Jsonnet files.

Note that this command needs to be run *within* a ksonnet app directory.

```
ks apply [env-name] [-f <file-or-dir>] [--dry-run]
```

### Examples

```
# Preview the set of actions that are executed by 'apply'.
ks apply dev --dry-run

# Update (but do not create) all resources described in a ksonnet
# application. Changes are deployed to the cluster pointed at the 'prod'
# environment. Can be used in any subdirectory of that application.
ks apply prod --create=false

# Create or update resources described in a JSON file. Changes are deployed
# to the cluster pointed at the 'dev' environment.
ks apply dev -f ./lib/pod.json

# Create or update resources described in a YAML file. Changes are deployed to the
# cluster pointed to by the current context of './kubeconfig'. This kubeconfig
# context overrides the 'default' environment.
ks apply default --kubeconfig=./kubeconfig -f ./lib/pod.yaml
```

### Options

```
      --as string                      Username to impersonate for the operation
      --certificate-authority string   Path to a cert. file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --create                         Option to create resources if they do not already exist on the cluster (default true)
      --dry-run                        Option to preview the list of operations without changing the cluster state
  -V, --ext-str stringSlice            Values of external variables
      --ext-str-file stringSlice       Read external variable from a file
  -f, --file stringArray               Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)
      --gc-tag string                  The specified tag is (1) added to all updated objects (2) used to identify what needs to be removed from the cluster (previously tagged objects that are no longer in the specified manifest)
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -J, --jpath stringSlice              Additional jsonnet library search path
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --resolve-images string          Change implementation of resolveImage native function. One of: noop, registry (default "noop")
      --resolve-images-error string    Action when resolveImage fails. One of ignore,warn,error (default "warn")
      --server string                  The address and port of the Kubernetes API server
      --skip-gc                        Option to skip garbage collection, even with --gc-tag specified
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

