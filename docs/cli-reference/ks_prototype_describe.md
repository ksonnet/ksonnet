## ks prototype describe

Describe a ksonnet prototype

### Synopsis


Output documentation, examples, and other information for some ksonnet
prototype uniquely identified by some (possibly partial) `prototype-name`. This
includes:

  1. a description of what gets generated during instantiation
  2. a list of parameters that are required to be passed in with CLI flags

`prototype-name` need only contain enough of the suffix of a name to uniquely
disambiguate it among known names. For example, 'deployment' may resolve
ambiguously, in which case 'use' will fail, while 'simple-deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.prototype.simple-deployment'.

```
ks prototype describe <prototype-name>
```

### Examples

```
# Display documentation about prototype, including:
ks prototype describe io.ksonnet.pkg.prototype.simple-deployment

# Display documentation about prototype using a unique suffix of an
# identifier. That is, this command only requires a long enough suffix to
# uniquely identify a ksonnet prototype. In this example, the suffix
# 'simple-deployment' is enough to uniquely identify
# 'io.ksonnet.pkg.prototype.simple-deployment', but 'deployment' might not
# be, as several names end with that suffix.
ks prototype describe simple-deployment
```

### Options inherited from parent commands

```
  -v, --verbose count[=-1]   Increase verbosity. May be given multiple times.
```

### SEE ALSO
* [ks prototype](ks_prototype.md)	 - Instantiate, inspect, and get examples for ksonnet prototypes

