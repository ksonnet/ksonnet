# Contributing to ksonnet

Thank you for taking the time to contribute to ksonnet! Before you submit any PRs, we encourage you to read the instructions for the [developer certificate of origin (DCO)](DCO-SIGNOFF.md), as well as the guidelines below:

* [Build](#build)
* [Manage dependencies](#manage-dependencies)
   * [Add a new dependency](#add-a-new-dependency)
   * [Pin a dependency to a specific version](#pin-a-dependency-to-a-specific-version)
   * [Clean up unnecessary dependencies](#clean-up-unnecessary-dependencies)
* [Test](#test)
* [Make a release](#make-a-release)

## Build

> Ensure that you already have (1) a proper `$GOPATH` and (2) ksonnet installed. If not, follow [these README instructions](/README.md#install) to get set up.

To build ksonnet locally (e.g. for open-source development), run the following in the root of your ksonnet repo (which is `$GOPATH/src/github.com/ksonnet/ksonnet` unless otherwise specified):
```
make all
```
This indirectly makes the following targets:
  * `make ks` - Compiles all the code into a local `ks` binary
  * `make docs` - Regenerates all the documentation for the ksonnet CLI, based on the code inside `cmd/`.

*Troubleshooting*

`make docs` relies on the `realpath` Linux utility. If you encounter a `realpath: command not found` error, and you are using OSX, run the following command:
```
brew install coreutils
```
Running `make` again afterwards should work.

## Manage dependencies

This project uses [`govendor`](https://github.com/kardianos/govendor) to manage Go dependencies.



To make things easier for reviewers, put updates to `vendor/` in a separate commit within the same PR.

### Add a new dependency

To introduce a new dependency to the ksonnet codebase, follow these steps:

1. [Open an issue](https://github.com/ksonnet/ksonnet/issues) to discuss the use case and need for the dependency with the project maintainers.

2. After building agreement, vendor the dependency using `govendor`.
    * Where possible, **please [pin the dependency](#pin-a-dependency-to-a-specific-version) to a specific stable version.** Please do not vendor HEAD if you can avoid it!
    * **Please introduce a vendored dependency in a dedicated commit to make review easier!**

3. Write the code change that imports and uses the new dependency.

4. Run the following in the root of the ksonnet repo:
    ```
    # View missing dependencies
    govendor list +missing

    # Download missing dependencies into vendor/
    govendor fetch -v +missing
    ```

5. Consider [cleaning up unused dependencies](#clean-up-unnecessary-dependencies) while you're at it.

6. Separate your `vendor/` and actual code changes into different pull requests:
  * Make a separate commit for your `vendor/` changes, and [cherry-pick it](https://git-scm.com/docs/git-cherry-pick) it into a new branch. Submit a PR for review.
  * After your `vendor/` changes are checked in, rebase your original code change and submit that PR.

7. Feel awesome for making a sizeable contribution to ksonnet! :tada:


### Pin a dependency to a specific version

Pinning a dependency avoids the issue of breaking updates.

We already do this for key libraries like `client-go` and `apimachinery`. To find the current version used for these particular packages, look in `vendor/vendor.json` for the `version` field (not `versionExact`).

Use that same version in the commands below:

```
# Pin all imported client-go packages to release v3.0
govendor fetch -v k8s.io/client-go/...@v3.0
```

*NOTE:* The command above may pull in new packages from `client-go`, which will
be imported at HEAD.  You need to re-run the above command until *all* imports are at the desired version.

If you make a mistake or miss some libraries, it is safe (and appropriate) to re-run `govendor fetch` with a different version.

### Clean up unnecessary dependencies

Before sending a non-trivial change for review, consider running the following command:

```
# See unused dependencies
govendor list +unused

# Remove unused dependencies
govendor remove +unused
```

## Test

Before making a PR, you should make sure to test your changes.

To do so, run `make test` in the root of the ksonnet repo, specifying any additional, custom Go flags with `$EXTRA_GO_FLAGS`.

## Make a Release

See our [release documentation](docs/release.md) for the process of creating a release.
