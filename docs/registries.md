# ksonnet Registries

ksonnet registries allow for sharing of code used to build ksonnet components. Currently, ksonnet supports two types of registries: `github` and `fs`. 

## GitHub Registries

`github` registries are hosted on GitHub. 

## Fs Registries

`fs` registries are hosted on the local filesystem. They can be used when developing a registry. 


## Creating a Registry

Registries require a `registry.yaml` file. This file contains configuration for the registry and pointers to its components. 

An example `registry.yaml`:

```yaml
apiVersion: 0.1.0
kind: ksonnet.io/registry
libraries:
  scheduling:
    path: scheduling
```

In this example, the registry contains a single library, `scheduling`, which lives in directory `scheduling`. This path is relative to the directory that contains `registry.yaml`. 

