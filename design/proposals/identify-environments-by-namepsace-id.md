# Identify environments by namespace ID

Status: Pending

Version: Alpha

## Summary

Ksonnet needs to identify environments so it knows where to apply configurations. Identifying clusters by address is not a robust solution since the cluster API may differ depending on how you access the cluster.

This proposal aims to describe a process to identify environments by namespace ID rather than cluster address.

## Motivation

Environments are described by a file, `environments/env/spec.json` with the following format:

```json
{
  "server": "https://cluster:8443",
  "namespace": "default"
}
```

This method works for a single instance, but presents difficulties when an single environment needs to be applied a cluster that can be accessed with multiple addresses. (e.g. directly, or through a proxy)   To allow usage in more Kubernetes installations, a more flexible method for identifying when cluster/namespace an environment maps to is needed.

## Objectives

Describe a method to identify environments by namespace ID and a process for converting existing ksonnet installations.

### Goals

* Identify new environments by cluster namespace ID.
* Upgrade existing ksonnet installations to identifying environments by cluster namespace ID.
* Provide a command line method to set the namespace ID to a value.

### Non-Goals

* Managing ksonnet lib files associated with an environment.

## Proposal

1. Deprecate `environments/env/spec.json`
1. Store local mappings of cluster namespace to an environment. This may or may not be checked in based on the environment.

### User Stories

#### Environments will map to a cluster id

#### Instances of ksonnet can map a a cluster address

#### New environments will be created as a mapping

As a ksonnet user, when I create an instance of ksonnet, I want a mapping from a cluster namespace to an environment.

#### Existing environments will be converted to a mapping

As a ksonnet user, when I'm using an instance of ksonnet created before the new environment mapping, I want my environment to be converted to a mapping from a cluster namespace to an environment.

#### Manage environment from cli

As a ksonnet user, I would like to be able to maintain mapping from a cluster namespace to an environment using a command line tool.

## Unresolved Questions

* What part of the configuration will be checked into git vs what's local only?
* What will using this method as a part of GitOps pipeline look like?