## ks env

Manage ksonnet environments

### Synopsis


An environment acts as a sort of "named cluster", allowing for commands like
 `ks apply dev` , which applies the ksonnet application to the 'dev cluster'.
Additionally, environments allow users to cache data about the cluster it points
to, including data needed to run 'verify', and a version of ksonnet-lib that is
generated based on the flags the API server was started with (e.g., RBAC enabled
or not).

An environment contains no user-specific data (such as the private key
often contained in a kubeconfig file), and

Environments are represented as a hierarchy in the 'environments' directory of a
ksonnet application. For example, in the example below, there are two
environments: 'default' and 'us-west/staging'. Each contains a cached version of
 `ksonnet-lib` , and a `spec.json` that contains the server and server cert that
uniquely identifies the cluster.

    environments/
      default/           [Default generated environment]
        .metadata/
          k.libsonnet
          k8s.libsonnet
          swagger.json
        spec.json
		default.jsonnet
        params.libsonnet		
      us-west/
        staging/         [Example of user-generated env]
          .metadata/
            k.libsonnet
            k8s.libsonnet
            swagger.json
          spec.json      [This will contain the API server address of the environment and other environment metadata]
		  staging.jsonnet
          params.libsonnet

```
ks env
```

### Options

```
      --as string                      Username to impersonate for the operation
      --certificate-authority string   Path to a cert. file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files
* [ks env add](ks_env_add.md)	 - Add a new environment to a ksonnet project
* [ks env list](ks_env_list.md)	 - List all environments in a ksonnet project
* [ks env rm](ks_env_rm.md)	 - Delete an environment from a ksonnet project
* [ks env set](ks_env_set.md)	 - Set environment fields such as the name, server, and namespace.

