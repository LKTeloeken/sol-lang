package parser

import (
	"testing"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/lexer"
)

func TestParseClass(t *testing.T) {
	src := `rise Foo { private int x; glow(int x) { this.x = x; } }`
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

func TestParseOrbitImport(t *testing.T) {
	src := `orbit "utils.sol";`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	imp, ok := prog.Decls[0].(*ast.ImportDecl)
	if !ok {
		t.Fatalf("expected ImportDecl, got %T", prog.Decls[0])
	}
	if imp.Path != "utils.sol" {
		t.Fatalf("path = %q, want utils.sol", imp.Path)
	}
}

func TestParseOrbitImportMissingString(t *testing.T) {
	src := `orbit foo;`
	p := New(lexer.New(src), "test.sol")
	_ = p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected parse error for non-string after orbit")
	}
}

func TestParseInheritance(t *testing.T) {
	src := `rise Child radiate Parent { private int x; }`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if prog == nil || len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
}

func TestParsePostfixArrayType(t *testing.T) {
	src := `var nums int[] = [1, 2]; var names string[][];`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 2 {
		t.Fatalf("expected 2 decls, got %d", len(prog.Decls))
	}
}

func TestParseStarAlias(t *testing.T) {
	src := `star TodoItems = string[];`
	p := New(lexer.New(src), "test.sol")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	alias, ok := prog.Decls[0].(*ast.TypeAliasDecl)
	if !ok {
		t.Fatalf("expected TypeAliasDecl, got %T", prog.Decls[0])
	}
	if alias.Name != "TodoItems" || !alias.Type.IsArray || alias.Type.ElemType.Base != "string" {
		t.Fatalf("unexpected alias: %+v", alias)
	}
}
