## ks env list

List all environments in a ksonnet application

### Synopsis


The `list` command lists all of the available environments for the
current ksonnet app. Specifically, this will display the (1) *name*,
(2) *server*, and (3) *namespace* of each environment.

### Related Commands

* `ks env add` — Add a new environment to a ksonnet application
* `ks env set` — Set environment-specific fields (name, namespace, server)
* `ks env rm` — Delete an environment from a ksonnet application

### Syntax


```
ks env list [flags]
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format. Valid options: table|json
```

### Options inherited from parent commands

```
      --dir string        Ksonnet application root to use; Defaults to CWD
      --tls-skip-verify   Skip verification of TLS server certificates
  -v, --verbose count     Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks env](ks_env.md)	 - Manage ksonnet environments

