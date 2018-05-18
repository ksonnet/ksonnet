## ks env

Manage ksonnet environments

### Synopsis


An environment is a deployment target for your ksonnet app and its constituent
components. You can use ksonnet to deploy a given app to *multiple* environments,
such as `dev` and `prod`.

Intuitively, an environment acts as a sort of "named cluster", similar to a
Kubernetes context. (Running `ks env add --help` provides more detail
about the fields that you need to create an environment).

**All of this environment info is cached in local files**. Environments are
represented as a hierarchy in the `environments/` directory of a ksonnet app, like
'default' and 'us-west/staging' in the example below.

```
├── environments
│   ├── base.libsonnet
│   ├── default                      // Default generated environment ('ks init')
│   │   ├── .metadata
│   │   │   ├── k.libsonnet
│   │   │   ├── k8s.libsonnet
│   │   │   └── swagger.json
│   │   ├── main.jsonnet
│   │   ├── params.libsonnet
│   │   └── spec.json
│   └── us-west
│       └── staging                  // Example of user-generated env ('ks env add')
│           ├── .metadata
│           │   ├── k.libsonnet      // Jsonnet library with Kubernetes-compatible types and definitions
│           │   ├── k8s.libsonnet
│           │   └── swagger.json
│           ├── main.libsonnet       // Main file that imports all components (expanded on apply, delete, etc). Add environment-specific logic here.
│           ├── params.libsonnet     // Customize components *per-environment* here.
│           └── spec.json            // Contains the environment's API server address and namespace
```
----


```
ks env [flags]
```

### Options

```
  -h, --help   help for env
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks](ks.md)	 - Configure your application to deploy to a Kubernetes cluster
* [ks env add](ks_env_add.md)	 - Add a new environment to a ksonnet application
* [ks env current](ks_env_current.md)	 - Sets the current environment
* [ks env describe](ks_env_describe.md)	 - Describe an environment
* [ks env list](ks_env_list.md)	 - List all environments in a ksonnet application
* [ks env rm](ks_env_rm.md)	 - Delete an environment from a ksonnet application
* [ks env set](ks_env_set.md)	 - Set environment-specific fields (name, namespace, server)
* [ks env targets](ks_env_targets.md)	 - Set module targets for an environment
* [ks env update](ks_env_update.md)	 - Updates the libs for an environment

