package prototype

import (
	"fmt"
	"strings"
)

//
// NOTE: These members would ordinarily be private and exposed by interfaces,
// but because Go requires public structs for un/marshalling, it is more
// convenient to simply expose all of them.
//

// SpecificationSchema is the JSON-serializable representation of a prototype
// specification.
type SpecificationSchema struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`

	// Unique identifier of the mixin library. The most reliable way to make a
	// name unique is to embed a domain you own into the name, as is commonly done
	// in the Java community.
	Name     string        `json:"name"`
	Params   ParamSchemas  `json:"params"`
	Template SnippetSchema `json:"template"`
}

// RequiredParams retrieves all parameters that are required by a prototype.
func (s *SpecificationSchema) RequiredParams() ParamSchemas {
	reqd := ParamSchemas{}
	for _, p := range s.Params {
		if p.Default == nil {
			reqd = append(reqd, p)
		}
	}

	return reqd
}

// OptionalParams retrieves all parameters that can optionally be provided to a
// prototype.
func (s *SpecificationSchema) OptionalParams() ParamSchemas {
	opt := ParamSchemas{}
	for _, p := range s.Params {
		if p.Default != nil {
			opt = append(opt, p)
		}
	}

	return opt
}

// SnippetSchema is the JSON-serializable representation of the TextMate snippet
// specification, as implemented by the Language Server Protocol.
type SnippetSchema struct {
	Prefix string `json:"prefix"`

	// Description describes what the prototype does.
	Description string `json:"description"`

	// Body of the prototype. Follows the TextMate snippets syntax, with several
	// features disallowed.
	Body []string `json:"body"`
}

// ParamSchema is the JSON-serializable representation of a parameter provided to a prototype.
type ParamSchema struct {
	Name        string  `json:"name"`
	Alias       *string `json:"alias"` // Optional.
	Description string  `json:"description"`
	Default     *string `json:"default"` // `nil` only if the parameter is optional.
}

// RequiredParam constructs a required parameter, i.e., a parameter that is
// meant to be required by some prototype, somewhere.
func RequiredParam(name, alias, description string) *ParamSchema {
	return &ParamSchema{
		Name:        name,
		Alias:       &alias,
		Description: description,
		Default:     nil,
	}
}

// OptionalParam constructs an optional parameter, i.e., a parameter that is
// meant to be optionally provided to some prototype, somewhere.
func OptionalParam(name, alias, description, defaultVal string) *ParamSchema {
	return &ParamSchema{
		Name:        name,
		Alias:       &alias,
		Description: description,
		Default:     &defaultVal,
	}
}

// ParamSchemas is a slice of `ParamSchema`
type ParamSchemas []*ParamSchema

// PrettyString creates a prettified string representing a collection of
// parameters.
func (ps ParamSchemas) PrettyString(prefix string) string {
	if len(ps) == 0 {
		return "  [none]"
	}

	flags := []string{}
	for _, p := range ps {
		alias := p.Name
		if p.Alias != nil {
			alias = *p.Alias
		}
		flags = append(flags, fmt.Sprintf("--%s=<%s>", p.Name, alias))
	}

	max := 0
	for _, flag := range flags {
		if flagLen := len(flag); max < flagLen {
			max = flagLen
		}
	}

	prettyFlags := []string{}
	for i := range flags {
		p := ps[i]
		flag := flags[i]

		defaultVal := ""
		if p.Default != nil {
			defaultVal = fmt.Sprintf(" [default: %s]", *p.Default)
		}

		// NOTE: If we don't add 1 here, the longest line will look like:
		// `--flag=<flag>Description is here.`
		space := strings.Repeat(" ", max-len(flag)+1)
		pretty := fmt.Sprintf(prefix + flag + space + p.Description + defaultVal)
		prettyFlags = append(prettyFlags, pretty)
	}

	return strings.Join(prettyFlags, "\n")
}
