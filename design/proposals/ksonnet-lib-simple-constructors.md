# Deprecate ksonnet lib custom constructors

Status: Pending

Version: Alpha

## Abstract

We want to deprecate the custom resource constructors (described [below](#use-cases)) in [ksonnet-lib](https://github.com/ksonnet/ksonnet-lib). Instead of the custom constructors, developers will be encouraged to use the builder pattern to construct their objects. Constructors can be reintroduced in hand rolled libraries upstream.

## Motivation

* `ksonnet-lib` includes constructors for popular resources. In a few cases, these constructors are incomplete or include items which may not be required. The incompleteness and inconsistencies are a barrier to adoption because developers do not know how to get started in a productive way.
* `ksonnet-lib` should be a direct translation of the OpenAPI swagger.json.

## Goal

* Developers should be able to construct objects using the builder pattern rather than constructors.
* Upstream projects should have a clean slate to build abstraction upon.

## Non Goals

The following are related but not the goals of this specific proposal

* Create upstream projects that implement constructors (e.g. v1 Deployment)

## Proposal

The change would be introducing a new constructor for all resources.

```jsonnet
group:: {
  v1:: {
    local apiVersion = {apiVersion: "group/v1"},
    kind:: {
      local kind = {kind: "Kind"},
      init() :: apiVersion + kind,
    },
  },
}
```

Many of the existing constructors adhere to this format. This change would make sure all resources have a constructor that only includes the `apiVersion` and `kind`.

## User Experience

### Use Cases

To deprecate the existing the existing custom constructors a new constructor will be introduced, `init()`. Currently, to construct a v1beta2 Deployment, you have to do the following:

```jsonnet
local params = ...;
local deployment = k.apps.v1beta2.deployment;
local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
local containerPort = container.portsType;
local labels = {app: params.name};
local targetPort = params.containerPort;

local appDeployment = deployment
  .new(
    params.name,
    params.replicas,
    container
      .new(params.name, params.image)
      .withPorts(containerPort.new(targetPort)),
    labels
);
```

This snippet illustrates a few issues:

* The constructor requires items that aren't required by the spec.
* The constructor requires a specific order.

Instead of the custom constructor, instead you can construct using init using the following:

```jsonnet
local params = ...;
local deployment = k.apps.v1beta2.deployment;
local containersType = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
local containerPort = container.portsType;
local labels = {app: params.name};
local targetPort = params.containerPort;

local port = containerPort
  .init()
  .withContainerPort(targetPort)

local container = containersType
  .init()
  .withName(params.name)
  .withImage(params.image)
  .withPorts(port)

local appDeployment = deployment
  .init()
  .mixin.metadata.withName(params.name)
  .mixin.spec.withReplicas(params.replicas)
  .mixin.spec.template.spec.withContainers(containers)
  .mixin.spec.template.metadata.withLabels(labels);
```

### Backwards compatibility

This change will not immediately impact any of the current usages of `ksonnet-lib`. To begin, the inclusion of `init()` constructors will be the only change. Only after a period of time (at most two releases of ksonnet) will existing constructors be removed.

## Alternatives considered

Instead of introducing new constructors, missing constructors can be be added to `ksonnet-lib`. Also, resources with incomplete, missing, or extra required items could fixed as well. To accomplish this, more custom logic would need to be inserted in the `ksonnet-lib` generator. The custom logic has the potential to make generator more complex and that is something we want to avoid.
