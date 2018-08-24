## ks env update

Updates the libs for an environment

### Synopsis


The `update` command updates packages for an environment.

### Related Commands

* `ks env list` — List all environments in a ksonnet application
* `ks env add` — Add a new environment to a ksonnet application
* `ks env set` — Set environment-specific fields (name, namespace, server)
* `ks delete` — Delete all the app components running in an environment (cluster)

### Syntax


```
ks env update <env-name> [flags]
```

### Examples

```

# Update the environment 'us-west/staging' packages.
ks env update us-west/staging
```

### Options

```
  -h, --help   help for update
```

### Options inherited from parent commands

```
      --tls-skip-verify      Skip verification of TLS server certificates
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks env](ks_env.md)	 - Manage ksonnet environments

