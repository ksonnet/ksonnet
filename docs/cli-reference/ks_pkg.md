## ks pkg

Manage packages and dependencies for the current ksonnet application

### Synopsis


Packages allow you to distribute and reuse third-party code in any ksonnet *application*.
Any package can be installed if it has a valid combination of:
1. Registry (similar to a container registry, this is a repository of ksonnet libraries)
2. Library (contains prototypes and helper jsonnet libraries)
3. Version (a unique identifier for a release of the library, e.g. a git commit SHA)

For more information, see the concept documentation, or related info with `ks registry --help`.

```
ks pkg
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files
* [ks pkg describe](ks_pkg_describe.md)	 - Describe a ksonnet package
* [ks pkg install](ks_pkg_install.md)	 - Install a package as a dependency in the current ksonnet application
* [ks pkg list](ks_pkg_list.md)	 - Lists information about all dependencies known to the current ksonnet app

