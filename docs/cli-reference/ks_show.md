## ks show

Show expanded resource definitions

### Synopsis


Show expanded resource definitions

```
ks show [<env>|-f <file-or-dir>]
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

