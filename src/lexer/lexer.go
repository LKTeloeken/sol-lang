package lexer

import (
	"strings"

	plexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/unisc/compiladores/sol/src/token"
)

// Def is the exported stateful lexer definition, used by both the old token
// stream interface and the participle-based parser.
var Def = plexer.MustStateful(plexer.Rules{
	"Root": {
		{Name: "Comment", Pattern: `//[^\n]*|(?s:/\*.*?\*/)`},
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "Float", Pattern: `\d+\.\d+`},
		{Name: "Int", Pattern: `\d+`},
		{Name: "String", Pattern: `"(?:[^"\\]|\\.)*"`},
		{Name: "DotDot", Pattern: `\.\.`},
		{Name: "Eq", Pattern: `==`},
		{Name: "NotEq", Pattern: `!=`},
		{Name: "LtEq", Pattern: `<=`},
		{Name: "GtEq", Pattern: `>=`},
		{Name: "And", Pattern: `&&`},
		{Name: "Or", Pattern: `\|\|`},
		// Keywords must appear before Ident so they take priority.
		{Name: "Rise", Pattern: `rise\b`},
		{Name: "Ray", Pattern: `ray\b`},
		{Name: "Glow", Pattern: `glow\b`},
		{Name: "Radiate", Pattern: `radiate\b`},
		{Name: "Emit", Pattern: `emit\b`},
		{Name: "Flare", Pattern: `flare\b`},
		{Name: "Orbit", Pattern: `orbit\b`},
		{Name: "Star", Pattern: `star\b`},
		{Name: "Public", Pattern: `public\b`},
		{Name: "Private", Pattern: `private\b`},
		{Name: "Var", Pattern: `var\b`},
		{Name: "New", Pattern: `new\b`},
		{Name: "This", Pattern: `this\b`},
		{Name: "If", Pattern: `if\b`},
		{Name: "Else", Pattern: `else\b`},
		{Name: "While", Pattern: `while\b`},
		{Name: "For", Pattern: `for\b`},
		{Name: "Each", Pattern: `each\b`},
		{Name: "In", Pattern: `in\b`},
		{Name: "Break", Pattern: `break\b`},
		{Name: "Continue", Pattern: `continue\b`},
		{Name: "Try", Pattern: `try\b`},
		{Name: "Catch", Pattern: `catch\b`},
		{Name: "True", Pattern: `true\b`},
		{Name: "False", Pattern: `false\b`},
		{Name: "Null", Pattern: `null\b`},
		{Name: "IntType", Pattern: `int\b`},
		{Name: "FloatType", Pattern: `float\b`},
		{Name: "BoolType", Pattern: `bool\b`},
		{Name: "StringType", Pattern: `string\b`},
		{Name: "VoidType", Pattern: `void\b`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Punct", Pattern: `[;,.()\[\]{}+\-*/%!=<>]`},
		{Name: "Illegal", Pattern: `.`},
	},
})

var sym = Def.Symbols()

type Lexer struct {
	inner plexer.Lexer
}

func New(input string) *Lexer {
	inner, err := Def.Lex("", strings.NewReader(input))
	if err != nil {
		panic(err)
	}
	return &Lexer{inner: inner}
}

func (l *Lexer) NextToken() token.Token {
	for {
		pt, err := l.inner.Next()
		if err != nil || pt.EOF() {
			return token.Token{Type: token.EOF}
		}
		t := pt.Type
		if t == sym["Whitespace"] || t == sym["Comment"] {
			continue
		}
		return mapToken(pt)
	}
}

func mapToken(pt plexer.Token) token.Token {
	line := pt.Pos.Line
	col := pt.Pos.Column
	val := pt.Value
	t := pt.Type

	switch t {
	case sym["Float"]:
		return token.Token{Type: token.FLOAT, Lexeme: val, Line: line, Column: col}
	case sym["Int"]:
		return token.Token{Type: token.INT, Lexeme: val, Line: line, Column: col}
	case sym["String"]:
		return token.Token{Type: token.STRING, Lexeme: val[1 : len(val)-1], Line: line, Column: col}
	case sym["DotDot"]:
		return token.Token{Type: token.DOTDOT, Lexeme: val, Line: line, Column: col}
	case sym["Eq"]:
		return token.Token{Type: token.EQ, Lexeme: val, Line: line, Column: col}
	case sym["NotEq"]:
		return token.Token{Type: token.NOT_EQ, Lexeme: val, Line: line, Column: col}
	case sym["LtEq"]:
		return token.Token{Type: token.LT_EQ, Lexeme: val, Line: line, Column: col}
	case sym["GtEq"]:
		return token.Token{Type: token.GT_EQ, Lexeme: val, Line: line, Column: col}
	case sym["And"]:
		return token.Token{Type: token.AND, Lexeme: val, Line: line, Column: col}
	case sym["Or"]:
		return token.Token{Type: token.OR, Lexeme: val, Line: line, Column: col}
	case sym["Rise"]:
		return token.Token{Type: token.RISE, Lexeme: val, Line: line, Column: col}
	case sym["Ray"]:
		return token.Token{Type: token.RAY, Lexeme: val, Line: line, Column: col}
	case sym["Glow"]:
		return token.Token{Type: token.GLOW, Lexeme: val, Line: line, Column: col}
	case sym["Radiate"]:
		return token.Token{Type: token.RADIATE, Lexeme: val, Line: line, Column: col}
	case sym["Emit"]:
		return token.Token{Type: token.EMIT, Lexeme: val, Line: line, Column: col}
	case sym["Flare"]:
		return token.Token{Type: token.FLARE, Lexeme: val, Line: line, Column: col}
	case sym["Orbit"]:
		return token.Token{Type: token.ORBIT, Lexeme: val, Line: line, Column: col}
	case sym["Star"]:
		return token.Token{Type: token.STAR, Lexeme: val, Line: line, Column: col}
	case sym["Public"]:
		return token.Token{Type: token.PUBLIC, Lexeme: val, Line: line, Column: col}
	case sym["Private"]:
		return token.Token{Type: token.PRIVATE, Lexeme: val, Line: line, Column: col}
	case sym["Var"]:
		return token.Token{Type: token.VAR, Lexeme: val, Line: line, Column: col}
	case sym["New"]:
		return token.Token{Type: token.NEW, Lexeme: val, Line: line, Column: col}
	case sym["This"]:
		return token.Token{Type: token.THIS, Lexeme: val, Line: line, Column: col}
	case sym["If"]:
		return token.Token{Type: token.IF, Lexeme: val, Line: line, Column: col}
	case sym["Else"]:
		return token.Token{Type: token.ELSE, Lexeme: val, Line: line, Column: col}
	case sym["While"]:
		return token.Token{Type: token.WHILE, Lexeme: val, Line: line, Column: col}
	case sym["For"]:
		return token.Token{Type: token.FOR, Lexeme: val, Line: line, Column: col}
	case sym["Each"]:
		return token.Token{Type: token.EACH, Lexeme: val, Line: line, Column: col}
	case sym["In"]:
		return token.Token{Type: token.IN, Lexeme: val, Line: line, Column: col}
	case sym["Break"]:
		return token.Token{Type: token.BREAK, Lexeme: val, Line: line, Column: col}
	case sym["Continue"]:
		return token.Token{Type: token.CONTINUE, Lexeme: val, Line: line, Column: col}
	case sym["Try"]:
		return token.Token{Type: token.TRY, Lexeme: val, Line: line, Column: col}
	case sym["Catch"]:
		return token.Token{Type: token.CATCH, Lexeme: val, Line: line, Column: col}
	case sym["True"]:
		return token.Token{Type: token.BOOL, Lexeme: val, Line: line, Column: col}
	case sym["False"]:
		return token.Token{Type: token.BOOL, Lexeme: val, Line: line, Column: col}
	case sym["Null"]:
		return token.Token{Type: token.NULL, Lexeme: val, Line: line, Column: col}
	case sym["IntType"]:
		return token.Token{Type: token.INT_TYPE, Lexeme: val, Line: line, Column: col}
	case sym["FloatType"]:
		return token.Token{Type: token.FLOAT_TYPE, Lexeme: val, Line: line, Column: col}
	case sym["BoolType"]:
		return token.Token{Type: token.BOOL_TYPE, Lexeme: val, Line: line, Column: col}
	case sym["StringType"]:
		return token.Token{Type: token.STRING_TYPE, Lexeme: val, Line: line, Column: col}
	case sym["VoidType"]:
		return token.Token{Type: token.VOID_TYPE, Lexeme: val, Line: line, Column: col}
	case sym["Ident"]:
		return token.Token{Type: token.IDENT, Lexeme: val, Line: line, Column: col}
	case sym["Punct"]:
		return mapPunct(val, line, col)
	default:
		return token.Token{Type: token.ILLEGAL, Lexeme: val, Line: line, Column: col}
	}
}

func mapPunct(val string, line, col int) token.Token {
	var typ token.Type
	switch val {
	case ";":
		typ = token.SEMICOLON
	case ",":
		typ = token.COMMA
	case ".":
		typ = token.DOT
	case "(":
		typ = token.LPAREN
	case ")":
		typ = token.RPAREN
	case "{":
		typ = token.LBRACE
	case "}":
		typ = token.RBRACE
	case "[":
		typ = token.LBRACK
	case "]":
		typ = token.RBRACK
	case "+":
		typ = token.PLUS
	case "-":
		typ = token.MINUS
	case "*":
		typ = token.ASTERISK
	case "/":
		typ = token.SLASH
	case "%":
		typ = token.PERCENT
	case "!":
		typ = token.BANG
	case "=":
		typ = token.ASSIGN
	case "<":
		typ = token.LT
	case ">":
		typ = token.GT
	default:
		typ = token.ILLEGAL
	}
	return token.Token{Type: typ, Lexeme: val, Line: line, Column: col}
}

// Tokenize reads all tokens until EOF.
func Tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
