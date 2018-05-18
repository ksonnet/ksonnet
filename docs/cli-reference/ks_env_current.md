## ks env current

Sets the current environment

### Synopsis


The `current` command lets you set the current ksonnet environment.

### Related Commands

* `ks env list` — List all environments in a ksonnet application

### Syntax


```
ks env current [--set <name> | --unset] [flags]
```

### Examples

```
#Update the current environment to 'us-west/staging'
ks env current --set us-west/staging

#Retrieve the current environment
ks env current

#Unset the current environment
ks env current --unset
```

### Options

```
  -h, --help         help for current
      --set string   Environment to set as current
      --unset        Unset current environment
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks env](ks_env.md)	 - Manage ksonnet environments

