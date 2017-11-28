## ks diff

Display differences between server and local config, or server and server config

### Synopsis


Display differences between server and local configuration, or server and server
configurations.

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.

```
ks diff [<env1> [<env2>]] [-f <file-or-dir>]
```

### Examples

```
# Show diff between resources described in a the local 'dev' environment
# specified by the ksonnet application and the remote cluster referenced by
# the same 'dev' environment. Can be used in any subdirectory of the application.
ksonnet diff dev

# Show diff between resources at remote clusters. This requires ksonnet
# application defined environments. Diff between the cluster defined at the
# 'us-west/dev' environment, and the cluster defined at the 'us-west/prod'
# environment. Can be used in any subdirectory of the application.
ksonnet diff remote:us-west/dev remote:us-west/prod

# Show diff between resources at a remote and a local cluster. This requires
# ksonnet application defined environments. Diff between the cluster defined
# at the 'us-west/dev' environment, and the cluster defined at the
# 'us-west/prod' environment. Can be used in any subdirectory of the
# application.
ksonnet diff local:us-west/dev remote:us-west/prod

# Show diff between resources described in a YAML file and the cluster
# referenced in '$KUBECONFIG'.
ks diff -f ./pod.yaml

# Show diff between resources described in a JSON file and the cluster
# referenced by the environment 'dev'.
ks diff dev -f ./pod.json

# Show diff between resources described in a YAML file and the cluster
# referred to by './kubeconfig'.
ks diff --kubeconfig=./kubeconfig -f ./pod.yaml
```

### Options

```
  -c, --component stringArray         Name of a specific component (multiple -c flags accepted, allows YAML, JSON, and Jsonnet)
      --diff-strategy string          Diff strategy, all or subset. (default "all")
  -V, --ext-str stringSlice           Values of external variables
      --ext-str-file stringSlice      Read external variable from a file
  -J, --jpath stringSlice             Additional jsonnet library search path
      --resolve-images string         Change implementation of resolveImage native function. One of: noop, registry (default "noop")
      --resolve-images-error string   Action when resolveImage fails. One of ignore,warn,error (default "warn")
  -A, --tla-str stringSlice           Values of top level arguments
      --tla-str-file stringSlice      Read top level argument from a file
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files

