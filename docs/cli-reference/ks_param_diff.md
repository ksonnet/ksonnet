## ks param diff

Display differences between the component parameters of two environments

### Synopsis


Pretty prints differences between the component parameters of two environments.

A component flag is accepted to diff against a single component. By default, the
diff is performed against all components.


```
ks param diff <env1> <env2>
```

### Examples

```
# Diff between the component parameters on environments 'dev' and 'prod'
ks param diff dev prod

# Diff between the component 'guestbook' on environments 'dev' and 'prod'
ks param diff dev prod --component=guestbook
```

### Options

```
      --component string   Specify the component to diff against
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks param](ks_param.md)	 - Manage ksonnet component parameters

