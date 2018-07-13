## ks env set

Set environment-specific fields (name, namespace, server)

### Synopsis


The `set` command lets you change the fields of an existing environment.
You can currently only update your environment's name.

Note that changing the name of an environment will also update the corresponding
directory structure in `environments/`.

### Related Commands

* `ks env list` — List all environments in a ksonnet application

### Syntax


```
ks env set <env-name> [flags]
```

### Examples

```
# Update the name of the environment 'us-west/staging'.
# Updating the name will update the directory structure in 'environments/'.
ks env set us-west/staging --name=us-east/staging

# Setting k8s API version for an environment
ks env set us-west/staging --api-spec=version:v1.8.0

# Updating the server
ks env set us-west/staging --server=https://192.168.99.100:8443

```

### Options

```
      --api-spec string    Kubernetes version for environment
  -h, --help               help for set
      --name string        Name used to uniquely identify the environment. Must not already exist within the ksonnet app
      --namespace string   Namespace for environment
      --server string      Cluster server for environment
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks env](ks_env.md)	 - Manage ksonnet environments

