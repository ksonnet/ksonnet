## ks env set

Set environment-specific fields (name, namespace, server)

### Synopsis



The `set` command lets you change the fields of an existing environment.
You can update any of your environment's (1) name (2) namespace and
(3) server (cluster URI).

Note that changing the name of an environment will also update the corresponding
directory structure in `environments/`.

### Related Commands

* `ks env list` — List all environments in a ksonnet application

### Syntax


```
ks env set <env-name>
```

### Examples

```
# Update the API server address of the environment 'us-west/staging'.
ks env set us-west/staging --server=http://example.com

# Update the namespace of the environment 'us-west/staging'.
ks env set us-west/staging --namespace=staging

# Update both the name and the server of the environment 'us-west/staging'.
# Updating the name will update the directory structure in 'environments/'.
ks env set us-west/staging --server=http://example.com --name=us-east/staging

# Update the API server address of the environment 'us-west/staging' based on the
# server in the 'staging-west' context of your kubeconfig file.
ks env set us-west/staging --context=staging-west
```

### Options

```
      --name string   Name used to uniquely identify the environment. Must not already exist within the ksonnet app
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

