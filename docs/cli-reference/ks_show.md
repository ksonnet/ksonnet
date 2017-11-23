## ks show

Show expanded manifests for a specific environment.

### Synopsis


Show expanded manifests (resource definitions) for a specific environment. Jsonnet manifests,
each defining a ksonnet component, are expanded into their JSON or YAML equivalents (YAML is the default).
Any parameters for a component are resolved based on environment-specific values.

When NO component is specified via the`-c`flag, this command expands all of the files in the `components/` directory into a list of resource definitions. This is the YAML version
of what gets deployed to your cluster with `ks apply <env>`.

When a component IS specified via the`-c`flag, this command only expands the manifest for that
particular component.

```
ks show <env> [-c <component-filename>]
```

### Examples

```
# Show all of the components for the 'dev' environment, in YAML
# (In other words, expands all manifests in the components/ directory)
ks show dev

# Show a single component from the 'prod' environment, in JSON
ks show prod -c redis -o json

# Show multiple components from the 'dev' environment, in YAML
ks show dev -c redis -c nginx-server

```

### Options

```
  -V, --ext-str stringSlice           Values of external variables
      --ext-str-file stringSlice      Read external variable from a file
  -f, --file stringArray              Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)
  -o, --format string                 Output format.  Supported values are: json, yaml (default "yaml")
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

