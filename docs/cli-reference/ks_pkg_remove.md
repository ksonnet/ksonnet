## ks pkg remove

Remove a package from the app or environment scope

### Synopsis


The `remove` command removes a reference to a ksonnet library.  The reference can either be
global or scoped to an environment. If the last reference to a library version is removed, the cached
files will be removed as well.

### Syntax


```
ks pkg remove <registry>/<library> [flags]
```

### Examples

```

# Remove an nginx dependency
ks pkg remove incubator/nginx

# Remove an nginx dependency from the stage environment
ks pkg remove incubator/nginx --env stage

```

### Options

```
      --env string   Environment to remove package from (optional)
  -h, --help         help for remove
```

### Options inherited from parent commands

```
      --tls-skip-verify      Skip verification of TLS server certificates
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks pkg](ks_pkg.md)	 - Manage packages and dependencies for the current ksonnet application

