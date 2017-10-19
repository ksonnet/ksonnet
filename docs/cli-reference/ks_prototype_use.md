## ks prototype use

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

Note that if we were to specify to expand the template as JSON or YAML, we would
generate a file with a '.json' or '.yaml' extension, respectively. See examples
below for an example of how to do this.

Note also that 'prototype-name' need only contain enough of the suffix of a name
to uniquely disambiguate it among known names. For example, 'deployment' may
resolve ambiguously, in which case 'use' will fail, while 'deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.single-port-deployment'.

```
ks prototype use <prototype-name> <componentName> [type] [parameter-flags]
```

### Examples

```
  # Instantiate prototype 'io.ksonnet.pkg.single-port-deployment', using the
  # 'nginx' image. The expanded prototype is placed in
  # 'components/nginx-depl.jsonnet'.
  ks prototype use io.ksonnet.pkg.prototype.simple-deployment nginx-depl \
    --name=nginx                                                              \
    --image=nginx

  # Instantiate prototype 'io.ksonnet.pkg.single-port-deployment' using the
  # unique suffix, 'deployment'. The expanded prototype is again placed in
  # 'components/nginx-depl.jsonnet'. See introduction of help message for more
  # information on how this works. Note that if you have imported another
  # prototype with this suffix, this may resolve ambiguously for you.
  ks prototype use deployment nginx-depl \
    --name=nginx                              \
    --image=nginx

  # Instantiate prototype 'io.ksonnet.pkg.single-port-deployment' as YAML,
  # placing the result in 'components/nginx-depl.yaml. Note that some templates
  # do not have a YAML or JSON versions.
  ks prototype use deployment nginx-depl yaml \
    --name=nginx                              \
    --image=nginx
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks prototype](ks_prototype.md)	 - Instantiate, inspect, and get examples for ksonnet prototypes

