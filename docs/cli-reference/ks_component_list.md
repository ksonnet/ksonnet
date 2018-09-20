## ks component list

List known components

### Synopsis


The `list` command displays all known components.

### Syntax


```
ks component list [flags]
```

### Examples

```

# List all components
ks component list
```

### Options

```
  -h, --help            help for list
      --module string   Component module
  -o, --output string   Output format. Valid options: table|json
```

### Options inherited from parent commands

```
      --dir string        Ksonnet application root to use; Defaults to CWD
      --tls-skip-verify   Skip verification of TLS server certificates
  -v, --verbose count     Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks component](ks_component.md)	 - Manage ksonnet components

