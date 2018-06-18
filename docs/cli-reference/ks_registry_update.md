## ks registry update

Update currently configured registries

### Synopsis


The `update` command updates a set of configured registries in your ksonnet app.
Unless a specific version is specified with `--version`, update will attempt to
fetch the lastest registry version matching the configured floating version specifer.

With `--version`, a specific version specifier (floating or concrete) can be set.

### Syntax


```
ks registry update [registry-name] [flags]
```

### Examples

```
# Update *all* registries to their latest matching versions
ks registry update

# Update a registry with the name 'databases' to version 0.0.2
ks registry update databases --version=0.0.1
```

### Options

```
  -h, --help             help for update
      --version string   Version to update registry to
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks registry](ks_registry.md)	 - Manage registries for current project

