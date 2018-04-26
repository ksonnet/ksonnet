# Overriding application configuration

ksonnet allows users to provide local overrides to application configurations. The overrides are stored in `app.override.yaml`. This feature gives users the ability to have local configurations that are not checked into Git.

## Overriding environments

To allow users more flexibility, environments can be overridden. If there is an environment, `dev`, specified in `app.yaml`, A user can specify local parameters:

```
ks env add dev --namespace local-namespace --server http://196.168.100.99 --override
```

This configuration will be used when present and not change the configuration present in the Git repository.

## Overriding registries

Add an override registry with the following syntax: `ks registry add <registry-name> <registry-uri> --override`. This configuration will be added to the local configuration that will not be checked into Git by default. If you specify a registry that already exists, the local version will be preferred. 

