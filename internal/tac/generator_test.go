package tac

import (
	"strings"
	"testing"

	"github.com/unisc/compiladores/sol/internal/lexer"
	"github.com/unisc/compiladores/sol/internal/parser"
	"github.com/unisc/compiladores/sol/internal/semantic"
)

func TestGenerateTAC(t *testing.T) {
	src := `rise Foo {
    private int x;
    glow(int x) { this.x = x; }
    public ray getX() { emit this.x; }
}
var f Foo = new Foo(10);`
	p := parser.New(lexer.New(src), "test.sol")
	prog := p.Parse()
	a := semantic.New("test.sol")
	a.Check(prog)
	if len(a.Errors()) > 0 {
		t.Fatalf("semantic errors: %v", a.Errors())
	}
	gen := New(a.Classes())
	out := gen.Generate(prog)
	if !strings.Contains(out, "Foo.glow") {
		t.Fatalf("expected Foo.glow in TAC, got:\n%s", out)
	}
	if !strings.Contains(out, "new Foo") {
		t.Fatalf("expected new Foo in TAC, got:\n%s", out)
	}
}

func TestGenerateCallTAC(t *testing.T) {
	src := `rise Foo {
    public ray bar() { }
}
var f Foo = new Foo();
f.bar();`
	p := parser.New(lexer.New(src), "test.sol")
	prog := p.Parse()
	a := semantic.New("test.sol")
	a.Check(prog)
	gen := New(a.Classes())
	out := gen.Generate(prog)
	if !strings.Contains(out, "call Foo.bar") {
		t.Fatalf("expected call Foo.bar in TAC, got:\n%s", out)
	}
}
