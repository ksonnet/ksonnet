## ks prototype

Instantiate, inspect, and get examples for ksonnet prototypes

### Synopsis


Manage, inspect, instantiate, and get examples for ksonnet prototypes.

Prototypes are Kubernetes app configuration templates with "holes" that can be
filled in by (e.g.) the ksonnet CLI tool or a language server. For example, a
prototype for a `apps.v1beta1.Deployment` might require a name and image, and
the ksonnet CLI could expand this to a fully-formed 'Deployment' object.

Commands:
    use      Instantiate prototype, filling in parameters from flags, and
             emitting the generated code to stdout.
    describe Display documentation and details about a prototype
    search   Search for a prototype

```
ks prototype
```

### Examples

```
# Display documentation about prototype
# 'io.ksonnet.pkg.prototype.simple-deployment', including:
#
#   (1) a description of what gets generated during instantiation
#   (2) a list of parameters that are required to be passed in with CLI flags
#
# NOTE: Many subcommands only require the user to specify enough of the
# identifier to disambiguate it among other known prototypes, which is why
# 'simple-deployment' is given as argument instead of the fully-qualified
# name.
ks prototype describe simple-deployment

# Instantiate prototype 'io.ksonnet.pkg.prototype.simple-deployment', using
# the 'nginx' image, and port 80 exposed.
#
# SEE ALSO: Note above for a description of why this subcommand can take
# 'simple-deployment' instead of the fully-qualified prototype name.
ks prototype use simple-deployment \
  --name=nginx                     \
  --image=nginx                    \
  --port=80                        \
  --portName=http

# Search known prototype metadata for the string 'deployment'.
ks prototype search deployment
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks](ks.md)	 - Synchronise Kubernetes resources with config files
* [ks prototype describe](ks_prototype_describe.md)	 - Describe a ksonnet prototype
* [ks prototype list](ks_prototype_list.md)	 - List all known ksonnet prototypes
* [ks prototype preview](ks_prototype_preview.md)	 - Expand prototype, emitting the generated code to stdout
* [ks prototype search](ks_prototype_search.md)	 - Search for a ksonnet prototype
* [ks prototype use](ks_prototype_use.md)	 - Expand prototype, place in components/ directory of ksonnet app

