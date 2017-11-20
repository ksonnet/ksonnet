## ks diff

Display differences between manifests, based on environment or location (local or remote)

### Synopsis


Display differences between manifests, for any of the following cases:
1. Remote and local manifests for a single environment
2. Remote manifests for two separate environments
3. Local manifests for two separate environments
4. A remote manifest in one environment and a local manifest in another environment

Essentially, any two manifests can be compared, as long as their location ("local" or
"remote") and environment are specified.

Note that "remote" refers to what's actually running on a given Kubernetes API server.

Diff can be called on an entire ksonnet application, or individual JSON, YAML, or
Jsonnet files.

```
ks diff [<env1> [<env2>]] [-f <file-or-dir>]
```

### Examples

```
# Show diff between remote and local manifests for a single 'dev'
# environment. This diffs the entire ksonnet app, and can be used in any
# subdirectory of that ksonnet application.
ks diff dev

# Show diff between the remote resources running in two different ksonnet
# environments, 'us-west/dev' and 'us-west/prod'. This command covers
# all resources defined in the ksonnet app.
ks diff remote:us-west/dev remote:us-west/prod

# Show diff between local manifests in the 'us-west/dev' environment and
# remote resources in the 'us-west/prod' environment, for an entire ksonnet app
ks diff local:us-west/dev remote:us-west/prod

# For the resources defined in './components/redis.jsonnet', show the diff
# between what's in the local manifest and what's actually running in the 'dev'
# environment
ks diff dev -f ./components/redis.jsonnet
```

### Options

```
      --diff-strategy string          Diff strategy, "all" or "subset". (default "all")
  -V, --ext-str stringSlice           Values of external variables
      --ext-str-file stringSlice      Read external variable from a file
  -f, --file stringArray              Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)
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

