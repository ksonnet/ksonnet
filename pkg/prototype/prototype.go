// Copyright 2018 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package prototype

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/ksonnet/ksonnet/pkg/util/version"
	"github.com/pkg/errors"
)

const (
	// DefaultAPIVersion is the default api version for a prototype.
	DefaultAPIVersion = "0.0.1"

	// DefaultKind is the default kind for a prototype.
	DefaultKind = "ksonnet.io/prototype"

	apiVersionTag       = "@apiVersion"
	nameTag             = "@name"
	descriptionTag      = "@description"
	shortDescriptionTag = "@shortDescription"
	paramTag            = "@param"
	optParamTag         = "@optionalParam"
)

// Prototype is the JSON-serializable representation of a prototype
// specification.
type Prototype struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`

	// Unique identifier of the mixin library. The most reliable way to make a
	// name unique is to embed a domain you own into the name, as is commonly done
	// in the Java community.
	Name     string        `json:"name"`
	Params   ParamSchemas  `json:"params"`
	Template SnippetSchema `json:"template"`
	Version  string        `json:"-"` // Version of container package. Not serialized.
}

func (s *Prototype) validate() error {
	compatVer, _ := semver.Make(DefaultAPIVersion)
	ver, err := semver.Make(s.APIVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to parse version in app spec")
	} else if compatVer.Compare(ver) != 0 {
		return fmt.Errorf(
			"Current app uses unsupported spec version '%s' (this client only supports %s)",
			s.APIVersion,
			DefaultAPIVersion)
	}

	return nil
}

// Prototypes is a slice of pointer to `SpecificationSchema`.
type Prototypes []*Prototype

// SortByVersion sorts a prototype list by package version.
func (p Prototypes) SortByVersion() {
	less := func(i, j int) bool {
		vI, err := version.Make(p[i].Version)
		vJ, err2 := version.Make(p[j].Version)
		if err != nil || err2 != nil {
			// Fall back to lexical sort
			return p[i].Version < p[j].Version
		}

		return vI.LT(vJ)
	}

	sort.Slice(p, less)
}

// RequiredParams retrieves all parameters that are required by a prototype.
func (s *Prototype) RequiredParams() ParamSchemas {
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
func (s *Prototype) OptionalParams() ParamSchemas {
	opt := ParamSchemas{}
	for _, p := range s.Params {
		if p.Default != nil {
			opt = append(opt, p)
		}
	}

	return opt
}

// TemplateType represents the possible type of a prototype.
type TemplateType string

const (
	// YAML represents a prototype written in YAML.
	YAML TemplateType = "yaml"

	// JSON represents a prototype written in JSON.
	JSON TemplateType = "json"

	// Jsonnet represents a prototype written in Jsonnet.
	Jsonnet TemplateType = "jsonnet"
)

// ParseTemplateType attempts to parse a string as a `TemplateType`.
func ParseTemplateType(t string) (TemplateType, error) {
	switch strings.ToLower(t) {
	case "yaml", "yml":
		return YAML, nil
	case "json":
		return JSON, nil
	case "jsonnet":
		return Jsonnet, nil
	default:
		return "", fmt.Errorf("Unrecognized template type '%s'; must be one of: [yaml, json, jsonnet]", t)
	}
}

// SnippetSchema is the JSON-serializable representation of the TextMate snippet
// specification, as implemented by the Language Server Protocol.
type SnippetSchema struct {
	Prefix string `json:"prefix"`

	// Description describes what the prototype does.
	Description string `json:"description"`

	// ShortDescription briefly describes what the prototype does.
	ShortDescription string `json:"shortDescription"`

	// Various body types of the prototype. Follows the TextMate snippets syntax,
	// with several features disallowed. At least one of these is required to be
	// filled out.
	JSONBody    []string `json:"jsonBody"`
	YAMLBody    []string `json:"yamlBody"`
	JsonnetBody []string `json:"jsonnetBody"`
}

// Body attempts to retrieve the template body associated with some
// type `t`.
func (schema *SnippetSchema) Body(t TemplateType) (template []string, err error) {
	switch t {
	case YAML:
		template = schema.YAMLBody
	case JSON:
		template = schema.JSONBody
	case Jsonnet:
		template = schema.JsonnetBody
	default:
		return nil, fmt.Errorf("Unrecognized template type '%s'; must be one of: [yaml, json, jsonnet]", t)
	}

	if len(template) == 0 {
		available := schema.AvailableTemplates()
		err = fmt.Errorf("Template does not have a template for type '%s'. Available types: %s", t, available)
	}

	return
}

// AvailableTemplates returns the list of available `TemplateType`s this
// prototype implements.
func (schema *SnippetSchema) AvailableTemplates() (ts []TemplateType) {
	if len(schema.YAMLBody) != 0 {
		ts = append(ts, YAML)
	}

	if len(schema.JSONBody) != 0 {
		ts = append(ts, JSON)
	}

	if len(schema.JsonnetBody) != 0 {
		ts = append(ts, Jsonnet)
	}

	return
}

// ParamSchema is the JSON-serializable representation of a parameter provided
// to a prototype.
type ParamSchema struct {
	Name        string    `json:"name"`
	Alias       *string   `json:"alias"` // Optional.
	Description string    `json:"description"`
	Default     *string   `json:"default"` // `nil` only if the parameter is optional.
	Type        ParamType `json:"type"`
}

// Quote will parse a prototype parameter and quote it appropriately, so that it
// shows up correctly in Jsonnet source code. For example, `--image nginx` would
// likely need to show up as `"nginx"` in Jsonnet source.
func (ps *ParamSchema) Quote(value string) (string, error) {
	switch ps.Type {
	case Number:
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("Could not convert parameter '%s' to a number", ps.Name)
		}
		return value, nil
	case String:
		return fmt.Sprintf("\"%s\"", value), nil
	case NumberOrString:
		_, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return value, nil
		}
		return fmt.Sprintf("\"%s\"", value), nil
	case Array, Object:
		return value, nil
	default:
		return "", fmt.Errorf("Unknown param type for param '%s'", ps.Name)
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

		var info string
		if p.Default != nil {
			info = fmt.Sprintf(" [default: %s, type: %s]", *p.Default, p.Type.String())
		} else {
			info = fmt.Sprintf(" [type: %s]", p.Type.String())
		}

		// NOTE: If we don't add 1 here, the longest line will look like:
		// `--flag=<flag>Description is here.`
		space := strings.Repeat(" ", max-len(flag)+1)
		pretty := fmt.Sprintf(prefix + flag + space + p.Description + info)
		prettyFlags = append(prettyFlags, pretty)
	}

	return strings.Join(prettyFlags, "\n")
}
