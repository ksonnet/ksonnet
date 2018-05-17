package diff

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/shazow/go-diff/difflib"
)

// DefaultDiffer returns a diffmatchpatch-based differ.
func DefaultDiffer() Differ {
	return difflib.New()
}

// Differ supplies a stream-based diffing algorithm implementation.
type Differ interface {
	// Diff writes a generated patch to out for the diff between a and b.
	Diff(out io.Writer, a io.ReadSeeker, b io.ReadSeeker) error
}

// Object is the minimum representation for generating diffs with git-style headers.
type Object struct {
	io.ReadSeeker
	// ID is the sha1 of the object
	ID [20]byte
	// Path is the root-relative path of the object
	Path string
	// Mode is the entry mode of the object
	Mode int
}

// ErrEmptyComparison is used when both src and dst are EmptyObjects
var ErrEmptyComparsion = errors.New("no objects to compare, both are empty")

// EmptyObject can be used when there is no corresponding src or dst entry, such as during deletions or creations.
var EmptyObject Object

// Writer writes diffs using a given Differ including git-style headers between each patch.
type Writer struct {
	io.Writer
	Differ
	SrcPrefix string
	DstPrefix string
}

// Diff writes the header and the generated the diff body when appropriate.
func (w *Writer) Diff(src, dst Object) error {
	// TODO: This can be optimized by skipping diff'ing during renames, mode changes, etc.
	if err := w.WriteHeader(src, dst); err != nil {
		return err
	}
	return w.WriteDiff(src, dst)
}

// WriteHeader writes only the header for the comparison between the src and dst Objects.
func (w *Writer) WriteHeader(src, dst Object) error {
	var srcPath, dstPath string
	if src == EmptyObject && dst == EmptyObject {
		return ErrEmptyComparsion
	}
	if src != EmptyObject {
		srcPath = filepath.Join(w.SrcPrefix, src.Path)
		dstPath = srcPath
	}
	if dst != EmptyObject {
		dstPath = filepath.Join(w.DstPrefix, dst.Path)
		if srcPath == "" {
			srcPath = dstPath
		}
	}
	// TODO: Detect renames?

	fmt.Fprintf(w, "diff --git %s %s\n", srcPath, dstPath)
	if src == EmptyObject {
		fmt.Fprintf(w, "new file mode %06d\n", dst.Mode)
		fmt.Fprintf(w, "index %x..%x\n", src.ID, dst.ID)
		fmt.Fprintf(w, "--- /dev/null\n")
		fmt.Fprintf(w, "+++ %s\n", dstPath)
	} else if dst == EmptyObject {
		fmt.Fprintf(w, "deleted file mode %06d\n", src.Mode)
		fmt.Fprintf(w, "index %x..%x\n", src.ID, dst.ID)
		fmt.Fprintf(w, "--- %s\n", srcPath)
		fmt.Fprintf(w, "+++ /dev/null\n")
	} else {
		fmt.Fprintf(w, "index %x..%x %06d\n", src.ID, dst.ID, dst.Mode)
		fmt.Fprintf(w, "--- %s\n", srcPath)
		fmt.Fprintf(w, "+++ %s\n", dstPath)
	}
	return nil
}

// WriteDiff performs a Diff between a and b and writes only the resulting diff. It does not write the header.
func (w *Writer) WriteDiff(a, b io.ReadSeeker) error {
	return w.Differ.Diff(w, a, b)
}
