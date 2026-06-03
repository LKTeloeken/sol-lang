package parser

import (
	"testing"

	"github.com/unisc/compiladores/sol/internal/lexer"
)

func TestParseClass(t *testing.T) {
	src := `shine Foo { private int x; glow(int x) { this.x = x; } }`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
}

func TestParseTopLevel(t *testing.T) {
	src := `var x int = 1;`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
}

func TestParseInheritance(t *testing.T) {
	src := `shine Child eclipse Parent { private int x; }`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if prog == nil || len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
}
