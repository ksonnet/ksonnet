package set

import (
	"io"

	"github.com/clipperhouse/typewriter"
)

func init() {
	err := typewriter.Register(NewSetWriter())
	if err != nil {
		panic(err)
	}
}

type SetWriter struct{}

func NewSetWriter() *SetWriter {
	return &SetWriter{}
}

func (sw *SetWriter) Name() string {
	return "set"
}

func (sw *SetWriter) Imports(t typewriter.Type) (result []typewriter.ImportSpec) {
	// none
	return result
}

func (sw *SetWriter) Write(w io.Writer, t typewriter.Type) error {
	tag, found := t.FindTag(sw)

	if !found {
		// nothing to be done
		return nil
	}

	license := `// Set is a modification of https://github.com/deckarep/golang-set
// The MIT License (MIT)
// Copyright (c) 2013 Ralph Caraveo (deckarep@gmail.com)
`

	if _, err := w.Write([]byte(license)); err != nil {
		return err
	}

	tmpl, err := templates.ByTag(t, tag)

	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, t); err != nil {
		return err
	}

	return nil
}
