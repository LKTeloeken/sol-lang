package lexer

import (
	"strings"

	plexer "github.com/alecthomas/participle/v2/lexer"
	"github.com/unisc/compiladores/sol/internal/token"
)

var def = plexer.MustStateful(plexer.Rules{
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
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Punct", Pattern: `[;,.()\[\]{}+\-*/%!=<>]`},
		{Name: "Illegal", Pattern: `.`},
	},
})

var sym = def.Symbols()

type Lexer struct {
	inner plexer.Lexer
}

func New(input string) *Lexer {
	inner, err := def.Lex("", strings.NewReader(input))
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

	case sym["Ident"]:
		typ := token.LookupIdent(val)

		if typ == token.TRUE || typ == token.FALSE {
			return token.Token{Type: token.BOOL, Lexeme: val, Line: line, Column: col}
		}

		return token.Token{Type: typ, Lexeme: val, Line: line, Column: col}
	case sym["Punct"]:
		return mapPunct(val, line, col)

	default: // Illegal
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
