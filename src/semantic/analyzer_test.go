package semantic

import (
	"testing"

	"github.com/unisc/compiladores/sol/src/parser"
)

func TestCheckValidProgram(t *testing.T) {
	src := `rise Foo { private int x; glow(int x) { this.x = x; } public ray getX() { emit this.x; } }
var f Foo = new Foo(1);`
	p := parser.New(src, "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", a.Errors())
	}
}

func TestCheckUnknownClass(t *testing.T) {
	src := `var x Foo = new Foo(1);`
	p := parser.New(src, "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) == 0 {
		t.Fatal("expected semantic errors")
	}
}

func TestCheckTypeMismatch(t *testing.T) {
	src := `var x int = "hello";`
	p := parser.New(src, "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) == 0 {
		t.Fatal("expected type mismatch error")
	}
}

func TestCheckStarAliasAndArrayMethods(t *testing.T) {
	src := `
star TodoItems = string[];

rise TodoList {
    private TodoItems items;

    glow() {
        this.items = [];
    }

    public ray add(string item) {
        this.items.push(item);
    }

    public ray size() int {
        emit this.items.length;
    }
}

var t TodoList = new TodoList();
t.add("a");
`
	p := parser.New(src, "test.sol")
	prog := p.Parse()
	a := New("test.sol")
	a.Check(prog)
	if len(a.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", a.Errors())
	}
}

func TestCheckArrayPrefixRejected(t *testing.T) {
	src := `var x [int] = [1];`
	p := parser.New(src, "test.sol")
	prog := p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected parse errors for prefix array syntax")
	}
	_ = prog
}
