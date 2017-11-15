## ks param set

Set component or environment parameters such as replica count or name

### Synopsis


"Set component or environment parameters such as replica count or name.

Parameters are set individually, one at a time. If you require customization of
more fields, we suggest that you modify your ksonnet project's
'components/params.libsonnet' file directly. Likewise, for greater customization
of environment parameters, we suggest modifying the
'environments/:name/params.libsonnet' file.


```
ks param set <component-name> <param-key> <param-value>
```

### Examples

```
  # Updates the replica count of the 'guestbook' component to 4.
  ks param set guestbook replicas 4

  # Updates the replica count of the 'guestbook' component to 2 for the environment
  # 'dev'
  ks param set guestbook replicas 2 --env=dev
```

### Options

```
      --env string   Specify environment to set parameters for
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks param](ks_param.md)	 - Manage ksonnet component parameters

