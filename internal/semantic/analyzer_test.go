package semantic

import (
	"testing"

	"github.com/unisc/compiladores/sol/internal/lexer"
	"github.com/unisc/compiladores/sol/internal/parser"
)

func TestCheckValidProgram(t *testing.T) {
	src := `shine Foo { private int x; glow(int x) { this.x = x; } public ray getX() { emit this.x; } }
var f Foo = new Foo(1);`
	p := parser.New(lexer.New(src), "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", a.Errors())
	}
}

func TestCheckUnknownClass(t *testing.T) {
	src := `var x Foo = new Foo(1);`
	p := parser.New(lexer.New(src), "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) == 0 {
		t.Fatal("expected semantic errors")
	}
}

func TestCheckTypeMismatch(t *testing.T) {
	src := `var x int = "hello";`
	p := parser.New(lexer.New(src), "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) == 0 {
		t.Fatal("expected type mismatch error")
	}
}
