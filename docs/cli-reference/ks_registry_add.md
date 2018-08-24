## ks registry add

Add a registry to the current ksonnet app

### Synopsis


The `add` command allows custom registries to be added to your ksonnet app,
provided that their file structures follow the appropriate schema. *You can look
at the `incubator` repo (https://github.com/ksonnet/parts/tree/master/incubator)
as an example.*

A registry is given a string identifier, which must be unique within a ksonnet application.

There are three supported registry protocols: **github**, **fs**, and **Helm**.

GitHub registries expect a path in a GitHub repository, and filesystem based
registries expect a path on the local filesystem.

During creation, all registries must specify a unique name and URI where the
registry lives. GitHub registries can specify a commit, tag, or branch to follow as part of the URI.

Registries can be overridden with `--override`.  Overridden registries
are stored in `app.override.yaml` and can be safely ignored using your
SCM configuration.

### Related Commands

* `ks registry list` — List all registries known to the current ksonnet app

### Syntax


```
ks registry add <registry-name> <registry-uri> [flags]
```

### Examples

```
# Add a registry with the name 'databases' at the uri 'github.com/example'
ks registry add databases github.com/example

# Add a registry with the name 'databases' at the uri
# 'github.com/org/example/tree/0.0.1/registry' (0.0.1 is the branch name)
ks registry add databases github.com/org/example/tree/0.0.1/registry

# Add a registry with a Helm Charts Repository uri
ks registry add helm-stable https://kubernetes-charts.storage.googleapis.com
```

### Options

```
  -h, --help       help for add
  -o, --override   Store in override configuration
```

### Options inherited from parent commands

```
      --tls-skip-verify      Skip verification of TLS server certificates
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks registry](ks_registry.md)	 - Manage registries for current project

