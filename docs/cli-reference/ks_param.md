## ks param

Manage ksonnet component parameters

### Synopsis


Parameters are the customizable fields defining ksonnet components. For
example, replica count, component name, or deployment image.

Parameters are also able to be defined separately across environments. Meaning,
this supports features to allow a "development" environment to only run a
single replication instance for it's components, whereas allowing a "production"
environment to run more replication instances to meet heavier production load
demands.

Environments are ksonnet "named clusters". For more information on environments,
run:
  ks env --help


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

