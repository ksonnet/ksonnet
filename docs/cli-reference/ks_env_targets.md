## ks env targets

Set target modules for an environment

### Synopsis


The `targets` command selects one or more modules to be applied by an
environment. The default environment target is the root module, `/`.

Changing targets for an environment will require specifying all desired modules including the root module.


```
ks env targets [flags]
```

### Examples

```

# Create a new module
ks module create db

# Generate a component and specify the module
ks generate redis-stateless redis --module db

# Change the default environment target from / to db
# The targets are tracked in app.yaml
ks env targets default --module db
```

### Options

```
  -h, --help             help for targets
      --module strings   Component modules to include
  -o, --override         Set targets in environment as override
```

### Options inherited from parent commands

```
      --dir string        Ksonnet application root to use; Defaults to CWD
      --tls-skip-verify   Skip verification of TLS server certificates
  -v, --verbose count     Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks env](ks_env.md)	 - Manage ksonnet environments

