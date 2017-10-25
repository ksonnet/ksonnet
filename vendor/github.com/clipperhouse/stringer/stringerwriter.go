package stringer

import (
	"io"

	"github.com/clipperhouse/typewriter"
)

func init() {
	typewriter.Register(&StringerWriter{})
}

type StringerWriter struct {
	g Generator
}

func (sw *StringerWriter) Name() string {
	return "stringer"
}

func (sw *StringerWriter) Imports(t typewriter.Type) []typewriter.ImportSpec {
	return []typewriter.ImportSpec{
		{Path: "fmt"},
	}
}

func (sw *StringerWriter) Write(w io.Writer, t typewriter.Type) error {
	_, found := t.FindTag(sw)

	if !found {
		return nil
	}

	if err := sw.g.parsePackageDir("./"); err != nil {
		return err
	}

	if err := sw.g.generate(t.Name); err != nil {
		return err
	}

	if _, err := io.Copy(w, &sw.g.buf); err != nil {
		return err
	}

	return nil
}
