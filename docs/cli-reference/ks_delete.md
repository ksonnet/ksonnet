## ks delete

Delete Kubernetes resources described in local config

### Synopsis


Delete Kubernetes resources from a cluster, as described in the local
configuration.

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.

```
ks delete [env-name] [-f <file-or-dir>]
```

### Examples

```
  # Delete all resources described in a ksonnet application, from the 'dev'
  # environment. Can be used in any subdirectory of the application.
  ks delete dev

  # Delete resources described in a YAML file. Automatically picks up the
  # cluster's location from '$KUBECONFIG'.
  ks delete -f ./pod.yaml

  # Delete resources described in the JSON file from the 'dev' environment. Can
  # be used in any subdirectory of the application.
  ks delete dev -f ./pod.json

  # Delete resources described in a YAML file, and running in the cluster
  # specified by the current context in specified kubeconfig file.
  ks delete --kubeconfig=./kubeconfig -f ./pod.yaml
```

### Options

```
      --as string                      Username to impersonate for the operation
      --certificate-authority string   Path to a cert. file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
  -V, --ext-str stringSlice            Values of external variables
      --ext-str-file stringSlice       Read external variable from a file
  -f, --file stringArray               Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)
      --grace-period int               Number of seconds given to resources to terminate gracefully. A negative value is ignored (default -1)
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

