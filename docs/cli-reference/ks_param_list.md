## ks param list

List all parameters for a component(s)

### Synopsis


List all component parameters or environment parameters.

This command displays all parameters for the specified component. If a
component is not specified, parameters for *all* components are listed.

Furthermore, parameters can be listed on a per-environment basis.


```
ks param list <component-name>
```

### Examples

```
# List all component parameters
ks param list

# List all parameters for the component "guestbook"
ks param list guestbook

# List all parameters for the environment "dev"
ks param list --env=dev

# List all parameters for the component "guestbook" in the environment "dev"
ks param list guestbook --env=dev
```

### Options

```
      --env string   Name of the environment to list parameters for
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks param](ks_param.md)	 - Manage parameters for ksonnet components

