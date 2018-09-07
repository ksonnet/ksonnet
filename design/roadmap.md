# Roadmap to 1.0

## Using ksonnet shouldn't require Jsonnet knowledge

* Users can initialize a ksonnet project, add components, configure parameters, customize their environment, and apply configurations without knowledge of Jsonnet.

## Component references

* Components can refer to parameters in other components.

## More flexible environments

* Users can use their ksonnet environments with multiple Kubernetes clusters. An environment is not directly tied to a Kubernetes cluster namespace.
* Users can reference the current namespace in components.
* Environments are hierarchical and can inherit their configurations.

## JSON and YAML support

* Users can add JSON, YAML, or Jsonnet file into `components/` and be able to apply them with `ks apply`. Only Jsonnet files can use parameters.
* To better facilitate migrations, users can use `ks` to convert JSON or YAML files to Jsonnet.

## Guidance for secrets management

* The ksonnet team will supply guidance on managing secrets in ksonnet. The team is currently investigating Bitnami's [sealed secrets](https://github.com/bitnami/sealed-secrets) as an approach.

## Editor integrations

* Extract the Jsonnet language server into its own project.
* In the VSCode extension:
  * Ensure the ksonnet lib locations are added to are added to the Jsonnet lib path.
  * Ensure dependency lib locations are added to the Jsonnet lib path.

## Docker Images

* Create a Dockerfile that will allow users to build Docker images for running ksonnet.
* Publish updated Docker images with each ksonnet release.