## ks pkg install

Install a package (e.g. extra prototypes) for the current ksonnet app

### Synopsis



The `install` command caches a ksonnet library locally, and make it available
for use in the current ksonnet application. Enough info and metadata is recorded in
 `app.yaml` that new users can retrieve the dependency after a fresh clone of this app.

The library itself needs to be located in a registry (e.g. Github repo). By default,
ksonnet knows about two registries: *incubator* and *stable*, which are the release
channels for official ksonnet libraries.

### Related Commands

* `ks pkg list` — List all packages known (downloaded or not) for the current ksonnet app
* `ks prototype list` — List all locally available ksonnet prototypes

### Syntax


```
ks pkg install <registry>/<library>@<version>
```

### Examples

```

# Install an nginx dependency, based on the 'master' branch.
# In a ksonnet source file, this can be referenced as:
#     local nginx = import "incubator/nginx/nginx.libsonnet";
ks pkg install incubator/nginx@master

```

### Options

```
      --name string   Name to give the dependency, to use within the ksonnet app
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks pkg](ks_pkg.md)	 - Manage packages and dependencies for the current ksonnet application

