## ks param

Manage parameters for ksonnet components

### Synopsis


Parameters are predetermined fields that can be used to customize ksonnet
components. Potential examples include replica count, component name, and deployment
image.

Parameters can also be defined on a per-environment basis. For example, this
allows a "development" environment to only run a single replica for component X,
whereas a production environment might run multiple replicas for component X, in
order to meet heavier production load demands.

Environments can be thought of as "named cluster contexts". For more information,
run `ks env --help` or see the concept documentation.

```
ks param
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files
* [ks param diff](ks_param_diff.md)	 - Display differences between the component parameters of two environments
* [ks param list](ks_param_list.md)	 - List all parameters for a component(s)
* [ks param set](ks_param_set.md)	 - Set component or environment parameters such as replica count or name

