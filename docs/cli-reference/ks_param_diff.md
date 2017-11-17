## ks param diff

Display differences between the component parameters of two environments

### Synopsis


Pretty print differences between the component parameters of two environments.

By default, the diff is performed for all components. Diff-ing for a single component
is supported via a component flag.


```
ks param diff <env1> <env2>
```

### Examples

```
# Diff between all component parameters for environments 'dev' and 'prod'
ks param diff dev prod

# Diff only between the parameters for the 'guestbook' component for environments 'dev' and 'prod'
ks param diff dev prod --component=guestbook
```

### Options

```
      --component string   Base name of the component file to diff against
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks param](ks_param.md)	 - Manage parameters for ksonnet components

