[![GoDoc](https://godoc.org/github.com/shazow/go-diff?status.svg)](https://godoc.org/github.com/shazow/go-diff)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/shazow/go-diff/master/LICENSE)
[![Build Status](https://travis-ci.org/shazow/go-diff.svg?branch=master)](https://travis-ci.org/shazow/go-diff)


# go-diff

Library for generating Git-style diff patchsets in Go.

Built to be used in the pure-Go implementation of the Git backend for
[sourcegraph's go-vcs](https://github.com/sourcegraph/go-vcs).

**Status**: Prototyping. None of the backends produce complete output yet, but
the difflib one is closest and least broken.


## Features

- Git-style patch headers for each file (are there other styles to support?).
- Bring your own diff algorithm by implementing the *Differ* interface.
- Includes a few diffing backends:
  * [pmezard/go-difflib](https://godoc.org/github.com/pmezard/go-difflib)
    (default)
  * [sergi/go-diff/diffmatchpatch](https://godoc.org/github.com/sergi/go-diff/diffmatchpatch)
    (offsets produced are inconsistent with git diff)
  * [mb0/diff](https://godoc.org/github.com/mb0/diff) (incomplete backend)


## Sponsors

Work on this package is sponsored by [Sourcegraph](https://sourcegraph.com/).


## License

MIT.
