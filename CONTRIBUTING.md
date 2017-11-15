## Contributing to ksonnet

Thank you for taking the time to contribute to ksonnet! Before you submit any PRs, we encourage you to read the instructions for the [developer certificate of origin (DCO)](DCO-SIGNOFF.md), as well as the guidelines below:

* [Build](#build)
* [Manage dependencies](#manage-dependencies)
   * [Add a new dependency](#add-a-new-dependency)
   * [Pin a dependency to a specific version](#pin-a-dependency-to-a-specific-version)
   * [Clean up unnecessary dependencies](#clean-up-unnecessary-dependencies)
* [Test](#test)
* [Make a release](#make-a-release)

### Build

To build ksonnet locally (e.g. for open-source development), run the following in the root of your ksonnet repo:
```
make all
```
This indirectly makes the following targets:
  * `make ks` - Compiles all the code into a `ks` binary
  * `make docs` - Regenerates all the documentation for the ksonnet CLI, based on the code inside `cmd/`.

*Troubleshooting*

`make docs` relies on the `realpath` Linux utility. If you encounter a `realpath: command not found` error, and you are using OSX, run the following command:
```
brew install coreutils
```
Running `make` again afterwards should work.

### Manage dependencies

This project uses `govendor` to manage Go dependencies. To make things easier for reviewers, put updates to `vendor/` in a separate commit within the same PR.

#### Add a new dependency

First make the code change that imports your new library, then run the following in the root of the repo:

```
# View missing dependencies
govendor list +missing

# Download missing dependencies into vendor/
govendor fetch -v +missing
```

#### Pin a dependency to a specific version

Note that pinning a library to a specific version requires extra work.

We do this for key libraries like `client-go` and `apimachinery`. To find the current version used for these particular packages, look in `vendor/vendor.json` for the `version` field (not `versionExact`).

Use that same version in the commands below:

```
# Pin all imported client-go packages to release v3.0
govendor fetch -v k8s.io/client-go/...@v3.0
```

*NOTE:* The command above may pull in new packages from `client-go`, which will
be imported at HEAD.  You need to re-run the above command until *all* imports are at the desired version.

If you make a mistake or miss some libraries, it is safe (and appropriate) to re-run `govendor fetch` with a different version.

#### Clean up unnecessary dependencies

Before sending a non-trivial change for review, consider running the following command:

```
# See unused dependencies
govendor list +unused

# Remove unused dependencies
govendor remove +unused
```

### Test

You can run ksonnet tests with `make test` in the root of the repo, specifying any additional, custom Go flags with `$EXTRA_GO_FLAGS`.

### Make a Release

To make a new release, follow these instructions:

1. Add an appropriate tag.  We do this via `git` (not the github UI) so that the tag is signed.  This process requires you to have write access to the real `master` branch (not your local fork).
   ```
   tag=vX.Y.Z
   git fetch   # update
   git tag -s -m $tag $tag origin/master
   git push origin tag $tag
   ```

2. Wait for the Travis autobuilders to build release binaries.

3. *Now* create the Github release, using the existing tag created
   above.
