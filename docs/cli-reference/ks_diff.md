## ks diff

Compare manifests, based on environment or location (local or remote)

### Synopsis


The `diff` command displays standard file diffs, and can be used to compare manifests
based on *environment* or location ('local' ksonnet app manifests or what's running
on a 'remote' server).

Using this command, you can compare:

1. *Remote* and *local* manifests for a single environment
2. *Remote* manifests for two separate environments
3. *Local* manifests for two separate environments
4. A *remote* manifest in one environment and a *local* manifest in another environment

To see the official syntax, see the examples below. Make sure that your $KUBECONFIG
matches what you've defined in environments.

When NO component is specified (no `-c` flag), this command diffs all of
the files in the `components/` directory.

When a component IS specified via the `-c` flag, this command only checks
the manifest for that particular component.

### Related Commands

* `ks param diff` — Display differences between the component parameters of two environments

### Syntax


```
ks diff <location1:env1> [location2:env2] [flags]
```

### Examples

```

# Show diff between remote and local manifests for a single 'dev' environment.
# This command diffs *all* components in the ksonnet app, and can be used in any
# of that app's subdirectories.
ks diff remote:dev local:dev

# Shorthand for the previous command (remote 'dev' and local 'dev')
ks diff dev

# Show diff between the remote resources running in two different ksonnet environments
# 'us-west/dev' and 'us-west/prod'. This command diffs all resources defined in
# the ksonnet app.
ks diff remote:us-west/dev remote:us-west/prod

# Show diff between local manifests in the 'us-west/dev' environment and remote
# resources in the 'us-west/prod' environment, for an entire ksonnet app
ks diff local:us-west/dev remote:us-west/prod

# Show diff between what's in the local manifest and what's actually running in the
# 'dev' environment, but for the Redis component ONLY
ks diff dev -c redis

```

### Options

```
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
  -V, --ext-str stringSlice            Values of external variables
      --ext-str-file stringSlice       Read external variable from a file
  -h, --help                           help for diff
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -J, --jpath stringSlice              Additional jsonnet library search path
      --kubeconfig string              Path to a kubeconfig file. Alternative to env var $KUBECONFIG.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
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

* [ks](ks.md)	 - Configure your application to deploy to a Kubernetes cluster

