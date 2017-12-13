# ksonnet Roadmap

This document captures open questions and features for ksonnet.
*Note that the ordering of items is unrelated to their priority or order of completion.*

Improvements are planned in the following areas:

* Dependency management
   * [Support third-party registries (beyond Github protocol)](#support-third-party-registries-beyond-github-protocol)
   * [Easily generate third-party registries](#easily-generate-third-party-registries)
   * [Lazily download and update dependencies (`ks install`)](#lazily-download-and-update-dependencies-ks-install)
* Component workflow
   * [Expose "environment metadata" to prototypes](#expose-environment-metadata-to-prototypes)
   * [Support JSON and YAML](#support-json-and-yaml)
   * [Reference components in other components](#reference-components-in-other-components)
   * [Establish clear `ks apply` semantics](#establish-clear-ks-apply-semantics)
   * [Cleanly remove components](#cleanly-remove-components)   
   * [Encryption support for secrets](#encryption-support-for-secrets)
* Development tooling
   * [Improve output of `ks diff`](#improve-output-of-ks-diff)
   * [Show errors in app files (`ks lint`)](#show-errors-in-app-files-ks-lint)
   * [Update VSCode extension](#update-vscode-extension)
* Under the hood
  * [Improve `ksonnet-lib` generation](#improve-ksonnet-lib-generation)
  * [Refactor common code into `ksonnet/client`](#refactor-common-code-into-ksonnetclient)


## Support third-party registries (beyond Github protocol)

*Planned for 0.9.0* ([#232](https://github.com/ksonnet/ksonnet/issues/232))

Currently ksonnet allows you to `ks generate` prototypes from three sources:
* **[Standard prototypes]** From ksonnet code itself (see [`systemPrototypes.go`](https://github.com/ksonnet/ksonnet/blob/master/prototype/systemPrototypes.go))
* **[Standard prototypes]** From the [`incubator`](https://github.com/ksonnet/parts/tree/master/incubator) registry in the `ksonnet/parts` Github repo
* **[Custom prototypes]** From Github directories that conform to the registry spec (see [docs](https://ksonnet.io/docs/cli-reference#ks-registryks-registry-add) for `ks registry add`)

The last option (`ks registry add`) allows users to write and share their own prototypes. *However, this feature is currently limited to Github.* Users may want more flexibility, e.g. a local-filesystem-based registry for development purposes.

**`ks registry add` functionality should be expanded to support other common protocols (local filesystem, S3, etc).** The only requirement thereafter will be having the proper YAML metadata files (e.g. `registry.yaml` in the root).

## Easily generate third-party registries

*Planned for 0.9.0* ([ #234](https://github.com/ksonnet/ksonnet/issues/234))

Users can create custom, third-party registries, but the `ks registry add` command will only recognize them if they adhere to the registry spec. In other words, they need to have the appropriate YAML files and directory structure.

**A new command (e.g. `ks registry create`) which autogenerates registry scaffolding files, may make this easier and less error prone.**

## Lazily download and update dependencies (`ks install`)

*Planned for 0.9.0* ([#217](https://github.com/ksonnet/ksonnet/issues/217), [#237](https://github.com/ksonnet/ksonnet/issues/237))

Currently, `ks init` fails when the user doesn't have an internet connection, because it relies on two external dependencies:
* The `k.libsonnet` for the `default` environment
* The `incubator` registry

To avoid this issue, **ksonnet needs to lazily initialize environments and registries**, rather than trying to set these up right away.

**This plays into the idea of a `ks install` command, which will cut down on the number of version-controlled files for a given ksonnet app**. ksonnet dependencies will be `.gitignore`-ed. *Only their metadata* (e.g. versions) will be tracked.

After a ksonnet app is downloaded (e.g. via `git clone`), `ks install` can be run to download and "flesh out" the necessary environments, registries, and packages. This is similar to `npm install`.

**An associated command (e.g. `ks update`) should allow users to download the latest versions of a dependency.**

## Expose "environment metadata" to prototypes

*Planned for 0.9.0* ([#222](https://github.com/ksonnet/ksonnet/issues/222))

Nearly all ksonnet prototypes expect a `namespace` parameter. However, in the Jsonnet code, there is currently no way of referencing the namespace of a given ksonnet environment.

This leads to duplicate info—the namespace is saved both (1) in the environment metadata and (2) as an environment parameter.

**There should be a way to surface this sort of environment metadata for use in prototype definitions, something along the lines of `env.namespace`.**

## Support JSON and YAML

*Planned for 0.9.0* ([#240](https://github.com/ksonnet/ksonnet/issues/240))

Right now only `*.jsonnet` files in the `components/` directory are recognized and used during commands like `ks apply` and `ks show`.

This makes it more difficult for users with existing Kubernetes manifests to transition their app configurations into ksonnet. Although there is a workaround (converting YAML manifests to JSON and saving them as `*.jsonnet`, since JSON is a subset of Jsonnet), this is *not* an ideal workflow.

**In future versions of ksonnet, users should be able to drop in a JSON, YAML, or Jsonnet file into `components/` and be able to deploy it with `ks apply`.** (Noting that only Jsonnet files can fully leverage parameters).

## Reference components in other components

*Planned for 1.0.0*

Right now ksonnet allows users to parameterize literal values (e.g. setting the `image` of a `deployed-service` component). **However, there also needs to be a way for components to reference other components, ideally with a supporting CLI command**.

For example, it might be useful for the Deployment of a `redis-stateless` component to know about a shared `configMap` component. The actual implementation of this is up in the air--whether it looks like component references via `ks param` or something like `ks add <component1> <component2>`.

Either way, it is important for this command to facilitate *incremental* additions in an intuitive way. After all, when developing Kubernetes apps, it's rare to know all of the various components you need from the beginning.

## Establish clear `ks apply` semantics

*Planned for 0.9.0* ([#200](https://github.com/ksonnet/ksonnet/issues/200))

Currently, `ks apply` garbage collects *all* previously created API resources from your ksonnet app, before recreating everything defined in `components/`. This has a lot of the advantages of declarative configuration:
* Deleting a component in your ksonnet app is actually reflected server-side
* It's generally easier for users to reason through what will run on their cluster

However, it also leads to bugs where `ks apply` clobbers fields that ought to have been retained. This means that users get a different `nodePort` each time they `ks apply` a `ClusterIP` Service, and that changes from horizontal pod autoscaling are lost.

To avoid this issue, `ks apply` needs to have a smarter way of merging in new changes (client) with existing configuration (server). Other tools like `kubectl` address this issue with two approaches:
* [Strategic, key-based merge](https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment) (`kubectl patch`)
* [Three-way merge](https://kubernetes.io/docs/tutorials/object-management-kubectl/declarative-object-management-configuration/#how-apply-calculates-differences-and-merges-changes) (`kubectl apply`).

**`ks` will likely adopt `kubectl`'s merge strategies to be in line with what users are familiar with**, and work with upstream `kubectl` to remove other unintuitive behaviors (documented in some [integration tests](https://github.com/hausdorff/ksonnet/commit/8e8522938fe0f940cd5af6f19bc1fa47bf24bfc6)). This overlaps with the work of the [Declarative App Def WG](https://github.com/kubernetes/community/tree/master/wg-app-def), which has discussed refactoring and standardizing `kubectl` behavior for other tools.

## Cleanly remove components

*Planned for 0.9.0* ([#243](https://github.com/ksonnet/ksonnet/issues/243))

When users `ks generate` a component (`foo`) from a prototype, the ksonnet framework does two things:
* Creates a new `foo.jsonnet` file in the `components/` directory
* Adds a K/V map of all of `foo`'s parameters into the `components/params.libsonnet` file

Modifying `foo` for a specific environment also creates a K/V parameters map in `environments/<env-name>/params.libsonnet`.

*All* of these traces need to be cleaned up for a component to be properly removed. Incomplete removals of a component can conflict with future changes, and cause `ks apply` to break. However, these sort of changes are not an uncommon use case. **For a better developer experience, deletion and renaming of components should be abstracted behind a CLI command.**

## Encryption support for secrets

*TBD, discussion of scope pending* ([#255](https://github.com/ksonnet/ksonnet/issues/255))

ksonnet takes a declarative approach, meaning that all configuration is managed in version-controllable files (and potentially integrated into a Gitops workflow).

However, this isn't actually quite the case for secrets. **While you can currently create a secret with `ksonnet-lib`, checking in its component file *could* risk exposing the secret**, because the secret's values would be available in plaintext within the repo (specifically, `<value> | base64`).

To allow secrets to be checked in with the rest of the ksonnet app, there needs to be some way of encrypting and decrypting them. Integration with [Bitnami's sealed secrets](https://github.com/bitnami/sealed-secrets) is one possible approach. See the issue linked above for more details on drawbacks and alternative approaches.

## Improve output of `ks diff`

*Planned for 1.0.0*

Currently `ks diff` output is not very user-friendly because it is a "dumb" diff. The parameters that the user *actually* changed are buried under fields autopopulated by the Kubernetes API server (e.g. `status`).

**Ideally, the fields that ksonnet modifies are tracked in a way that makes it possible to hide other, non-essential fields during `ks diff`.** This work may tie into the [changes to `ks apply` semantics](#establish-clear-ks-apply-semantics). Proposed solutions include an annotation that lists ksonnet-specific fields (so that other can be filtered), but more discussion is necessary.

## Show errors in app files (`ks lint`)

*Planned for 0.9.0* ([#61](https://github.com/ksonnet/ksonnet/issues/61))

ksonnet's "magic" largely results from file autogeneration, but that also makes it more vulnerable to errors when users delete or rename files. One solution is to add CLI commands to cover more use cases (like [removing components](#cleanly-remove-components)), so that users have the tools that they need and can avoid mucking around in the files.

However, there will always be unique scenarios where advanced users need the fine-grained control of editing Jsonnet files themselves. **These developers need a static checker, `ks lint`, to warn them if their Jsonnet code or ksonnet app structure is no longer valid.** Such a tool can be integrated into whatever CI/CD workflows that ksonnet developers use.

## Update VSCode extension

*Planned for 0.9.0* ([#259](https://github.com/ksonnet/ksonnet/issues/259))

[Current workaround](https://kubernetes.slack.com/archives/C6JLE4L9X/p1513191982000078)

The existing VSCode extension does not work out-of-the-box with `ks`-created `*.jsonnet` files. This is due to a few reasons:
* **The `k.libsonnet` files are nested under `environments/<env-name>/.metadata`.** For the parsing to work, these ksonnet libraries need to be included in the `jsonnet.libPaths` VSCode setting. Even if the user clones `ksonnet/ksonnet-lib`, a given ksonnet app might be using a different version of this library.

* **External prototypes and libraries in `vendor/`** also need to have the appropriate paths in the component files where they're used.

* **The extension currently doesn't support `std.extVar`**. `ks` uses this to resolve components and parameters in a hierarchical manner.

**The VSCode extension needs to be updated to work better with `ks`-generated files**, especially because the extension provides autocompletion that makes ksonnet development easier and faster.

## Improve `ksonnet-lib` generation

*Planned for 0.9.0* (issues linked below)

A couple of changes are planned to improve the way in which [`ksonnet-lib`](https://github.com/ksonnet/ksonnet-lib/tree/master/ksonnet.beta.3), the Jsonnet library for the Kubernetes API, is generated:
* Conversion of the K8s API --> Jsonnet code via an [AST](https://en.wikipedia.org/wiki/Abstract_syntax_tree) ([#221](https://github.com/ksonnet/ksonnet/issues/221))
* Autogenerated documentation, since there is none at the moment ([#239](https://github.com/ksonnet/ksonnet/issues/239))
* Support for Kubernetes 1.9 ([#260](https://github.com/ksonnet/ksonnet/issues/260))

## Refactor common code into `ksonnet/client`

*Planned for 0.9.0* ([#215](https://github.com/ksonnet/ksonnet/issues/215))

At the moment, [`ksonnet/ksonnet`](https://github.com/ksonnet/ksonnet) is a hard fork of a related project, [`ksonnet/kubecfg`](https://github.com/ksonnet/kubecfg). Both projects continue to be active, as they are intended for different use cases (`ksonnet` provides a more opinionated *framework*).

Forking was the fastest way to leverage `kubecfg`'s pre-existing code, but is not a long-term solution. **There is ongoing work to refactor this common code into a separate `ksonnet/client` repo.**
