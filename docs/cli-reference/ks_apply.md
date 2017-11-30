## ks apply

Apply local Kubernetes manifests to remote clusters

### Synopsis



Update (or optionally create) Kubernetes resources on the cluster using your
local Kubernetes manifests. Use the `--create` flag to control whether
they are created if they do not exist (default: true).

The local Kubernetes manifests that are applied reside in your `components/`
directory. When applied, the manifests are fully expanded using the paremeters
of the specified environment.

By default, all manifests are applied. To apply a subset of manifests, use the
`--component` flag, as seen in the examples below.

### Related Commands

* `ks delete` — Delete the component manifests on your cluster

### Syntax


```
ks apply <env-name>
```

### Examples

```

# Create or update all resources described in a ksonnet application, and
# running in the 'dev' environment. Can be used in any subdirectory of the
# application.
#
# This is equivalent to applying all components in the 'components/' directory.
ks apply dev

# Create or update the single resource 'guestbook-ui' described in a ksonnet
# application, and running in the 'dev' environment. Can be used in any
# subdirectory of the application.
#
# This is equivalent to applying the component with the same file name (excluding
# the extension) 'guestbook-ui' in the 'components/' directory.
ks apply dev -c guestbook-ui

# Create or update the multiple resources, 'guestbook-ui' and 'nginx-depl'
# described in a ksonnet application, and running in the 'dev' environment. Can
# be used in any subdirectory of the application.
#
# This is equivalent to applying the component with the same file name (excluding
# the extension) 'guestbook-ui' and 'nginx-depl' in the 'components/' directory.
ks apply dev -c guestbook-ui -c nginx-depl

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
      --create                         Create missing resources (default true)
      --dry-run                        Perform only read-only operations
  -V, --ext-str stringSlice            Values of external variables
      --ext-str-file stringSlice       Read external variable from a file
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

