## ks env add

Add a new environment to a ksonnet project

### Synopsis


Add a new environment to a ksonnet project. Names are restricted to not
include punctuation, so names like '../foo' are not allowed.

An environment acts as a sort of "named cluster", allowing for commands like
'ks apply dev', which applies the ksonnet application to the "dev cluster".
For more information on what an environment is and how they work, run 'help
env'.

Environments are represented as a hierarchy in the 'environments' directory of a
ksonnet application, and hence 'env add' will add to this directory structure.
For example, in the example below, there are two environments: 'default' and
'us-west/staging'. 'env add' will add a similar directory to this environment.

environments/
  default/           [Default generated environment]
    .metadata/
      k.libsonnet
      k8s.libsonnet
      swagger.json
    spec.json
    default.jsonnet
  us-west/
    staging/         [Example of user-generated env]
      .metadata/
        k.libsonnet
        k8s.libsonnet
        swagger.json
      spec.json      [This will contain the API server address of the environment and other environment metadata],
      staging.jsonnet

```
ks env add <env-name>
```

### Examples

```
  # Initialize a new staging environment at 'us-west'.
	# The environment will be setup using the current context in your kubecfg file. The directory
	# structure rooted at 'us-west' in the documentation above will be generated.
  ks env add us-west/staging

  # Initialize a new staging environment at 'us-west' with the namespace 'staging', using
  # the OpenAPI specification generated in the Kubernetes v1.7.1 build to generate 'ksonnet-lib'.
  ks env add us-west/staging --api-spec=version:v1.7.1 --namespace=staging

  # Initialize a new environment using the 'dev' context in your kubeconfig file.
  ks env add my-env --context=dev

  # Initialize a new environment using a server address.
  ks env add my-env --server=https://ksonnet-1.us-west.elb.amazonaws.com
```

### Options

```
      --api-spec string   Manually specify API version from OpenAPI schema, cluster, or Kubernetes version (default "version:v1.7.0")
```

### Options inherited from parent commands

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
  -v, --verbose count[=-1]             Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks env](ks_env.md)	 - Manage ksonnet environments

