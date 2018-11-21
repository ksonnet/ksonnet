# Change Log

## [v0.13.1](https://github.com/ksonnet/ksonnet/tree/v0.13.1) (2018-11-21)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.13.0...v0.13.1)

ksonnet 0.13.1 introduces the following changes:

**Bug Fixes:**

* Fix panic during ksonnet-lib generate - when parameters exist after body parameter in cluster OpenAPI spec

## [v0.13.0](https://github.com/ksonnet/ksonnet/tree/v0.13.0) (2018-09-20)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.12.0...v0.13.0)

ksonnet 0.13 introduces the following changes:

**Enhancements:**

* Added `--without-modules` flag to list params set within the environment [\#817](https://github.com/ksonnet/ksonnet/pull/817)
* Added `--tls-skip-verify` global flag to disable TLS verification  [\#831](https://github.com/ksonnet/ksonnet/pull/831)
* Added reuse of installed packages in other scopes  [\#833](https://github.com/ksonnet/ksonnet/pull/833)
* Added `ks pkg remove` [\#834](https://github.com/ksonnet/ksonnet/pull/834)
* Added validation to check environment exists prior to installing package [\#835](https://github.com/ksonnet/ksonnet/pull/835)
* Added garbage collection for vendored packages  [\#837](https://github.com/ksonnet/ksonnet/pull/837)
* Helm client caches chart configuration to avoid multiple back-to-back downloads  [\#841](https://github.com/ksonnet/ksonnet/pull/841)
* Changed `ks import` to use `/` by default [\#843](https://github.com/ksonnet/ksonnet/pull/843)
* Upgraded cobra and pflag [\#845](https://github.com/ksonnet/ksonnet/pull/845)
* Added `--dir` flag to allow running ksonnet from any path  [\#847](https://github.com/ksonnet/ksonnet/pull/847)
* Added `--override` flag to `ks env set` [\#855](https://github.com/ksonnet/ksonnet/pull/855)
* Qualified packages to allow identically named packages from different registries to be installed  [\#855](https://github.com/ksonnet/ksonnet/pull/855)
* Dropped support for 0.0.1 apps [\#855](https://github.com/ksonnet/ksonnet/pull/855)
* Added `--override` flag to `ks env targets` [\#865](https://github.com/ksonnet/ksonnet/pull/865)
* Added sort for resources to handle semantic dependencies [\#867](https://github.com/ksonnet/ksonnet/pull/867)

**Bug Fixes:**

* Fixed `ks env current` to allow setting an environment only if it exists [\#818](https://github.com/ksonnet/ksonnet/pull/818)
* Updated deployment example from `labels` to `withLabels`  [\#823](https://github.com/ksonnet/ksonnet/pull/823)
* Fixed sourcePath for imported json on root module [\#844](https://github.com/ksonnet/ksonnet/pull/844)
* Fixed failing Jsonnet version and pipeline e2e tests [\#852](https://github.com/ksonnet/ksonnet/pull/852)
* Upgraded ksonnet-lib printer to avoid double-escaping param values [\#863](https://github.com/ksonnet/ksonnet/pull/863)

## [v0.12.0](https://github.com/ksonnet/ksonnet/tree/v0.12.0) (2018-8-2)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.11.0...v0.12.0)

ksonnet 0.12.0 introduces the following changes:

**Enhancements:**

* Added Helm Registry support where charts operate as ksonnet parts [\#583](https://github.com/ksonnet/ksonnet/pull/583)
* Added docker-image target to Makefile for building an image with ks [\#588](https://github.com/ksonnet/ksonnet/pull/588)
* Added `--installed` flag to `ks pkg list` to show installed packages [\#600](https://github.com/ksonnet/ksonnet/pull/600)
* Changed location of cached registry manifests [\#604](https://github.com/ksonnet/ksonnet/pull/604)
* Added `--server` to change Kubernetes server address in an environment [\#612](https://github.com/ksonnet/ksonnet/pull/612)
* Added `--api-spec` to change Kubernetes API version in an environment [\#618](https://github.com/ksonnet/ksonnet/pull/618)
* Added `ks registry set <name> --uri <uri>` command to update a registry URI [\#622](https://github.com/ksonnet/ksonnet/pull/622)
* Changed GitHub-based registries to automatically follow remote branches [\#622](https://github.com/ksonnet/ksonnet/pull/622)
* Removed versioning for ksonnet registries [\#632](https://github.com/ksonnet/ksonnet/pull/632)
* Added retries up to five times for `ks apply` [\#639](https://github.com/ksonnet/ksonnet/pull/639)
* Added `lib/` to jsonnet path [\#647](https://github.com/ksonnet/ksonnet/pull/647)
* Changed vendored packages to be qualified with their versions and allow side-by-side installation [\#669](https://github.com/ksonnet/ksonnet/pull/669)
* Updated `ks pkg list` to show canonical versions for packages [\#673](https://github.com/ksonnet/ksonnet/pull/673)
* Added package version to `ks pkg list` [\#673](https://github.com/ksonnet/ksonnet/pull/673)
* Added package versioning support where fully-qualified package identifiers can be specified by `<registry>/<pkg>@<version>` [\#683](https://github.com/heptio/ksonnet-website/pull/683)
* Added `--output=json` to print tabular output as JSON [\#695](https://github.com/ksonnet/ksonnet/pull/695)
* Added support for packages to be installed in an environment with `ks pkg install --env <env> <registry/package>` [\#697](https://github.com/ksonnet/ksonnet/pull/697)
* Added environment scope for packages with `ks pkg list` [\#727](https://github.com/ksonnet/ksonnet/pull/727)
* Added a force option to allow re-installing an existing version of a package with `ks pkg install --force` [\#744](https://github.com/ksonnet/ksonnet/pull/744)
* Updated `ks upgrade` to change environment target separators from `/` to `.` [\#792](https://github.com/ksonnet/ksonnet/pull/792)
* Change `ks component list` to aggregate components from all modules [\#797](https://github.com/ksonnet/ksonnet/pull/797)
* Updated go-jsonnet version from [dfddf2b](https://github.com/google/go-jsonnet/commit/dfddf2b4e3aec377b0dcdf247ff92e7d078b8179) to [v0.11.2](https://github.com/google/go-jsonnet/releases/tag/v0.11.2) [\#800](https://github.com/ksonnet/ksonnet/pull/800)

**Bug Fixes:**

* Allowed component selection for `ks diff` [\#592](https://github.com/ksonnet/ksonnet/pull/592)
* Ensured registry paths exist prior to adding [\#601](https://github.com/ksonnet/ksonnet/pull/601)
* Picked up proper kubernetes version when running on OpenShift [\#640](https://github.com/ksonnet/ksonnet/pull/640)
* Re-added docker image resolver for setting parameters [\#645](https://github.com/ksonnet/ksonnet/pull/645)
* Fixed case where `ks apply --dry-run` modified the cluster [\#699](https://github.com/ksonnet/ksonnet/pull/699)
* Reworked failing end-to-end tests [\#706](https://github.com/ksonnet/ksonnet/pull/706)
* Fixed error message when passing verbose flag to `ks generate` [\#772](https://github.com/ksonnet/ksonnet/pull/772)
* Update resource version on retry due to conflict [\#787](https://github.com/ksonnet/ksonnet/pull/787)
* Fixed case where params are required to render multiple components [\#790](https://github.com/ksonnet/ksonnet/pull/790)
* Allow removing component in a module with dot notation [\#796](https://github.com/ksonnet/ksonnet/pull/796)
* Fixed sorting of objects for consistent `ks diff` behavior [\#808](https://github.com/ksonnet/ksonnet/pull/808)
* Fixed `ks diff` to return all object types [\#811](https://github.com/ksonnet/ksonnet/pull/811)
* Updated `ks diff` to use environment rather than current context [\#811](https://github.com/ksonnet/ksonnet/pull/811)

## [v0.11.0](https://github.com/ksonnet/ksonnet/tree/v0.11.0) (2018-6-1)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.10.2...v0.11.0)

ksonnet 0.11 introduces the following changes:

* `ks apply` will now use merge patching when it can, so Service NodePorts will not be reassigned
* `ks diff` now provides more concise output
* ksonnet built-in prototypes have been reworked to not use ksonnet-lib
* Allow parameters with underscores in the name
* Jsonnet printer now produces jsonnet formatter compliant output
* `ks show` output in JSON format will use a Kubernetes list object
* Components can now return lists and ksonnet will wrap them in a Kubernetes list object
* App override configuration now has a kind/apiVersion
* Fixed various panics due to ksonnet not being run in an application or receiving unknown input

## [v0.10.2](https://github.com/ksonnet/ksonnet/tree/v0.10.2) (2018-5-11)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.10.1...v0.10.2)

ksonnet 0.10.2 is a bug fix release

* gracefully recover when setting parameters from the command line
* add module paths to jsonnet path when evaluating components
* support all Jsonnet language functions
* adds option to reduce verbosity of debug logging
* re-enable native functions for Jsonnet vm
* list components in alphabetical order
* disable ksonnet-lib import in environments unless explicitly required
* omit library gitVersion if it is null

## [v0.10.1](https://github.com/ksonnet/ksonnet/tree/v0.10.1) (2018-04-27)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.10.0...v0.10.1)

ksonnet 0.10.1 fixes issues with evaluating components

## [v0.10.0](https://github.com/ksonnet/ksonnet/tree/v0.10.0) (2018-04-26)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.9.0...v0.10.0)

ksonnet 0.10.0 introduces the following new features:

* support for YAML/JSON components
* global environment parameters
* use Jsonnet AST for all Jsonnet transformations
* filesystem based registries
* upgrade jsonnet to 0.10
* registry and environment overrides
* many bug fixes and usability improvements

**Closed issues:**

- Iteration plan for release 0.10 [\#360](https://github.com/ksonnet/ksonnet/issues/360)
- Public Git Repository Setup [\#352](https://github.com/ksonnet/ksonnet/issues/352)
- Generate registry scaffolding [\#234](https://github.com/ksonnet/ksonnet/issues/234)
- ks disk-based registry can't generate prototype [\#480](https://github.com/ksonnet/ksonnet/issues/480)
- param::import is broken in alpha version [\#475](https://github.com/ksonnet/ksonnet/issues/475)
- ks apply --component not working on 0.10.0-alpha.2 or 0.10.0-alpha.3 [\#472](https://github.com/ksonnet/ksonnet/issues/472)
- ks param set does not seem to work for yaml components [\#468](https://github.com/ksonnet/ksonnet/issues/468)
- ks import doesn't take namespace into account [\#467](https://github.com/ksonnet/ksonnet/issues/467)
- X.509 certificate issue when executing ks init [\#436](https://github.com/ksonnet/ksonnet/issues/436)
- Move `main.go` to `cmd/ks` [\#112](https://github.com/ksonnet/ksonnet/issues/112)

**Merged pull requests:**

- Adding initial override documentation [\#487](https://github.com/ksonnet/ksonnet/pull/487) ([bryanl](https://github.com/bryanl))
- Adding initial registry document [\#486](https://github.com/ksonnet/ksonnet/pull/486) ([bryanl](https://github.com/bryanl))
- updating ksonnet-lib to v0.1.0 [\#485](https://github.com/ksonnet/ksonnet/pull/485) ([bryanl](https://github.com/bryanl))
- Removing index columns from e2e tests [\#484](https://github.com/ksonnet/ksonnet/pull/484) ([bryanl](https://github.com/bryanl))
- edit ksonnet license info so that GitHub recognizes it [\#483](https://github.com/ksonnet/ksonnet/pull/483) ([eirinikos](https://github.com/eirinikos))
- Rework jsonnet external vars [\#482](https://github.com/ksonnet/ksonnet/pull/482) ([bryanl](https://github.com/bryanl))
- Update printer for jsonnet nulls [\#481](https://github.com/ksonnet/ksonnet/pull/481) ([bryanl](https://github.com/bryanl))
- Support double quoted params in prototypes [\#479](https://github.com/ksonnet/ksonnet/pull/479) ([bryanl](https://github.com/bryanl))
- .yml is a valid YAML file extension [\#477](https://github.com/ksonnet/ksonnet/pull/477) ([bryanl](https://github.com/bryanl))
- Cleaning up travis build [\#474](https://github.com/ksonnet/ksonnet/pull/474) ([bryanl](https://github.com/bryanl))
- bug: filter components [\#473](https://github.com/ksonnet/ksonnet/pull/473) ([bryanl](https://github.com/bryanl))
- Allow setting nested global env vars from cli [\#471](https://github.com/ksonnet/ksonnet/pull/471) ([bryanl](https://github.com/bryanl))
- Add random id to YAML imported objects [\#469](https://github.com/ksonnet/ksonnet/pull/469) ([bryanl](https://github.com/bryanl))

## [v0.10.0-alpha.3](https://github.com/ksonnet/ksonnet/tree/v0.10.0-alpha.3) (2018-04-20)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.10.0-alpha.2...v0.10.0-alpha.3)

**Fixed bugs:**

- Support GitHub registries with contents in the root path. [\#459](https://github.com/ksonnet/ksonnet/issues/459)

**Closed issues:**

- ksonnet version 0.10.0-alpha.2 unable to import libsonnet files [\#464](https://github.com/ksonnet/ksonnet/issues/464)
- Cannot use importstr: ERROR find objects: output: unknown node type: \(\*ast.ImportStr\) [\#461](https://github.com/ksonnet/ksonnet/issues/461)
- update ksonnet libs [\#454](https://github.com/ksonnet/ksonnet/issues/454)
- ks param set command is not able to set array or map. [\#448](https://github.com/ksonnet/ksonnet/issues/448)
- ks param list command shows error if components/params.libsonnet has null or array. [\#447](https://github.com/ksonnet/ksonnet/issues/447)
- ks import -f some-component.yaml does not set component index in components/params.libsonnet [\#437](https://github.com/ksonnet/ksonnet/issues/437)
- Regression from 0.9.1: Getting "ERROR Unauthorized" when using --token parameter in 0.9.2 [\#430](https://github.com/ksonnet/ksonnet/issues/430)

**Merged pull requests:**

- Reorganize layout [\#466](https://github.com/ksonnet/ksonnet/pull/466) ([bryanl](https://github.com/bryanl))
- Supply proper codes for building env jsonnet [\#465](https://github.com/ksonnet/ksonnet/pull/465) ([bryanl](https://github.com/bryanl))
- Updating ksonnet-lib printer to handle importstr [\#463](https://github.com/ksonnet/ksonnet/pull/463) ([bryanl](https://github.com/bryanl))
- When importing YAML, extract objects into separate component files [\#462](https://github.com/ksonnet/ksonnet/pull/462) ([bryanl](https://github.com/bryanl))
- Generate proper vendor paths for gh registry with in root [\#460](https://github.com/ksonnet/ksonnet/pull/460) ([bryanl](https://github.com/bryanl))
- evaluator is missing ksonnet vendor path [\#458](https://github.com/ksonnet/ksonnet/pull/458) ([bryanl](https://github.com/bryanl))
- Fixed year 2017 -\> 2018 [\#456](https://github.com/ksonnet/ksonnet/pull/456) ([uthark](https://github.com/uthark))
- Add env update command [\#455](https://github.com/ksonnet/ksonnet/pull/455) ([bryanl](https://github.com/bryanl))
- Handle null as a value [\#453](https://github.com/ksonnet/ksonnet/pull/453) ([bryanl](https://github.com/bryanl))

## [v0.10.0-alpha.2](https://github.com/ksonnet/ksonnet/tree/v0.10.0-alpha.2) (2018-04-17)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.10.0-alpha.1...v0.10.0-alpha.2)

**Fixed bugs:**

- Drop-in YAML components with parameters [\#426](https://github.com/ksonnet/ksonnet/issues/426)

**Closed issues:**

- ksonnet 1.10 no longer uses environments/env/main.jsonnet [\#446](https://github.com/ksonnet/ksonnet/issues/446)
- ks diff local:prod local:dev: RUNTIME ERROR: max stack frames exceeded [\#445](https://github.com/ksonnet/ksonnet/issues/445)
- `ks generate` panics "runtime error: index out of range" instead of printing usage [\#441](https://github.com/ksonnet/ksonnet/issues/441)
- `ks param list`: ERROR retrieve values for mixin.spec.template.spec.hostNetwork: can't handle type bool [\#438](https://github.com/ksonnet/ksonnet/issues/438)
- `ks generate prototype --help` should be a thing [\#120](https://github.com/ksonnet/ksonnet/issues/120)

**Merged pull requests:**

- Set maps in arrays from cli [\#452](https://github.com/ksonnet/ksonnet/pull/452) ([bryanl](https://github.com/bryanl))
- Convert diff to use pipline [\#451](https://github.com/ksonnet/ksonnet/pull/451) ([bryanl](https://github.com/bryanl))
- Ensure param diff works with all component types [\#450](https://github.com/ksonnet/ksonnet/pull/450) ([bryanl](https://github.com/bryanl))
- Use main.jsonnet for rendering components [\#449](https://github.com/ksonnet/ksonnet/pull/449) ([bryanl](https://github.com/bryanl))
- Support bool for object values in YAML [\#444](https://github.com/ksonnet/ksonnet/pull/444) ([bryanl](https://github.com/bryanl))
- Don't panic when runing `ks generate` with no args [\#443](https://github.com/ksonnet/ksonnet/pull/443) ([bryanl](https://github.com/bryanl))
- Merge in new YAML params [\#442](https://github.com/ksonnet/ksonnet/pull/442) ([bryanl](https://github.com/bryanl))
- Fixing broken link in docs and also typos [\#439](https://github.com/ksonnet/ksonnet/pull/439) ([Maerville](https://github.com/Maerville))
- Add help when generating or previewing prototypes [\#435](https://github.com/ksonnet/ksonnet/pull/435) ([bryanl](https://github.com/bryanl))
- FS registry can handle relative paths [\#434](https://github.com/ksonnet/ksonnet/pull/434) ([bryanl](https://github.com/bryanl))

## [v0.10.0-alpha.1](https://github.com/ksonnet/ksonnet/tree/v0.10.0-alpha.1) (2018-04-10)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.9.2...v0.10.0-alpha.1)

**Implemented enhancements:**

- Generated environment parameters can't assume component parameters path [\#354](https://github.com/ksonnet/ksonnet/issues/354)
- Upgrade to client-go version 6 [\#297](https://github.com/ksonnet/ksonnet/issues/297)
- Support local folders for registries [\#232](https://github.com/ksonnet/ksonnet/issues/232)
- Load test data as fixtures [\#168](https://github.com/ksonnet/ksonnet/issues/168)
- Document or provide better error messages for breaking ks changes [\#155](https://github.com/ksonnet/ksonnet/issues/155)
- Consider naming the default environment after the initialized context [\#82](https://github.com/ksonnet/ksonnet/issues/82)
- Move `constructBaseObj` out of root.go [\#72](https://github.com/ksonnet/ksonnet/issues/72)
- Autogenerate .gitignore file \(e.g. for libs or hidden files\) [\#55](https://github.com/ksonnet/ksonnet/issues/55)

**Fixed bugs:**

- Misleading kubeconfig log message  [\#287](https://github.com/ksonnet/ksonnet/issues/287)
- Consolidate type checking between params and prototypes [\#44](https://github.com/ksonnet/ksonnet/issues/44)

**Closed issues:**

- In-cluster API access [\#431](https://github.com/ksonnet/ksonnet/issues/431)
- inception server not working [\#429](https://github.com/ksonnet/ksonnet/issues/429)
- Question: import not available kubeflow/core/all.libsonnet and other problems with master [\#420](https://github.com/ksonnet/ksonnet/issues/420)
- ks pkg install does not install the correct branch set in the registry [\#398](https://github.com/ksonnet/ksonnet/issues/398)
- 1.9 & 1.10 support? [\#394](https://github.com/ksonnet/ksonnet/issues/394)
- `ks -v` should print an equivalent `jsonnet` command line string  [\#378](https://github.com/ksonnet/ksonnet/issues/378)
- Proposal: Create `ks param unset` command for "unsetting" params [\#325](https://github.com/ksonnet/ksonnet/issues/325)
- use /swagger.json from apiserver for api-spec [\#264](https://github.com/ksonnet/ksonnet/issues/264)
- Update VSCode extension to work better with the new `ks`-generated files [\#224](https://github.com/ksonnet/ksonnet/issues/224)
- Implement `ks install` [\#217](https://github.com/ksonnet/ksonnet/issues/217)
- Increase test coverage [\#216](https://github.com/ksonnet/ksonnet/issues/216)
- Should be able to be used offline [\#204](https://github.com/ksonnet/ksonnet/issues/204)
- Add bash completion [\#124](https://github.com/ksonnet/ksonnet/issues/124)
- Consolidate `expandEnvObjs` and `expandEnvCmdObjs` into one function [\#70](https://github.com/ksonnet/ksonnet/issues/70)
- Move core logic of new commands \(`dep` and `registry`\) to pkg kubecfg [\#65](https://github.com/ksonnet/ksonnet/issues/65)

**Merged pull requests:**

- Action option tests [\#433](https://github.com/ksonnet/ksonnet/pull/433) ([bryanl](https://github.com/bryanl))
- bug: param list with env [\#432](https://github.com/ksonnet/ksonnet/pull/432) ([bryanl](https://github.com/bryanl))
- listing params for yaml components works for more cases [\#428](https://github.com/ksonnet/ksonnet/pull/428) ([bryanl](https://github.com/bryanl))
- add .ks\_environment to generated gitignore [\#425](https://github.com/ksonnet/ksonnet/pull/425) ([bryanl](https://github.com/bryanl))
- Add `ks env current` command [\#424](https://github.com/ksonnet/ksonnet/pull/424) ([bryanl](https://github.com/bryanl))
- Add JSON output to `env list` [\#423](https://github.com/ksonnet/ksonnet/pull/423) ([bryanl](https://github.com/bryanl))
- import component from a http URL [\#422](https://github.com/ksonnet/ksonnet/pull/422) ([bryanl](https://github.com/bryanl))
- Support global env parameters [\#421](https://github.com/ksonnet/ksonnet/pull/421) ([bryanl](https://github.com/bryanl))
- Add e2e for built in prototypes [\#417](https://github.com/ksonnet/ksonnet/pull/417) ([bryanl](https://github.com/bryanl))
- Print location of kubeconfig [\#416](https://github.com/ksonnet/ksonnet/pull/416) ([bryanl](https://github.com/bryanl))
- Update `show` to ksonnet action [\#415](https://github.com/ksonnet/ksonnet/pull/415) ([bryanl](https://github.com/bryanl))
- Update delete action [\#414](https://github.com/ksonnet/ksonnet/pull/414) ([bryanl](https://github.com/bryanl))
- Print more stack frames on infinite loops [\#413](https://github.com/ksonnet/ksonnet/pull/413) ([redbaron](https://github.com/redbaron))
- generate env params with ext var to support modules [\#412](https://github.com/ksonnet/ksonnet/pull/412) ([bryanl](https://github.com/bryanl))
- Param unset [\#411](https://github.com/ksonnet/ksonnet/pull/411) ([bryanl](https://github.com/bryanl))
- Override default env name [\#410](https://github.com/ksonnet/ksonnet/pull/410) ([bryanl](https://github.com/bryanl))
- ignore app override by default [\#409](https://github.com/ksonnet/ksonnet/pull/409) ([bryanl](https://github.com/bryanl))
- Handle YAML components which end with `--` [\#408](https://github.com/ksonnet/ksonnet/pull/408) ([bryanl](https://github.com/bryanl))
- Rename component namespaces to modules [\#407](https://github.com/ksonnet/ksonnet/pull/407) ([bryanl](https://github.com/bryanl))
- Allow for regeneration of lib and registry cache [\#406](https://github.com/ksonnet/ksonnet/pull/406) ([bryanl](https://github.com/bryanl))
- Allow user to skip inclusion of default registries [\#405](https://github.com/ksonnet/ksonnet/pull/405) ([bryanl](https://github.com/bryanl))
- remove stringsAppendToPath [\#404](https://github.com/ksonnet/ksonnet/pull/404) ([bryanl](https://github.com/bryanl))
- moved env to pkg/env [\#403](https://github.com/ksonnet/ksonnet/pull/403) ([bryanl](https://github.com/bryanl))
- update actions to use action args [\#402](https://github.com/ksonnet/ksonnet/pull/402) ([bryanl](https://github.com/bryanl))
- Reworks apply action [\#401](https://github.com/ksonnet/ksonnet/pull/401) ([bryanl](https://github.com/bryanl))
- Create plumbing for cluster based e2e tests [\#399](https://github.com/ksonnet/ksonnet/pull/399) ([bryanl](https://github.com/bryanl))
- Action validate [\#397](https://github.com/ksonnet/ksonnet/pull/397) ([bryanl](https://github.com/bryanl))
- introduce method to test cmd flags [\#395](https://github.com/ksonnet/ksonnet/pull/395) ([bryanl](https://github.com/bryanl))
- removing need for github access for registry tests [\#393](https://github.com/ksonnet/ksonnet/pull/393) ([bryanl](https://github.com/bryanl))
- use provided tls configuration if possible when fetching api spec [\#392](https://github.com/ksonnet/ksonnet/pull/392) ([bryanl](https://github.com/bryanl))

## [v0.9.1](https://github.com/ksonnet/ksonnet/tree/v0.9.1) (2018-03-08)

This patch focuses on fixes around usability bugs.

**Closed issues:**

- version 0.9 - ks show <env> not picking up env param overrides [\#346](https://github.com/ksonnet/ksonnet/issues/346)
- ks delete default fails [\#342](https://github.com/ksonnet/ksonnet/issues/342)

**Merged pull requests:**

- bug: mapContainer extension typo [\#350](https://github.com/ksonnet/ksonnet/pull/350) ([bryanl](https://github.com/bryanl))
- bug: env used incorrent params for rendering [\#349](https://github.com/ksonnet/ksonnet/pull/349) ([bryanl](https://github.com/bryanl))
- bug: 0.1.0 apps don't rename envs in config [\#347](https://github.com/ksonnet/ksonnet/pull/347) ([bryanl](https://github.com/bryanl))
- Fix formatting for param diff [\#344](https://github.com/ksonnet/ksonnet/pull/344) ([jessicayuen](https://github.com/jessicayuen))
- Parse server version from GitVersion [\#343](https://github.com/ksonnet/ksonnet/pull/343) ([jessicayuen](https://github.com/jessicayuen))

## [v0.9.0](https://github.com/ksonnet/ksonnet/tree/v0.9.0) (2018-03-05)
[Iteration Plan](https://github.com/ksonnet/ksonnet/issues/306)

[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.8.0...v0.9.0)

To update older ksonnet applications, run `ks upgrade --help`.

### Overview

This release focuses on two major areas:

1. Changes to the underlying ksonnet-lib dependency and utilizing them in ks.
The changes involve a major uplift to how the ksonnet language APIs are
generated, so that support for future Kubernetes versions are easier.

2. Improvements to the support for environments and components. Environments
are now able to specify targets, to apply a subset of components as opposed
to all components. We've introduced the concept of component namespaces to
add division and hierarchy to components. We've also added commands to support
removing and listing of components.

### Changes to Environment Metadata

#### spec.json > app.yaml

In _0.8.0_, each *ks* application contained a _spec.json_ file per environment.
This file contained the environment specification details; corresponding to a
Kubernetes cluster server address and namespace.

_i.e._,
```
{
  "server": "https://35.230.00.00",
  "namespace": "default"
}
```

With _0.9.0_, we will be consolidating majority of the application specification
details into a top-level _app.yaml_ file. As opposed to opening multiple files,
this approach will make it easier for users to configure changes in a single
place. Similar to `spec.json`, this metadata will be auto-generated when a
environment is created or modified from the command line. However, it is also
meant to be user-modifiable.

_i.e._,
```
apiVersion: 0.0.1
environments:
  default:
    destination:
      namespace: default
      server: https://35.230.00.00
    k8sVersion: v1.7.0
    path: default
kind: ksonnet.io/app
name: new
registries:
  incubator:
    gitVersion:
      commitSha: 422d521c05aa905df949868143b26445f5e4eda5
      refSpec: master
    protocol: github
    uri: github.com/ksonnet/parts/tree/master/incubator
version: 0.0.1
```

You will notice a couple new fields under the _environments_ field.

1. _destination_ is identical to the contents of the previous _spec.json_,
   containing the server address and namespace that this environment points to.

2. _k8sVersion_ is the Kubernetes version of the server pointed to at the _destination_
   field.

3. _path_ is the relative path in the _environments_ directory that contains
   other metadata associated with this environment. By default, path is the
   environment name.

#### Consolidation of lib files

In _0.8.0_, each environment's `.metadata` directory stored 3 files related to
the generated `ksonnet-lib`. It was unecessary and also costly as the number of
environments grow. We didn't need to store multiple copies of the same API
version on disk.

With _0.9.0_, the Kubernetes API version that each environment uses will be
recorded in the environment specification (as seen in the previous section).
The metadata files are cached once locally per k8s API version in `lib`.

These files also no longer need to be checked into source control, as *ks*
will auto-generate lib files that aren't found.

#### Targets & Component Namespaces

In _0.8.0_, there was no simple way for users to declare that a environment
should only operate on a subset of components.

With _0.9.0_, environments can now choose the set of components that they wish
to operate (_i.e._, _apply_, _delete_, etc.) on. These targets can be specified
in the _app.yaml_ file, mentioned in an earlier section.

For example, if the _components_ directory is structured as follows:
```
my-ks-app
├── components
│   ├── auth
│   │   ├── ca-secret.jsonnet
│   │   ├── params.libsonnet
│   │   └── tls-certificate.jsonnet
│   ├── dev
│   │   ├── memcached.jsonnet
│   │   └── params.libsonnet
│   ├── params.libsonnet
│   └── prod
│   ...
```

An environment configuration in _app.yaml_ may appear as follows:
```
environments:
  dev:
    k8sVersion: 1.7.0
    destinations:
      namespace: default
      server: https://35.230.00.00
    targets:
      - auth
      - dev
```

In the above example, the _dev_ environment would only operate on the
components within the _auth_ and _dev_ component namespaces.

Note: Component files do not need to be namespaced. Top-level components
and individual component files can also be referenced by _targets_.

### Command Changes

#### ks component list

`ks component list` is a new command. See docs [here](https://github.com/ksonnet/ksonnet/blob/v0.9.0/docs/cli-reference/ks_component_list.md).

#### ks component rm

`ks component rm` is a new command. See docs [here](https://github.com/ksonnet/ksonnet/blob/v0.9.0/docs/cli-reference/ks_component_rm.md).

### ksonnet-lib Changes

* Create Jsonnet AST printer
* Convert ksonnet-lib generation process to Asonnet AST

### Github

**Closed issues:**

- tutorial document as linked from Google seems semi-broken? [\#322](https://github.com/ksonnet/ksonnet/issues/322)
- Incorrect imports in ks generated files [\#321](https://github.com/ksonnet/ksonnet/issues/321)
- delete component ERROR strconv.Atoi: parsing "8+": invalid syntax [\#316](https://github.com/ksonnet/ksonnet/issues/316)
- ks param set when used with boolean does not create string value [\#311](https://github.com/ksonnet/ksonnet/issues/311)
- Move custom constructors k8s.libsonnet to k.libsonnet [\#304](https://github.com/ksonnet/ksonnet/issues/304)
- ERROR user: Current not implemented on linux/amd64 [\#298](https://github.com/ksonnet/ksonnet/issues/298)
- Difficulty handling components unique to environments [\#292](https://github.com/ksonnet/ksonnet/issues/292)
- ks delete ERROR strconv.Atoi [\#272](https://github.com/ksonnet/ksonnet/issues/272)
- Create darwin binaries and make the available via brew [\#270](https://github.com/ksonnet/ksonnet/issues/270)
- Unable to install packages with the same name under different registries [\#269](https://github.com/ksonnet/ksonnet/issues/269)
- prototypes can't rely on a registry name, but they do [\#262](https://github.com/ksonnet/ksonnet/issues/262)
- Confirm that ksonnet-lib generation works for Kubernetes 1.9[\#260](https://github.com/ksonnet/ksonnet/issues/260)
- ks can't recognise the registry 'kubeflow'[\#258](https://github.com/ksonnet/ksonnet/issues/258)
- ksonnet.io website is not available[\#256](https://github.com/ksonnet/ksonnet/issues/256)
- ks init fails when using $KUBECONFIG env var[\#251](https://github.com/ksonnet/ksonnet/issues/251)
- Badly formatted client-go version string[\#250](https://github.com/ksonnet/ksonnet/issues/250)
- Remove components[\#243](https://github.com/ksonnet/ksonnet/issues/243)
- List components[\#242](https://github.com/ksonnet/ksonnet/issues/242)

**Merged pull requests:**

- ksonnet app.yaml format changes in next minor release. Handle both versions [\#338](https://github.com/ksonnet/ksonnet/pull/338) ([bryanl](https://github.com/bryanl))
- Attempt to generate lib directory when not found [\#337](https://github.com/ksonnet/ksonnet/pull/337) ([jessicayuen](https://github.com/jessicayuen))
- Fix the execution paths for the 0.8 > 0.9 migration warning [\#335](https://github.com/ksonnet/ksonnet/pull/337) ([jessicayuen](https://github.com/jessicayuen))
- Resolve api spec based on swagger version [\#334](https://github.com/ksonnet/ksonnet/pull/334) ([jessicayuen](https://github.com/jessicayuen))
- Use new ksonnet lib generator [\#333](https://github.com/ksonnet/ksonnet/pull/333) ([bryanl](https://github.com/bryanl))
- Set param boolean types as strings [\#331](https://github.com/ksonnet/ksonnet/pull/331) ([jessicayuen](https://github.com/jessicayuen))
- Create support for plugins in ksonnet [\#330](https://github.com/ksonnet/ksonnet/pull/333) ([bryanl](https://github.com/bryanl))
- Fix bug with invalid base.libsonnet import path [\#329](https://github.com/ksonnet/ksonnet/pull/329) ([jessicayuen](https://github.com/jessicayuen))
- App spec to take a single destination [\#328](https://github.com/ksonnet/ksonnet/pull/328) ([jessicayuen](https://github.com/jessicayuen))
- Add warning for running deprecated ks app against ks >= 0.9.0 [\#327](https://github.com/ksonnet/ksonnet/pull/327) ([jessicayuen](https://github.com/jessicayuen))
- Pull client-go logic out of cmd/root.go and into client/ package [\#324](https://github.com/ksonnet/ksonnet/pull/324) ([jessicayuen](https://github.com/jessicayuen))
- Introduce component namespaces [\#323](https://github.com/ksonnet/ksonnet/pull/323) ([bryanl](https://github.com/bryanl))
- Add LibManager for managing k8s API and ksonnet-lib metadata [\#315](https://github.com/ksonnet/ksonnet/pull/315) ([jessicayuen](https://github.com/jessicayuen))
- Migrate environment spec.json to the app.yaml model [\#309](https://github.com/ksonnet/ksonnet/pull/309) ([jessicayuen](https://github.com/jessicayuen))
- Add interface for Environment Spec [\#308](https://github.com/ksonnet/ksonnet/pull/309) ([jessicayuen](https://github.com/jessicayuen))
- Extract ksonnet generator [\#307](https://github.com/ksonnet/ksonnet/pull/307) ([bryanl](https://github.com/bryanl))
- [docs] Clarify prototypes + add troubleshooting issue [\#303](https://github.com/ksonnet/ksonnet/pull/303) ([abiogenesis-now](https://github.com/abiogenesis-now))
- updating 1.0 roadmap [\#302](https://github.com/ksonnet/ksonnet/pull/302) ([bryanl](https://github.com/bryanl))
- use afero when possible [\#301](https://github.com/ksonnet/ksonnet/pull/301) ([bryanl](https://github.com/bryanl))
- Upgrade to client-go version 5 [\#299](https://github.com/ksonnet/ksonnet/pull/299) ([jessicayuen](https://github.com/jessicayuen))
- Proposal: Modular components and cleaner environments [\#295](https://github.com/ksonnet/ksonnet/pull/295) ([jessicayuen](https://github.com/jessicayuen))
- pruning vendor from dep conversion [\#293](https://github.com/ksonnet/ksonnet/pull/293) ([bryanl](https://github.com/bryanl))
- Versions retrospective and fixes [\#289](https://github.com/ksonnet/ksonnet/pull/289) ([hausdorff](https://github.com/hausdorff))
- Add remove component functionality [\#288](https://github.com/ksonnet/ksonnet/pull/288) ([jessicayuen](https://github.com/jessicayuen))
- Construct apimachinery version [\#285](https://github.com/ksonnet/ksonnet/pull/285) ([bryanl](https://github.com/bryanl))
- Removing realpath [\#283](https://github.com/ksonnet/ksonnet/pull/283) ([kris-nova](https://github.com/kris-nova))
- Implement explicit env metadata [\#282](https://github.com/ksonnet/ksonnet/pull/282) ([jessicayuen](https://github.com/jessicayuen))
- Design: propose improvements to the "fresh clone" story [\#280](https://github.com/ksonnet/ksonnet/pull/280) ([hausdorff](https://github.com/hausdorff))
- Design proposal: explicit environment metadata [\#279](https://github.com/ksonnet/ksonnet/pull/279) ([jessicayuen](https://github.com/jessicayuen))
- Small fixes to release process [\#275](https://github.com/ksonnet/ksonnet/pull/275) ([jbeda](https://github.com/jbeda))
- Document using goreleaser [\#274](https://github.com/ksonnet/ksonnet/pull/274) ([jbeda](https://github.com/jbeda))
- Clarify error message for duplicate packages on install [\#271](https://github.com/ksonnet/ksonnet/pull/271) ([jessicayuen](https://github.com/jessicayuen))
- Add command 'component list' [\#268](https://github.com/ksonnet/ksonnet/pull/268) ([jessicayuen](https://github.com/jessicayuen))
- design proposal: ksonnet lib simple constructors [\#267](https://github.com/ksonnet/ksonnet/pull/267) ([bryanl](https://github.com/bryanl))
- convert from govendor to dep [\#265](https://github.com/ksonnet/ksonnet/pull/265) ([bryanl](https://github.com/bryanl))
- Reference current Slack channel in README [\#257](https://github.com/ksonnet/ksonnet/pull/257) ([lblackstone](https://github.com/lblackstone))
- Supports k8s version number including symbols etc. [\#254](https://github.com/ksonnet/ksonnet/pull/254) ([kyamazawa](https://github.com/kyamazawa))
- Handle case where KUBECONFIG is set without named context [\#253](https://github.com/ksonnet/ksonnet/pull/253) ([lblackstone](https://github.com/lblackstone))
- Create a GitHub issue template [\#252](https://github.com/ksonnet/ksonnet/pull/252) ([lblackstone](https://github.com/lblackstone))
- Allow make file to generate ks bin with custom version and name [\#249](https://github.com/ksonnet/ksonnet/pull/249) ([bryanl](https://github.com/bryanl))

## [v0.8.0](https://github.com/ksonnet/ksonnet/tree/v0.8.0) (2017-12-20)
[Full Changelog](https://github.com/ksonnet/ksonnet/compare/v0.7.0...v0.8.0)

**Implemented enhancements:**

- Package list/install is awkward [\#195](https://github.com/ksonnet/ksonnet/issues/195)
- Rework demos/examples in light of \#169 [\#194](https://github.com/ksonnet/ksonnet/issues/194)

**Fixed bugs:**

- `param set` incorrectly supporting hyphenated param names [\#214](https://github.com/ksonnet/ksonnet/issues/214)
- Makefile hardcodes version [\#198](https://github.com/ksonnet/ksonnet/issues/198)
- Accurately read/write non-ASCII param identifiers [\#219](https://github.com/ksonnet/ksonnet/pull/219) ([jessicayuen](https://github.com/jessicayuen))

**Closed issues:**

- Packages should be able to depend on other packages [\#238](https://github.com/ksonnet/ksonnet/issues/238)
- YAML components are currently disabled; docs say they aren't [\#208](https://github.com/ksonnet/ksonnet/issues/208)
- Confusing info in `ks version` [\#199](https://github.com/ksonnet/ksonnet/issues/199)
- Support/document using github token to increase rate limits [\#196](https://github.com/ksonnet/ksonnet/issues/196)
- Issue with redis-stateless prototype [\#193](https://github.com/ksonnet/ksonnet/issues/193)
- ks version missing/incorrect data [\#170](https://github.com/ksonnet/ksonnet/issues/170)
- Create binary releases [\#131](https://github.com/ksonnet/ksonnet/issues/131)
- Check apiVersion numbers [\#75](https://github.com/ksonnet/ksonnet/issues/75)
- Add links to http://ksonnet.heptio.com/ [\#20](https://github.com/ksonnet/ksonnet/issues/20)

**Merged pull requests:**

- Test all branches of GH URI-parsing code [\#245](https://github.com/ksonnet/ksonnet/pull/245) ([hausdorff](https://github.com/hausdorff))
- Implement command `ks registry add` [\#228](https://github.com/ksonnet/ksonnet/pull/228) ([jessicayuen](https://github.com/jessicayuen))
- Check error while enumerating environments [\#220](https://github.com/ksonnet/ksonnet/pull/220) ([tanner-bruce](https://github.com/tanner-bruce))
- \[docs\] Fix premature claim of YAML support in components explanation [\#213](https://github.com/ksonnet/ksonnet/pull/213) ([abiogenesis-now](https://github.com/abiogenesis-now))
- Reverse name, registry columns for pkg list [\#206](https://github.com/ksonnet/ksonnet/pull/206) ([jessicayuen](https://github.com/jessicayuen))
- \[docs\] Remove \(now optional\) `--name` syntax from `ks generate` commands [\#205](https://github.com/ksonnet/ksonnet/pull/205) ([abiogenesis-now](https://github.com/abiogenesis-now))
- Revert "Update default version in Makefile to v0.7.0" [\#203](https://github.com/ksonnet/ksonnet/pull/203) ([jessicayuen](https://github.com/jessicayuen))
- Implement github token to work around rate limits [\#201](https://github.com/ksonnet/ksonnet/pull/201) ([jbeda](https://github.com/jbeda))
