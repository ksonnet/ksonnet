## ks registry

Manage registries for current project

### Synopsis


Manage and inspect ksonnet registries.

Registries contain a set of versioned libraries that the user can install and
manage in a ksonnet project using the CLI. A typical library contains:

  1. A set of "parts", pre-fabricated API objects which can be combined together
     to configure a Kubernetes application for some task. For example, the Redis
     library may contain a Deployment, a Service, a Secret, and a
     PersistentVolumeClaim, but if the user is operating it as a cache, they may
     only need the first three of these.
  2. A set of "prototypes", which are pre-fabricated combinations of these
     parts, made to make it easier to get started using a library. See the
     documentation for 'ks prototype' for more information.

```
ks registry
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Configure your application to deploy to a Kubernetes cluster
* [ks registry describe](ks_registry_describe.md)	 - Describe a ksonnet registry
* [ks registry list](ks_registry_list.md)	 - List all registries known to the current ksonnet app.

