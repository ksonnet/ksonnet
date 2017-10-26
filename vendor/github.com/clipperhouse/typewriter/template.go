package typewriter

import (
	"fmt"
	"strings"

	"text/template"
)

// Template includes the text of a template as well as requirements for the types to which it can be applied.
// +gen * slice:"Where"
type Template struct {
	Name, Text     string
	FuncMap        map[string]interface{}
	TypeConstraint Constraint
	// Indicates both the number of required type parameters, and the constraints of each (if any)
	TypeParameterConstraints []Constraint
}

// Parse parses (converts) a typewriter.Template to a *template.Template
func (tmpl *Template) Parse() (*template.Template, error) {
	return template.New(tmpl.Name).Funcs(tmpl.FuncMap).Parse(tmpl.Text)
}

// TryTypeAndValue verifies that a given Type and TagValue satisfy a Template's type constraints.
func (tmpl *Template) TryTypeAndValue(t Type, v TagValue) error {
	if err := tmpl.TypeConstraint.TryType(t); err != nil {
		return fmt.Errorf("cannot apply %s to %s: %s", v.Name, t, err)
	}

	if len(tmpl.TypeParameterConstraints) != len(v.TypeParameters) {
		return fmt.Errorf("%s requires %d type parameters", v.Name, len(v.TypeParameters))
	}

	for i := range v.TypeParameters {
		c := tmpl.TypeParameterConstraints[i]
		tp := v.TypeParameters[i]
		if err := c.TryType(tp); err != nil {
			return fmt.Errorf("cannot apply %s on %s: %s", v, t, err)
		}
	}

	return nil
}

// Funcs assigns non standard functions used in the template
func (ts TemplateSlice) Funcs(FuncMap map[string]interface{}) {
	for _, tmpl := range ts {
		tmpl.FuncMap = FuncMap
	}
}

// ByTag attempts to locate a template which meets type constraints, and parses it.
func (ts TemplateSlice) ByTag(t Type, tag Tag) (*template.Template, error) {
	// templates which might work
	candidates := ts.Where(func(tmpl *Template) bool {
		return strings.EqualFold(tmpl.Name, tag.Name)
	})

	if len(candidates) == 0 {
		err := fmt.Errorf("could not find template for %q", tag.Name)
		return nil, err
	}

	// try to find one that meets type constraints
	for _, tmpl := range candidates {
		if err := tmpl.TypeConstraint.TryType(t); err == nil {
			// eagerly return on success
			return tmpl.Parse()
		}
	}

	// send back the first error message; not great but OK most of the time
	return nil, candidates[0].TypeConstraint.TryType(t)
}

// ByTagValue attempts to locate a template which meets type constraints, and parses it.
func (ts TemplateSlice) ByTagValue(t Type, v TagValue) (*template.Template, error) {
	// a bit of poor-man's type resolution here

	// templates which might work
	candidates := ts.Where(func(tmpl *Template) bool {
		return strings.EqualFold(tmpl.Name, v.Name)
	})

	if len(candidates) == 0 {
		err := fmt.Errorf("%s is unknown", v.Name)
		return nil, err
	}

	// try to find one that meets type constraints
	for _, tmpl := range candidates {
		if err := tmpl.TryTypeAndValue(t, v); err == nil {
			// eagerly return on success
			return tmpl.Parse()
		}
	}

	// send back the first error message; not great but OK most of the time
	return nil, candidates[0].TryTypeAndValue(t, v)
}
