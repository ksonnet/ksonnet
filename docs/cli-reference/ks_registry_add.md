## ks registry add

Add a registry to the current ksonnet app

### Synopsis



The `add` command allows custom registries to be added to your ksonnet app.

A registry is uniquely identified by its:

1. Name
2. Version

Currently, only registries supporting the GitHub protocol can be added.

All registries must specify a unique name and URI where the registry lives.
Optionally, a version can be provided. If a version is not specified, it will
default to  `latest`.


### Related Commands

* `ks registry list` — List all registries known to the current ksonnet app.

### Syntax


```
ks registry add <registry-name> <registry-uri>
```

### Examples

```
# Add a registry with the name 'databases' at the uri 'github.com/example'
ks registry add databases github.com/example

# Add a registry with the name 'databases' at the uri 'github.com/example' and
# the version 0.0.1
ks registry add databases github.com/example --version=0.0.1
```

### Options

```
      --version string   Version of the registry to add
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks registry](ks_registry.md)	 - Manage registries for current project

