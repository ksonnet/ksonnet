// Copyright 2018 The ksonnet authors
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
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type itemType int

const (
	itemDirective itemType = iota
	itemBody
)

var (
	reCommentText = regexp.MustCompile(`\s*//\s+(.*?)$`)
)

type item struct {
	typ itemType
	val string
}

type stateFn func(p *parser) stateFn

// JsonnetParse parses a source Jsonnet document into a Prototype.
func JsonnetParse(src string) (*Prototype, error) {
	items := parse(src)

	s := &Prototype{
		Kind: "ksonnet.io/prototype",
	}

	for item := range items {
		switch item.typ {
		case itemDirective:
			fn := newDirective(item.val)
			if err := fn(s); err != nil {
				return nil, err
			}
		case itemBody:
			s.Template.JsonnetBody = append(s.Template.JsonnetBody, item.val)
		}
	}

	return s, nil
}

func parse(src string) chan item {
	lines := strings.Split(src, "\n")

	p := parser{
		lines: lines,
		items: make(chan item),
	}

	go p.run()

	return p.items
}

type parser struct {
	lines   []string
	cur     int
	context string
	items   chan item
}

func (p *parser) run() {
	for state := parseLine; state != nil; {
		state = state(p)
	}

	close(p.items)
}

func (p *parser) emit(t itemType, value string) {
	p.items <- item{
		typ: t,
		val: value,
	}
}

func (p *parser) appendToBody() {
	p.emit(itemBody, p.lines[p.cur])
	p.cur++
}

func parseLine(p *parser) stateFn {
	if p.cur > len(p.lines)-1 {
		return nil
	}

	line := p.lines[p.cur]
	if reCommentText.MatchString(line) {
		p.context = line
		return parseComment(p)
	}

	p.appendToBody()

	if p.cur <= len(p.lines)-1 {
		return parseLine
	}
	return nil
}

func parseComment(p *parser) stateFn {
	s := commentText(p.context)
	if s != "" {
		p.context = s
		return parseMetadata(p)
	}

	return parseLine
}

func commentText(src string) string {
	match := reCommentText.FindAllStringSubmatch(src, 1)

	if len(match) == 1 && len(match[0]) == 2 {
		return match[0][1]
	}

	return ""
}

func isDirective(src string) bool {
	return strings.HasPrefix(src, "@")
}

func parseMetadata(p *parser) stateFn {
	if !isDirective(p.context) {
		p.appendToBody()
		return parseLine(p)
	}

	var buf bytes.Buffer
	buf.WriteString(strings.TrimPrefix(p.context, "@"))

	for {
		if p.cur+1 > len(p.lines)-1 {
			break
		}
		next := commentText(p.lines[p.cur+1])
		if isDirective(next) || next == "" {
			break
		}
		p.cur++

		buf.WriteString("\n")
		buf.WriteString(next)
	}

	p.emit(itemDirective, buf.String())
	p.cur++
	return parseLine(p)
}

type directive func(*Prototype) error

func newDirective(src string) directive {
	parts := strings.SplitN(src, " ", 2)

	if len(parts) != 2 {
		return func(*Prototype) error {
			return errors.Errorf("%q is not a valid directive")
		}
	}

	switch parts[0] {
	case "apiVersion":
		return applyDirective(parts[1])
	case "name":
		return nameDirective(parts[1])
	case "shortDescription":
		return shortDescriptionDirective(parts[1])
	case "description":
		return descriptionDirective(parts[1])
	case "param":
		return paramDirective(parts[1])
	case "optionalParam":
		return optParamDirective(parts[1])
	default:
		return func(*Prototype) error {
			return errors.Errorf("unknown prototype directive %q", parts[0])
		}
	}
}

func applyDirective(version string) func(*Prototype) error {
	return func(s *Prototype) error {
		s.APIVersion = version
		return nil
	}
}

func nameDirective(name string) func(*Prototype) error {
	return func(s *Prototype) error {
		s.Name = name
		return nil
	}
}

func descriptionDirective(description string) func(*Prototype) error {
	return func(s *Prototype) error {
		s.Template.Description = description
		return nil
	}
}

func shortDescriptionDirective(description string) func(*Prototype) error {
	return func(s *Prototype) error {
		s.Template.ShortDescription = description
		return nil
	}
}

func paramDirective(src string) func(*Prototype) error {
	return func(s *Prototype) error {
		split := strings.SplitN(src, " ", 3)
		if len(split) < 3 {
			return fmt.Errorf("param fields must have '<name> <type> <description>, but got:\n%s", src)
		}

		pt, err := parseParamType(split[1])
		if err != nil {
			return errors.Wrap(err, "invalid param tag")
		}

		s.Params = append(s.Params, &ParamSchema{
			Name:        split[0],
			Alias:       &split[0],
			Description: split[2],
			Default:     nil,
			Type:        pt,
		})

		return nil
	}
}

func optParamDirective(src string) func(*Prototype) error {
	return func(s *Prototype) error {
		split := strings.SplitN(src, " ", 4)
		if len(split) < 4 {
			return fmt.Errorf("optional param fields must have '<name> <type> <default-val> <description> (<default-val> currently cannot contain spaces), but got:\n%s", src)
		}

		pt, err := parseParamType(split[1])
		if err != nil {
			return err
		}

		s.Params = append(s.Params, &ParamSchema{
			Name:        split[0],
			Alias:       &split[0],
			Default:     &split[2],
			Description: split[3],
			Type:        pt,
		})

		return nil
	}
}
