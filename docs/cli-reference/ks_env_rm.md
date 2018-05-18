## ks env rm

Delete an environment from a ksonnet application

### Synopsis


The `rm` command deletes an environment from a ksonnet application. This is
the same as removing the `<env-name>` environment directory and all files
contained. All empty parent directories are also subsequently deleted.

NOTE: This does *NOT* delete the components running in `<env-name>`. To do that, you
need to use the `ks delete` command.

### Related Commands

* `ks env list` — List all environments in a ksonnet application
* `ks env add` — Add a new environment to a ksonnet application
* `ks env set` — Set environment-specific fields (name, namespace, server)
* `ks delete` — Delete all the app components running in an environment (cluster)

### Syntax


```
ks env rm <env-name> [flags]
```

### Examples

```

# Remove the directory 'environments/us-west/staging' and all of its contents.
# This will also remove the parent directory 'us-west' if it is empty.
ks env rm us-west/staging
```

### Options

```
  -h, --help       help for rm
  -o, --override   Remove the overridden environment
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks env](ks_env.md)	 - Manage ksonnet environments

