## ks registry remove

Remove a registry from the current ksonnet app

### Synopsis


The `remove` command allows added registries to be removed from your ksonnet app.

### Related Commands

* `ks registry add` — Add a registry to the current ksonnet app

### Syntax


```
ks registry remove <registry-name> [flags]
```

### Examples

```
# Remove a registry with the name 'databases'
ks registry remove databases
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --dir string        Ksonnet application root to use; Defaults to CWD (default "/Users/arve/go/src/github.com/ksonnet/ksonnet")
      --tls-skip-verify   Skip verification of TLS server certificates
  -v, --verbose count     Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks registry](ks_registry.md)	 - Manage registries for current project

