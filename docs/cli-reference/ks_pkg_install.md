## ks pkg install

Install a package as a dependency in the current ksonnet application

### Synopsis


Cache a ksonnet library locally, and make it available for use in the current
ksonnet project. This particularly means that we record enough information in
'app.yaml' for new users to retrieve the dependency after a fresh clone of the
app repository.

For example, inside a ksonnet application directory, run:

  ks pkg install incubator/nginx@v0.1

This can then be referenced in a source file in the ksonnet project:

  local nginx = import "kspkg://nginx";

By default, ksonnet knows about two registries: incubator and stable, which are
the release channels for official ksonnet libraries. Additional registries can
be added with the 'ks registry' command.

Note that multiple versions of the same ksonnet library can be cached and used
in the same project, by explicitly passing in the '--name' flag. For example:

  ks pkg install incubator/nginx@v0.1 --name nginxv1
  ks pkg install incubator/nginx@v0.2 --name nginxv2

With these commands, a user can 'import "kspkg://nginx1"', and
'import "kspkg://nginx2"' with no conflict.

```
ks pkg install <registry>/<library>@<version>
```

### Options

```
      --name string   Name to give the dependency
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks pkg](ks_pkg.md)	 - Manage packages and dependencies for the current ksonnet project

