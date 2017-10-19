## ks prototype preview

Expand prototype, emitting the generated code to stdout

### Synopsis


Expand prototype uniquely identified by (possibly partial)
'prototype-name', filling in parameters from flags, and emitting the generated
code to stdout.

Note also that 'prototype-name' need only contain enough of the suffix of a name
to uniquely disambiguate it among known names. For example, 'deployment' may
resolve ambiguously, in which case 'use' will fail, while 'deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.single-port-deployment'.

```
ks prototype preview <prototype-name> [type] [parameter-flags]
```

### Examples

```
  # Preview prototype 'io.ksonnet.pkg.single-port-deployment', using the
  # 'nginx' image, and port 80 exposed.
  ks prototype preview io.ksonnet.pkg.prototype.simple-deployment \
    --name=nginx                                                       \
    --image=nginx

  # Preview prototype using a unique suffix of an identifier. See
  # introduction of help message for more information on how this works.
  ks prototype preview simple-deployment \
    --name=nginx                              \
    --image=nginx

  # Preview prototype 'io.ksonnet.pkg.single-port-deployment' as YAML,
  # placing the result in 'components/nginx-depl.yaml. Note that some templates
  # do not have a YAML or JSON versions.
  ks prototype preview deployment nginx-depl yaml \
    --name=nginx                                       \
    --image=nginx
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks prototype](ks_prototype.md)	 - Instantiate, inspect, and get examples for ksonnet prototypes

