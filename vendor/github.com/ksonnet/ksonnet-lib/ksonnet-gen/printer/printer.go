package printer

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/parser"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/pkg/errors"
)

const (
	space       = byte(' ')
	tab         = byte('\t')
	newline     = byte('\n')
	comma       = byte(',')
	doubleQuote = byte('"')
	singleQuote = byte('\'')

	syntaxSugar = '+'
)

// quoteMode is an enumeration specifying how a jsonnet string should be quoted.
type quoteMode int

const (
	quoteModeNone quoteMode = iota
	quoteModeSingle
	quoteModeDouble
	quoteModeBlock
)

// Fprint prints a node to the supplied writer using the default
// configuration.
func Fprint(output io.Writer, node ast.Node) error {
	return DefaultConfig.Fprint(output, node)
}

// DefaultConfig is a default configuration.
var DefaultConfig = Config{
	IndentSize:  2,
	PadArrays:   false,
	PadObjects:  true,
	SortImports: true,
}

// IndentMode is the indent mode for Config.
type IndentMode int

const (
	// IndentModeSpace indents with spaces.
	IndentModeSpace IndentMode = iota
	// IndentModeTab indents with tabs.
	IndentModeTab
)

// Config is a configuration for the printer.
type Config struct {
	IndentSize  int
	IndentMode  IndentMode
	PadArrays   bool
	PadObjects  bool
	SortImports bool
}

// Fprint prints a node to the supplied writer.
func (c *Config) Fprint(output io.Writer, node ast.Node) error {
	p := printer{cfg: *c}

	p.print(node)

	if p.err != nil {
		return errors.Wrap(p.err, "output")
	}

	_, err := output.Write(p.output)
	return err
}

type printer struct {
	cfg Config

	output      []byte
	indentLevel int
	inFunction  bool

	err error
}

func (p *printer) indent() {
	if len(p.output) == 0 {
		return
	}

	r := p.indentLevel
	var ch byte
	if p.cfg.IndentMode == IndentModeTab {
		ch = tab
	} else {
		ch = space
		r = r * p.cfg.IndentSize
	}

	last := p.output[len(p.output)-1]
	if last == newline {
		pre := bytes.Repeat([]byte{ch}, r)
		p.output = append(p.output, pre...)
	}
}

func (p *printer) writeByte(ch byte, n int) {
	if p.err != nil {
		return
	}

	for i := 0; i < n; i++ {
		p.output = append(p.output, ch)
	}

	p.indent()
}

func (p *printer) writeString(s string) {
	for _, b := range []byte(s) {
		p.writeByte(b, 1)
	}
}

func (p *printer) writeStringNoIndent(s string) {
	for _, b := range []byte(s) {
		p.output = append(p.output, b)
	}
}

// detectQuoteMode decides what quote style to use for serializing
// a jsonnet string, with logic similar to that of `jsonnet fmt`.
//
// Briefly, single quotes are preferred, unless escaping can be avoided
// by using double quotes.
// In cases where both single and double quotes are detected,
// the status-quo is preferred (as specificed by `kind`).
func detectQuoteMode(s string, kind ast.LiteralStringKind) quoteMode {
	hasSingle := strings.ContainsRune(s, '\'')
	hasDouble := strings.ContainsRune(s, '"')

	switch kind {
	default:
		return quoteModeNone
	case ast.StringSingle:
		// Go with single unless there's only single quotes already.
		useSingle := !(hasSingle && !hasDouble)
		if useSingle {
			return quoteModeSingle
		}

	case ast.StringDouble:
		// Cases:
		// 1. [" 'abc' "] -> [" 'abc' "]
		// 2. [" \"abc\" "] -> [' "abc" ']
		// 3. [" 'abc' \"123\" "] -> [" 'abc' \"123\" "]
		// 4. [" abc "] -> [' abc ']
		useDouble := (hasSingle && !hasDouble) ||
			(hasSingle && hasDouble)
		if useDouble {
			return quoteModeDouble
		}
	case ast.StringBlock:
		return quoteModeBlock
	}

	return quoteModeNone
}

func unquote(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}

	sb := strings.Builder{}
	sb.Grow(len(s))

	tail := s
	for len(tail) > 0 {
		c, width := utf8.DecodeRuneInString(tail)

		switch c {
		case utf8.RuneError:
			tail = tail[width:]
			continue // Skip this character
		case '"':
			// strconv.UnquoteChar won't allow a bare quote in double-quote mode, but we will
			sb.WriteRune(c)
			tail = tail[width:]
		default:
			c, _, t2, err := strconv.UnquoteChar(tail, byte('"'))
			if err != nil {
				// Skip character. Ensure we move forward.
				tail = tail[width:]
				continue
			}
			tail = t2
			sb.WriteRune(c)
		}
	}

	return sb.String()
}

// quote returns a single or double quoted string, escaped for jsonnet.
// This function does *not* protect against double-escaping. Instead use `stringQuote`.
func quote(s string, useSingle bool) string {
	var quote rune
	if useSingle {
		quote = '\''
	} else {
		quote = '"'
	}

	sb := strings.Builder{}
	sb.Grow(len(s) + 2)

	sb.WriteRune(quote)

	for _, c := range s {
		switch c {
		case '\'':
			if useSingle {
				sb.WriteString("\\'")
			} else {
				sb.WriteRune(c)
			}
		case '"':
			if !useSingle {
				sb.WriteString("\\\"")
			} else {
				sb.WriteRune(c)
			}
		default:
			q := strconv.QuoteRune(c) // This is returned with unneeded quotes
			sb.WriteString(q[1 : len(q)-1])
		}
	}

	sb.WriteRune(quote)
	return sb.String()
}

// stringQuote returns a quoted jsonnet string ready for serialization.
// Appropriate measures are taken to avoid double-escaping any control characters.
// `useSingle` specifies whether to use single quotes, otherwise double-quotes are used.
// Note the following characters will be escaped (with leading backslash): "'\/bfnrt
// Quotes (single or double) will only be escaped to avoid conflict with the enclosing quote type.
func stringQuote(s string, useSingle bool) string {
	unquoted := unquote(s) // Avoid double-escaping control characters

	return quote(unquoted, useSingle)
}

// printer prints a node.
// nolint: gocyclo
func (p *printer) print(n interface{}) {
	if p.err != nil {
		return
	}

	if n == nil {
		return
	}

	switch t := n.(type) {
	default:
		p.err = errors.Errorf("unknown node type: (%T) %#v", n, n)
		return
	case *ast.Apply:
		p.handleApply(t)
	case ast.Arguments:
		p.handleArguments(t)
	case *ast.ApplyBrace:
		p.print(t.Left)
		p.writeByte(space, 1)
		p.print(t.Right)
	case *ast.Array:
		oneLine := false
		if loc := t.NodeBase.Loc(); loc != nil && loc.Begin.Line == loc.End.Line {
			oneLine = true
		}
		shouldPad := oneLine && p.cfg.PadArrays && len(t.Elements) > 0

		p.writeString("[")
		if !oneLine {
			p.indentLevel++
			p.writeByte(newline, 1)
		}
		if shouldPad {
			p.writeByte(space, 1)
		}

		for i := 0; i < len(t.Elements); i++ {
			p.print(t.Elements[i])

			if i < len(t.Elements)-1 {
				if oneLine {
					p.writeString(", ")
				} else {
					p.writeString(",")
					p.writeByte(newline, 1)
				}
			}
		}

		// Trailing comma
		if !oneLine && len(t.Elements) > 0 {
			p.writeByte(comma, 1)
		}

		if !oneLine {
			p.indentLevel--
			p.writeByte(newline, 1)
		}
		if shouldPad {
			p.writeByte(space, 1)
		}

		p.writeString("]")
	case *ast.ArrayComp:
		p.handleArrayComp(t)
	case *ast.Assert:
		p.writeString("assert ")
		p.print(t.Cond)

		if t.Message != nil {
			p.writeString(" : ")
			p.print(t.Message)
		}

		p.writeString("; ")
		p.print(t.Rest)
	case *ast.Binary:
		oneLine := true
		leftLoc := t.Left.Loc()
		rightLoc := t.Right.Loc()

		if leftLoc != nil && rightLoc != nil {
			oneLine = leftLoc.End.Line == rightLoc.Begin.Line
		}

		p.print(t.Left)
		p.writeByte(space, 1)

		p.writeString(t.Op.String())

		if !oneLine {
			p.writeByte(newline, 1)
		} else {
			p.writeByte(space, 1)
		}

		p.print(t.Right)
	case *ast.Conditional:
		p.handleConditional(t)
	case *ast.Dollar:
		p.writeString("$")
	case *ast.Error:
		p.writeString("error ")
		p.print(t.Expr)
	case *ast.Function:
		p.writeString("function")
		p.addMethodSignature(t)
		p.writeString(" ")
		p.print(t.Body)
	case ast.IfSpec:
		p.writeString("if ")
		p.print(t.Expr)
	case *ast.Import:
		p.writeString("import ")
		p.print(t.File)
	case *ast.ImportStr:
		p.writeString("importstr ")
		p.print(t.File)
	case *ast.Index:
		p.handleIndex(t)
	case *ast.InSuper:
		p.print(t.Index)
		p.writeString(" in super")
	case *ast.Local:
		p.handleLocal(t)
	case *ast.Object:
		isSingleLine := p.isObjectSingleLine(t)
		shouldPad := isSingleLine && p.cfg.PadObjects && len(t.Fields) > 0
		needTrailingComma := !isSingleLine && len(t.Fields) > 0
		p.writeString("{")
		if shouldPad {
			p.writeByte(space, 1)
		}

		for i, field := range t.Fields {
			if !p.isObjectSingleLine(t) {
				p.indentLevel++
				p.writeByte(newline, 1)
			}

			p.print(field)
			if i < len(t.Fields)-1 {
				p.writeByte(comma, 1)
				if p.isObjectSingleLine(t) {
					p.writeByte(space, 1)
				}
			}

			if !p.isObjectSingleLine(t) {
				p.indentLevel--
			}
		}

		if needTrailingComma {
			p.writeByte(comma, 1)
		}

		// write an extra newline at the end
		if !p.isObjectSingleLine(t) {
			p.writeByte(newline, 1)
		}

		if shouldPad {
			p.writeByte(space, 1)
		}
		p.writeString("}")
	case *astext.Object:
		isSingleLine := p.isObjectSingleLine(t)
		shouldPad := isSingleLine && p.cfg.PadObjects && len(t.Fields) > 0
		needTrailingComma := !isSingleLine && len(t.Fields) > 0
		p.writeString("{")
		if shouldPad {
			p.writeByte(space, 1)
		}

		for i, field := range t.Fields {
			if !p.isObjectSingleLine(t) {
				p.indentLevel++
				p.writeByte(newline, 1)
			}

			p.print(field)
			if i < len(t.Fields)-1 {
				p.writeByte(comma, 1)
				if p.isObjectSingleLine(t) {
					p.writeByte(space, 1)
				}
			}

			if !p.isObjectSingleLine(t) {
				p.indentLevel--
			}
		}

		if needTrailingComma {
			p.writeByte(comma, 1)
		}

		// write an extra newline at the end
		if !p.isObjectSingleLine(t) {
			p.writeByte(newline, 1)
		}

		if shouldPad {
			p.writeByte(space, 1)
		}
		p.writeString("}")
	case *ast.ObjectComp:
		p.handleObjectComp(t)
	case astext.ObjectField, ast.ObjectField:
		p.handleObjectField(t)
	case *ast.LiteralBoolean:
		if t.Value {
			p.writeString("true")
		} else {
			p.writeString("false")
		}
	case *ast.LiteralString:
		qm := detectQuoteMode(t.Value, t.Kind)

		switch t.Kind {
		default:
			p.err = errors.Errorf("unknown string literal kind %#v", t.Kind)
			return
		case ast.StringSingle, ast.StringDouble:
			useSingle := (qm != quoteModeDouble)

			// Unescape newlines and tabs. The jsonnet parser will escape
			// these during parsing.
			val := strings.Replace(t.Value, "\\n", "\n", -1)
			val = strings.Replace(val, "\\t", "\t", -1)

			quoted := stringQuote(val, useSingle)
			p.writeString(quoted)
		case ast.StringBlock:
			p.writeString("|||")
			p.writeByte(newline, 1)
			p.writeString(t.Value)
			p.writeStringNoIndent("\n|||")
		case ast.VerbatimStringDouble:
			p.writeString("@")
			p.writeByte(doubleQuote, 1)
			p.writeString(t.Value)
			p.writeByte(doubleQuote, 1)
		case ast.VerbatimStringSingle:
			p.writeString("@")
			p.writeByte(singleQuote, 1)
			p.writeString(t.Value)
			p.writeByte(singleQuote, 1)
		}

	case *ast.LiteralNumber:

		p.writeString(t.OriginalString)
	case *ast.Parens:
		p.writeString("(")
		p.print(t.Inner)
		p.writeString(")")
	case *ast.Self:
		p.writeString("self")
	case *ast.Slice:
		p.print(t.Target)
		p.writeString("[")
		if t.BeginIndex != nil {
			p.print(t.BeginIndex)
		}
		p.writeString(":")
		if t.EndIndex != nil {
			p.print(t.EndIndex)
		}
		if t.Step != nil {
			p.writeString(":")
			p.print(t.Step)
		}
		p.writeString("]")
	case *ast.Unary:
		p.writeString(t.Op.String())
		p.print(t.Expr)
	case *ast.Var:
		p.writeString(string(t.Id))
	case *ast.LiteralNull:
		p.writeString("null")
	case *ast.SuperIndex:
		p.writeString("super.")
		p.writeString(string(*t.Id))
	}
}

func (p *printer) handleApply(a *ast.Apply) {
	switch a.Target.(type) {
	default:
		p.writeString("function")
		p.writeString("(")
		p.print(a.Arguments)
		p.writeString(")")
		p.writeByte(space, 1)
		p.print(a.Target)
		if a.TailStrict {
			p.writeString(" tailstrict")
		}
	case *ast.Apply, *ast.Index, *ast.Self, *ast.Var, *ast.Parens:
		p.print(a.Target)
		p.writeString("(")
		p.print(a.Arguments)
		p.writeString(")")
		if a.TailStrict {
			p.writeString(" tailstrict")
		}
	}
}

func (p *printer) handleArguments(a ast.Arguments) {
	argCount := 0

	for i, arg := range a.Positional {
		argCount++
		p.print(arg)
		if i < len(a.Positional)-1 {
			p.writeByte(comma, 1)
			p.writeByte(space, 1)
		}
	}

	if argCount > 0 && len(a.Named) > 0 {
		p.writeByte(comma, 1)
		p.writeByte(space, 1)
	}

	for i, named := range a.Named {
		p.writeString(string(named.Name))
		p.writeString("=")
		p.print(named.Arg)
		if i < len(a.Named)-1 {
			p.writeByte(comma, 1)
			p.writeByte(space, 1)
		}
	}
}

func (p *printer) handleConditional(c *ast.Conditional) {
	p.writeString("if ")
	p.print(c.Cond)

	p.writeString(" then ")
	p.print(c.BranchTrue)

	if c.BranchFalse != nil {
		p.writeString(" else ")
		p.print(c.BranchFalse)
	}
}

func (p *printer) writeComment(c *astext.Comment) {
	if c == nil {
		return
	}

	lines := strings.Split(c.Text, "\n")
	for _, line := range lines {
		p.writeString("//")
		if len(line) > 0 {
			p.writeByte(space, 1)
		}
		p.writeString(strings.TrimSpace(line))
		p.writeByte(newline, 1)
	}
}

func (p *printer) handleIndex(i *ast.Index) {
	if i == nil {
		p.err = errors.New("index is nil")
		return
	}

	p.print(i.Target)
	p.indexID(i)
}

func (p *printer) handleLocal(l *ast.Local) {
	p.writeString("local ")

	for i, bind := range l.Binds {
		p.writeString(string(bind.Variable))
		p.addMethodSignature(bind.Fun)

		switch t := bind.Body.(type) {
		default:
			p.writeString(" = ")
			p.print(bind.Body)
		case *ast.Function:
			p.addMethodSignature(t)
			p.writeString(" =")

			switch t.Body.(type) {
			default:
				p.writeString(" ")
				p.print(t.Body)
			case *ast.Local:
				p.indentLevel++
				p.writeByte(newline, 1)
				p.print(t.Body)
				p.indentLevel--
			}
		}

		if l := len(l.Binds); l > 1 {
			if i < l-1 {
				p.writeString(", ")
			}
		}
	}

	c := 1
	if _, ok := l.Body.(*ast.Local); !ok {
		c = 2
	}
	p.writeString(";")
	p.writeByte(newline, c)

	p.print(l.Body)
}

func (p *printer) handleLocalFunction(f *ast.Function) {
	p.addMethodSignature(f)
	p.writeString(" =")
	switch f.Body.(type) {
	default:
		p.writeByte(space, 1)
		p.print(f.Body)
	case *ast.Local:
		p.indentLevel++
		p.writeByte(newline, 1)
		p.print(f.Body)
		p.indentLevel--
	}
}

// shouldUnquoteFieldID determines whether s can be an unquoted field identifier.
// Things that can't be unquoted: keywords, expressions.
func shouldUnquoteFieldID(s string) bool {
	// Try to pose the string as a raw (unquoted) object field identifier.
	// If this succeeds, quotes are not needed.
	toParse := fmt.Sprintf("{%s: 'value'}", s)
	tokens, err := parser.Lex("", toParse)
	if err != nil {
		return false
	}

	if len(tokens) < 3 {
		return false
	}

	// The following pattern would show s tokenized entirely as a single identifier.
	// NOTE we resort to string matching as `token.kind` is not exported.
	idTokenMatch := fmt.Sprintf("(IDENTIFIER, \"%s\")", s)

	if tokens[1].String() == idTokenMatch &&
		tokens[2].String() == "\":\"" {
		return true
	}

	return false
}

// reID matches `id` as defined in the jsonnet spec
var reID = regexp.MustCompile(`^[_a-zA-Z][_a-zA-Z0-9]*$`)

func (p *printer) fieldID(kind ast.ObjectFieldKind, expr1 ast.Node, id *ast.Identifier) {
	if expr1 != nil {
		switch t := expr1.(type) {
		default:
			p.print(t)
			return
		case *ast.LiteralString:
			qm := detectQuoteMode(t.Value, t.Kind)
			useSingle := (qm == quoteModeSingle)
			quoted := stringQuote(t.Value, useSingle)

			// Block quotes (|||) are always retained:
			if qm == quoteModeBlock {
				sb := strings.Builder{}
				sb.WriteString("|||")
				sb.WriteByte(newline)
				// Indent the value using BlockIndent, but be aware that our caller
				// will also be indenting using indentLevel. Remove those many bytes.
				padding := strings.Repeat(" ", p.indentLevel*p.cfg.IndentSize)
				replaceCount := strings.Count(t.Value, "\n") - 1 // Replace all but the last newline
				indented := padding + strings.Replace(t.Value, "\n", ("\n"+padding), replaceCount)
				sb.WriteString(indented)
				sb.WriteString("|||")
				p.writeString(sb.String())
				return
			}

			// Return identity if quotes aren't strictly necessary,
			// that is it could pass for a bare identifier without ambiguity.
			// This is not the case for keywords, expressions,
			// e.g. 'guestbook-ui' or 'error'.
			switch kind {
			case ast.ObjectFieldID, ast.ObjectFieldStr:
				if shouldUnquoteFieldID(t.Value) {
					p.writeString(t.Value)
					return
				}
			}

			// Example where quotes are needed: kind==ObjectFieldExpr
			p.writeString(quoted)
			return
		}
	}

	if id != nil {
		p.writeString(string(*id))
		return
	}
}

func (p *printer) handleObjectComp(oc *ast.ObjectComp) {
	p.writeString("{")
	p.indentLevel++
	p.writeByte(newline, 1)
	p.handleObjectField(oc)
	p.indentLevel--
	p.writeByte(newline, 1)
	p.writeString("}")
}

func (p *printer) handleArrayComp(ac *ast.ArrayComp) {
	p.writeString("[")
	p.indentLevel++
	p.writeByte(newline, 1)
	p.print(ac.Body)
	p.writeByte(newline, 1)
	p.forSpec(ac.Spec)
	p.indentLevel--
	p.writeByte(newline, 1)
	p.writeString("]")
}

func (p *printer) forSpec(spec ast.ForSpec) {
	if spec.Outer != nil {
		p.forSpec(*spec.Outer)
		p.writeByte(newline, 1)
	}

	if spec.VarName != "" {
		p.writeString(fmt.Sprintf("for %s in ", string(spec.VarName)))
		p.print(spec.Expr)

		for _, ifSpec := range spec.Conditions {
			p.writeByte(newline, 1)
			p.print(ifSpec)
		}
	}
}

func (p *printer) handleObjectField(n interface{}) {
	var ofHide ast.ObjectFieldHide
	var ofKind ast.ObjectFieldKind
	var ofID *ast.Identifier
	var ofMethod *ast.Function
	var ofSugar bool
	var ofExpr1 ast.Node
	var ofExpr2 ast.Node
	var ofExpr3 ast.Node

	var forSpec ast.ForSpec

	switch t := n.(type) {
	default:
		p.err = errors.Errorf("unknown object field type %T", t)
		return
	case ast.ObjectField:
		ofHide = t.Hide
		ofKind = t.Kind
		ofID = t.Id
		ofMethod = t.Method
		ofSugar = t.SuperSugar
		ofExpr1 = t.Expr1
		ofExpr2 = t.Expr2
		ofExpr3 = t.Expr3
	case astext.ObjectField:
		ofHide = t.Hide
		ofKind = t.Kind
		ofID = t.Id
		ofMethod = t.Method
		ofSugar = t.SuperSugar
		ofExpr1 = t.Expr1
		ofExpr2 = t.Expr2
		ofExpr3 = t.Expr3
		p.writeComment(t.Comment)
	case *ast.ObjectComp:
		field := t.Fields[0]
		ofHide = field.Hide
		ofKind = field.Kind
		ofMethod = field.Method
		ofExpr1 = field.Expr1
		ofSugar = field.SuperSugar
		ofExpr2 = field.Expr2

		forSpec = t.Spec
	}

	var fieldType string

	switch ofHide {
	default:
		p.err = errors.Errorf("unknown Hide type %#v", ofHide)
		return
	case ast.ObjectFieldHidden:
		fieldType = "::"
	case ast.ObjectFieldVisible:
		fieldType = ":::"
	case ast.ObjectFieldInherit:
		fieldType = ":"
	}

	switch ofKind {
	default:
		p.err = errors.Errorf("unknown Kind type (%T) %#v", ofKind, ofKind)
		return
	case ast.ObjectAssert:
		p.writeString("assert ")
		p.print(ofExpr2)
		if ofExpr3 != nil {
			p.writeString(": ")
			p.print(ofExpr3)
		}
	case ast.ObjectFieldID:
		p.fieldID(ofKind, ofExpr1, ofID)
		if ofMethod != nil {
			p.addMethodSignature(ofMethod)
		}

		if ofSugar {
			p.writeByte(syntaxSugar, 1)
		}

		p.writeString(fieldType)

		if isLocal(ofExpr2) {
			p.indentLevel++
			p.writeByte(newline, 1)
			p.print(ofExpr2)
			p.indentLevel--

		} else {
			p.writeByte(space, 1)
			p.print(ofExpr2)
		}

	case ast.ObjectLocal:
		p.writeString("local ")
		p.fieldID(ofKind, ofExpr1, ofID)
		p.addMethodSignature(ofMethod)
		p.writeString(" = ")
		p.print(ofExpr2)
	case ast.ObjectFieldStr:
		p.fieldID(ofKind, ofExpr1, ofID)
		if ofSugar {
			p.writeByte(syntaxSugar, 1)
		}
		p.writeString(fieldType)
		p.writeByte(space, 1)
		p.print(ofExpr2)
	case ast.ObjectFieldExpr:
		p.writeString("[")
		p.fieldID(ofKind, ofExpr1, ofID)
		p.writeString("]: ")
		p.print(ofExpr2)
		if forSpec.VarName != "" {
			p.writeByte(newline, 1)
			p.forSpec(forSpec)
		}
	}
}

func isLocal(node ast.Node) bool {
	switch node.(type) {
	default:
		return false
	case *ast.Local:
		return true
	}
}

func (p *printer) addMethodSignature(fun *ast.Function) {
	if fun == nil {
		return
	}
	params := fun.Parameters

	p.writeString("(")
	var args []string
	for _, arg := range params.Required {
		args = append(args, string(arg))
	}

	for _, opt := range params.Optional {
		if opt.DefaultArg == nil {
			continue
		}
		var arg string
		arg += string(opt.Name)
		arg += "="

		child := printer{cfg: p.cfg}
		child.inFunction = true
		child.print(opt.DefaultArg)
		if child.err != nil {
			p.err = errors.Wrapf(child.err, "invalid argument for %s", string(opt.Name))
			return
		}

		arg += string(child.output)

		args = append(args, arg)
	}

	p.writeString(strings.Join(args, ", "))
	p.writeString(")")
}

var reDotIndex = regexp.MustCompile(`^\w[A-Za-z0-9]*$`)

func (p *printer) indexID(i *ast.Index) {
	if i == nil {
		p.err = errors.New("index is nil")
		return
	}

	if i.Index != nil {
		switch t := i.Index.(type) {
		default:
			p.err = errors.Errorf("can't handle index type %T", t)
			return
		case *ast.LiteralNumber:
			p.writeString(fmt.Sprintf(`[%s]`, t.OriginalString))
		case *ast.LiteralString:
			if t == nil {
				p.err = errors.New("string id is nil")
				return
			}

			id := t.Value
			if reDotIndex.MatchString(id) {
				p.writeString(fmt.Sprintf(".%s", id))
				return
			}
			quoted := stringQuote(id, true)
			p.writeString(fmt.Sprintf(`[%s]`, quoted))
		case *ast.Unary:
			p.writeString("[")
			p.print(t)
			p.writeString("]")
		case *ast.Var:
			p.writeString(fmt.Sprintf("[%s]", string(t.Id)))
		}
	} else if i.Id != nil {
		id := string(*i.Id)
		quoted := stringQuote(id, true)
		index := fmt.Sprintf(`[%s]`, quoted)
		if reDotIndex.MatchString(id) {
			index = fmt.Sprintf(`.%s`, id)
		}
		p.writeString(index)

	} else {
		p.err = errors.New("index and id can't both be blank")
		return
	}
}

func (p *printer) isObjectSingleLine(i interface{}) bool {
	if p.inFunction {
		return true
	}

	var loc *ast.LocationRange
	switch t := i.(type) {
	default:
		return false
	case *astext.Object:
		if len(t.Fields) == 0 {
			return true
		}
		if t.Oneline {
			return true
		}
		loc = t.NodeBase.Loc()
	case *ast.Object:
		if len(t.Fields) == 0 {
			return true
		}
		loc = t.NodeBase.Loc()
	}

	if loc == nil {
		return false
	}

	if loc.Begin.Line == 0 {
		return false
	}

	return loc.Begin.Line == loc.End.Line
}
