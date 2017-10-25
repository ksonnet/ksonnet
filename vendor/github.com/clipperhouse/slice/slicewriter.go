package slice

import (
	"io"
	"regexp"
	"strings"

	"github.com/clipperhouse/typewriter"
)

func init() {
	err := typewriter.Register(NewSliceWriter())
	if err != nil {
		panic(err)
	}
}

func SliceName(typ typewriter.Type) string {
	return typ.Name + "Slice"
}

type SliceWriter struct{}

func NewSliceWriter() *SliceWriter {
	return &SliceWriter{}
}

func (sw *SliceWriter) Name() string {
	return "slice"
}

func (sw *SliceWriter) Imports(typ typewriter.Type) (result []typewriter.ImportSpec) {
	// typewriter uses golang.org/x/tools/imports, depend on that
	return
}

func (sw *SliceWriter) Write(w io.Writer, typ typewriter.Type) error {
	tag, found := typ.FindTag(sw)

	if !found {
		return nil
	}

	if includeSortImplementation(tag.Values) {
		s := `// Sort implementation is a modification of http://golang.org/pkg/sort/#Sort
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at http://golang.org/LICENSE.

`
		w.Write([]byte(s))
	}

	// start with the slice template
	tmpl, err := templates.ByTag(typ, tag)

	if err != nil {
		return err
	}

	m := model{
		Type:      typ,
		SliceName: SliceName(typ),
	}

	if err := tmpl.Execute(w, m); err != nil {
		return err
	}

	for _, v := range tag.Values {
		var tp typewriter.Type

		if len(v.TypeParameters) > 0 {
			tp = v.TypeParameters[0]
		}

		m := model{
			Type:          typ,
			SliceName:     SliceName(typ),
			TypeParameter: tp,
			TagValue:      v,
		}

		tmpl, err := templates.ByTagValue(typ, v)

		if err != nil {
			return err
		}

		if err := tmpl.Execute(w, m); err != nil {
			return err
		}
	}

	if includeSortInterface(tag.Values) {
		tmpl, err := sortInterface.Parse()

		if err != nil {
			return err
		}

		if err := tmpl.Execute(w, m); err != nil {
			return err
		}
	}

	if includeSortImplementation(tag.Values) {
		tmpl, err := sortImplementation.Parse()

		if err != nil {
			return err
		}

		if err := tmpl.Execute(w, m); err != nil {
			return err
		}
	}

	return nil
}

func includeSortImplementation(values []typewriter.TagValue) bool {
	for _, v := range values {
		if strings.HasPrefix(v.Name, "SortBy") {
			return true
		}
	}
	return false
}

func includeSortInterface(values []typewriter.TagValue) bool {
	reg := regexp.MustCompile(`^Sort(Desc)?$`)
	for _, v := range values {
		if reg.MatchString(v.Name) {
			return true
		}
	}
	return false
}
