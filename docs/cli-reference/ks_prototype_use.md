## ks prototype use

Use the specified prototype to generate a component manifest

### Synopsis



The `generate` command (aliased from `prototype use`) generates Kubernetes-
compatible, Jsonnet manifests for components in your ksonnet app. Each prototype
corresponds to a single manifest in the `components/` directory. This manifest
can define one or more Kubernetes resources.

1. The first argument, the **prototype name**, can either be fully qualified
(e.g. `io.ksonnet.pkg.single-port-service`) or a partial match (e.g. `service`).
If using a partial match, note that any ambiguity in resolving the name will
result in an error.

2. The second argument, the **component name**, determines the filename for the
generated component manifest. For example, the following command will expand
template `io.ksonnet.pkg.single-port-deployment` and place it in the
file `components/nginx-depl.jsonnet` . Note that by default ksonnet will
expand prototypes into Jsonnet files.

       ks prototype use io.ksonnet.pkg.single-port-deployment nginx-depl \
         --name=nginx                                                    \
         --image=nginx

3. Prototypes can be further customized by passing in **parameters** via additional
command line flags, such as `--name` and `--image` in the example above. Note that
different prototypes support their own unique flags.

### Related Commands

* `ks apply` — Apply your component manifests to a cluster
* `ks param set` — Change the values of an existing component

### Syntax


```
ks prototype use <prototype-name> <componentName> [type] [parameter-flags]
```

### Examples

```

# Instantiate prototype 'io.ksonnet.pkg.single-port-deployment', using the
# 'nginx' image. The expanded prototype is placed in
# 'components/nginx-depl.jsonnet'.
ks prototype use io.ksonnet.pkg.single-port-deployment nginx-depl \
  --name=nginx                                                    \
  --image=nginx

# Instantiate prototype 'io.ksonnet.pkg.single-port-deployment' using the
# suffix, 'deployment'. The expanded prototype is again placed in
# 'components/nginx-depl.jsonnet'. NOTE: if you have imported another
# prototype with this suffix, this may resolve ambiguously for you.
ks prototype use deployment nginx-depl \
  --name=nginx                         \
  --image=nginx
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks prototype](ks_prototype.md)	 - Instantiate, inspect, and get examples for ksonnet prototypes

