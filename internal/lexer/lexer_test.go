package lexer

import (
	"testing"

	"github.com/unisc/compiladores/sol/internal/token"
)

func TestNextToken_Basic(t *testing.T) {
	input := `rise Foo { private int x; }`
	tests := []struct {
		typ    token.Type
		lexeme string
	}{
		{token.RISE, "rise"},
		{token.IDENT, "Foo"},
		{token.LBRACE, "{"},
		{token.PRIVATE, "private"},
		{token.INT_TYPE, "int"},
		{token.IDENT, "x"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.typ {
			t.Fatalf("tests[%d] - type wrong. expected=%q, got=%q", i, tt.typ, tok.Type)
		}
		if tt.lexeme != "" && tok.Lexeme != tt.lexeme {
			t.Fatalf("tests[%d] - lexeme wrong. expected=%q, got=%q", i, tt.lexeme, tok.Lexeme)
		}
	}
}

func TestNextToken_Operators(t *testing.T) {
	input := `== != <= >= && ||`
	l := New(input)
	want := []token.Type{token.EQ, token.NOT_EQ, token.LT_EQ, token.GT_EQ, token.AND, token.OR, token.EOF}
	for i, wt := range want {
		tok := l.NextToken()
		if tok.Type != wt {
			t.Fatalf("tests[%d] expected %q got %q", i, wt, tok.Type)
		}
	}
}

func TestNextToken_RangeOperator(t *testing.T) {
	input := `0..10 3.14 .field`
	l := New(input)
	want := []token.Type{token.INT, token.DOTDOT, token.INT, token.FLOAT, token.DOT, token.IDENT, token.EOF}
	for i, wt := range want {
		tok := l.NextToken()
		if tok.Type != wt {
			t.Fatalf("tests[%d] expected %q got %q", i, wt, tok.Type)
		}
	}
}

func TestNextToken_Comments(t *testing.T) {
	input := `// comment
var x int = 1; /* block */ var y int = 2;`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.VAR {
		t.Fatalf("expected VAR after line comment, got %v", tok.Type)
	}
	for {
		tok = l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.VAR && tok.Lexeme == "var" {
			// second var after block comment
			return
		}
	}
	t.Fatal("expected second var declaration")
}

func TestNextToken_OrbitImport(t *testing.T) {
	input := `orbit "utils.sol";`
	l := New(input)
	want := []struct {
		typ    token.Type
		lexeme string
	}{
		{token.ORBIT, "orbit"},
		{token.STRING, "utils.sol"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	for i, tt := range want {
		tok := l.NextToken()
		if tok.Type != tt.typ {
			t.Fatalf("tests[%d] type: want %q got %q", i, tt.typ, tok.Type)
		}
		if tt.lexeme != "" && tok.Lexeme != tt.lexeme {
			t.Fatalf("tests[%d] lexeme: want %q got %q", i, tt.lexeme, tok.Lexeme)
		}
	}
}

func TestNextToken_StringAndNumbers(t *testing.T) {
	input := `"hello" 42 3.14 true null`
	l := New(input)
	tests := []token.Type{token.STRING, token.INT, token.FLOAT, token.BOOL, token.NULL, token.EOF}
	for i, wt := range tests {
		tok := l.NextToken()
		if tok.Type != wt {
			t.Fatalf("tests[%d] expected %q got %q", i, wt, tok.Type)
		}
	}
}
