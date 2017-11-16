## ks generate

Expand prototype, place in components/ directory of ksonnet app

### Synopsis


Expand prototype uniquely identified by (possibly partial) 'prototype-name',
filling in parameters from flags, and placing it into the file
'components/componentName', with the appropriate extension set. For example, the
following command will expand template 'io.ksonnet.pkg.single-port-deployment'
and place it in the file 'components/nginx-depl.jsonnet' (since by default
ksonnet will expand templates as Jsonnet).

  ks prototype use io.ksonnet.pkg.single-port-deployment nginx-depl \
    --name=nginx                                                         \
    --image=nginx

Note also that 'prototype-name' need only contain enough of the suffix of a name
to uniquely disambiguate it among known names. For example, 'deployment' may
resolve ambiguously, in which case 'use' will fail, while 'deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.single-port-deployment'.

```
ks generate <prototype-name> <component-name> [type] [parameter-flags]
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files

