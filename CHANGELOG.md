# Change Log

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
