stringer
========

This is a typewriter package for use with [gen](https://github.com/clipperhouse/gen), a tool for type-driven code generation. It is a fork of Rob Pike’s [tool](https://godoc.org/golang.org/x/tools/cmd/stringer) of the same name, which generates readable strings for consts.

It is one of gen’s built-in typewriters.

To use it:

```
go get -u github.com/clipperhouse/gen
```

Then, mark up a type in your package, for example:

```
// +gen stringer
type Pill int

const (
	Placebo Pill = iota
	Aspirin
	Ibuprofen
	Paracetamol
	Acetaminophen = Paracetamol
)
```

...and run `gen` on your package. You should see a new file named `mytype_stringer.go`. See the [gen docs](https://clipperhouse.github.io/gen/) for more information.
