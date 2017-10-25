package typewriter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// unlike the go build tool, the parser does not ignore . and _ files
var ignored = func(f os.FileInfo) bool {
	return !strings.HasPrefix(f.Name(), "_") && !strings.HasPrefix(f.Name(), ".")
}

func getPackages(directive string, conf *Config) ([]*Package, error) {
	// wrap filter with default filter
	filt := func(f os.FileInfo) bool {
		if conf.Filter != nil {
			return ignored(f) && conf.Filter(f)
		}
		return ignored(f)
	}

	// get the AST
	fset := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fset, "./", filt, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	var pkgs []*Package
	var typeCheckErrors []*TypeCheckError

	for _, a := range astPkgs {
		pkg, err := getPackage(fset, a, conf)

		if err != nil {
			err.ignored = conf.IgnoreTypeCheckErrors
			typeCheckErrors = append(typeCheckErrors, err)

			// if we have type check errors, and are not ignoring them, bail
			if err := combine(typeCheckErrors); err != nil && !conf.IgnoreTypeCheckErrors {
				return pkgs, err
			}
		}

		pkgs = append(pkgs, pkg)

		specs := getTaggedComments(a, directive)

		for s, c := range specs {
			pointer, tags, err := parse(fset, c, directive)

			if err != nil {
				return nil, err
			}

			// evaluate the annotated type
			typ, evalErr := pkg.Eval(pointer.String() + s.Name.Name)

			if evalErr != nil {
				// if we're not ignoring, can return immediately, normal behavior
				if !conf.IgnoreTypeCheckErrors {
					return pkgs, evalErr
				}

				// is it a TypeCheckError?
				tc, isTypeCheckError := evalErr.(*TypeCheckError)

				// if not a TypeCheckError, can return immediately, normal behavior
				if !isTypeCheckError {
					return pkgs, evalErr
				}

				tc.ignored = conf.IgnoreTypeCheckErrors
				tc.addPos(fset, s.Pos())

				typeCheckErrors = append(typeCheckErrors, tc)
			}

			// evaluate type parameters
			for _, tag := range tags {
				for i, val := range tag.Values {
					for _, item := range val.typeParameters {
						tp, evalErr := pkg.Eval(item.val)

						if evalErr != nil {
							// if we're not ignoring, can return immediately, normal behavior
							if !conf.IgnoreTypeCheckErrors {
								return pkgs, evalErr
							}

							// is it a TypeCheckError?
							tc, isTypeCheckError := evalErr.(*TypeCheckError)

							// if not a TypeCheckError, can return immediately, normal behavior
							if !isTypeCheckError {
								return pkgs, evalErr
							}

							tc.ignored = conf.IgnoreTypeCheckErrors
							tc.addPos(fset, item.pos+c.Slash)

							typeCheckErrors = append(typeCheckErrors, tc)
						}

						val.TypeParameters = append(val.TypeParameters, tp)
					}
					tag.Values[i] = val // mutate the original
				}
				typ.Tags = append(typ.Tags, tag)
			}

			typ.test = test(strings.HasSuffix(fset.Position(s.Pos()).Filename, "_test.go"))

			pkg.Types = append(pkg.Types, typ)
		}
	}

	// if we have type check errors, but are ignoring them, output as FYI
	if err := combine(typeCheckErrors); err != nil && conf.IgnoreTypeCheckErrors {
		fmt.Println(err)
	}

	return pkgs, nil
}

// getTaggedComments walks the AST and returns types which have directive comment
// returns a map of TypeSpec to directive
func getTaggedComments(pkg *ast.Package, directive string) map[*ast.TypeSpec]*ast.Comment {
	specs := make(map[*ast.TypeSpec]*ast.Comment)

	ast.Inspect(pkg, func(n ast.Node) bool {
		g, ok := n.(*ast.GenDecl)

		// is it a type?
		// http://golang.org/pkg/go/ast/#GenDecl
		if !ok || g.Tok != token.TYPE {
			// never mind, move on
			return true
		}

		if g.Lparen == 0 {
			// not parenthesized, copy GenDecl.Doc into TypeSpec.Doc
			g.Specs[0].(*ast.TypeSpec).Doc = g.Doc
		}

		for _, s := range g.Specs {
			t := s.(*ast.TypeSpec)

			if c := findAnnotation(t.Doc, directive); c != nil {
				specs[t] = c
			}
		}

		// no need to keep walking, we don't care about TypeSpec's children
		return false
	})

	return specs
}

// findDirective return the first line of a doc which contains a directive
// the directive and '//' are removed
func findAnnotation(doc *ast.CommentGroup, directive string) *ast.Comment {
	if doc == nil {
		return nil
	}

	// check lines of doc for directive
	for _, c := range doc.List {
		l := c.Text
		// does the line start with the directive?
		t := strings.TrimLeft(l, "/ ")
		if !strings.HasPrefix(t, directive) {
			continue
		}

		// remove the directive from the line
		t = strings.TrimPrefix(t, directive)

		// must be eof or followed by a space
		if len(t) > 0 && t[0] != ' ' {
			continue
		}

		return c
	}

	return nil
}

type parsr struct {
	lex       *lexer
	token     [2]item // two-token lookahead for parser.
	peekCount int
	fset      *token.FileSet
	offset    token.Pos
}

// next returns the next token.
func (p *parsr) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// backup backs the input stream up one token.
func (p *parsr) backup() {
	p.peekCount++
}

// peek returns but does not consume the next token.
func (p *parsr) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func (p *parsr) errorf(item item, format string, args ...interface{}) error {
	// some errors come with empty pos
	format = strings.TrimLeft(format, ":- ")
	// prepend position information (file name, line, column)
	format = fmt.Sprintf("%s: %s", p.fset.Position(item.pos+p.offset), format)
	return fmt.Errorf(format, args...)
}

func (p *parsr) unexpected(item item) error {
	return p.errorf(item, "unexpected '%v'", item.val)
}

func parse(fset *token.FileSet, comment *ast.Comment, directive string) (Pointer, TagSlice, error) {
	var pointer Pointer
	var tags TagSlice
	p := &parsr{
		lex:    lex(comment.Text),
		fset:   fset,
		offset: comment.Slash,
	}

	// to ensure no duplicate tags
	exists := make(map[string]struct{})

Loop:
	for {
		item := p.next()
		switch item.typ {
		case itemEOF:
			break Loop
		case itemError:
			err := p.errorf(item, item.val)
			return false, nil, err
		case itemCommentPrefix:
			// don't care, move on
			continue
		case itemDirective:
			// is it the directive we care about?
			if item.val != directive {
				return false, nil, nil
			}
			continue
		case itemPointer:
			// have we already seen a pointer?
			if pointer {
				err := p.errorf(item, "second pointer declaration")
				return false, nil, err
			}

			// have we already seen tags? pointer must be first
			if len(tags) > 0 {
				err := p.errorf(item, "pointer declaration must precede tags")
				return false, nil, err
			}

			pointer = true
		case itemTag:
			// we have an identifier, start a tag
			tag := Tag{
				Name: item.val,
			}

			// check for duplicate
			if _, seen := exists[tag.Name]; seen {
				err := p.errorf(item, "duplicate tag %q", tag.Name)
				return pointer, nil, err
			}

			// mark tag as previously seen
			exists[tag.Name] = struct{}{}

			// tag has values
			if p.peek().typ == itemColonQuote {
				p.next() // absorb the colonQuote
				negated, vals, err := parseTagValues(p)

				if err != nil {
					return false, nil, err
				}

				tag.Negated = negated
				tag.Values = vals
			}

			tags = append(tags, tag)
		default:
			return false, nil, p.unexpected(item)
		}
	}

	return pointer, tags, nil
}

func parseTagValues(p *parsr) (bool, []TagValue, error) {
	var negated bool
	var vals []TagValue

	for {
		item := p.next()

		switch item.typ {
		case itemError:
			err := p.errorf(item, item.val)
			return false, nil, err
		case itemEOF:
			// shouldn't happen within a tag
			err := p.errorf(item, "expected a close quote")
			return false, nil, err
		case itemMinus:
			if len(vals) > 0 {
				err := p.errorf(item, "negation must precede tag values")
				return false, nil, err
			}
			negated = true
		case itemTagValue:
			val := TagValue{
				Name: item.val,
			}

			if p.peek().typ == itemTypeParameter {
				tokens, err := parseTypeParameters(p)
				if err != nil {
					return false, nil, err
				}
				val.typeParameters = tokens
			}

			vals = append(vals, val)
		case itemCloseQuote:
			// we're done
			return negated, vals, nil
		default:
			return false, nil, p.unexpected(item)
		}
	}
}

func parseTypeParameters(p *parsr) ([]item, error) {
	var result []item

	for {
		item := p.next()

		switch item.typ {
		case itemTypeParameter:
			result = append(result, item)
		default:
			p.backup()
			return result, nil
		}
	}
}
