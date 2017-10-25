package typewriter

import (
	"fmt"
	"regexp"
	"strings"

	"go/types"
)

type Type struct {
	Pointer                      Pointer
	Name                         string
	Tags                         TagSlice
	comparable, numeric, ordered bool
	test                         test
	types.Type
}

type test bool

// a convenience for using bool in file name, see WriteAll
func (t test) String() string {
	if t {
		return "_test"
	}
	return ""
}

func (t Type) String() (result string) {
	return fmt.Sprintf("%s%s", t.Pointer.String(), t.Name)
}

// LongName provides a name that may be useful for generated names.
// For example, map[string]Foo becomes MapStringFoo.
func (t Type) LongName() string {
	s := strings.Replace(t.String(), "[]", "Slice[]", -1) // hacktastic

	r := regexp.MustCompile(`[\[\]{}*]`)
	els := r.Split(s, -1)

	var parts []string

	for _, s := range els {
		parts = append(parts, strings.Title(s))
	}

	return strings.Join(parts, "")
}

func (t Type) FindTag(tw Interface) (Tag, bool) {
	for _, tag := range t.Tags {
		if tag.Name == tw.Name() {
			return tag, true
		}
	}
	return Tag{}, false
}

// Pointer exists as a type to allow simple use as bool or as String, which returns *
type Pointer bool

func (p Pointer) String() string {
	if p {
		return "*"
	}
	return ""
}
