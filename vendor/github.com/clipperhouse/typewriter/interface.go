// The typewriter package provides a framework for type-driven code generation. It implements the core functionality of gen.
//
// This package is primarily of interest to those who wish to extend gen with third-party functionality.
//
// More docs are available at https://clipperhouse.github.io/gen/typewriters.
package typewriter

import (
	"io"
)

// Interface is the interface to be implemented for code generation via gen
type Interface interface {
	Name() string
	// Imports is a slice of imports required for the type; each will be written into the imports declaration.
	Imports(t Type) []ImportSpec
	// Write writes to the body of the generated code, following package declaration and imports.
	Write(w io.Writer, t Type) error
}
