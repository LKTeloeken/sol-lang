package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/lexer"
	"github.com/unisc/compiladores/sol/src/parser"
	"github.com/unisc/compiladores/sol/src/semantic"
)

func parseFile(t *testing.T, path, src string) *ast.Program {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	p := parser.New(lexer.New(src), path)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return prog
}

func TestExpandImportsClass(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.sol")
	main := filepath.Join(dir, "main.sol")

	parseFile(t, lib, `rise Greeter {
    glow() { }
    public ray ping() { }
}`)
	mainSrc := `orbit "lib.sol";
var g Greeter = new Greeter();`
	parseFile(t, main, mainSrc)

	prog := parser.New(lexer.New(mainSrc), main).Parse()
	expanded, errs := Expand(prog, main)
	if len(errs) > 0 {
		t.Fatalf("expand errors: %v", errs)
	}
	hasClass := false
	for _, d := range expanded.Decls {
		if c, ok := d.(*ast.ClassDecl); ok && c.Name == "Greeter" {
			hasClass = true
		}
	}
	if !hasClass {
		t.Fatal("expected Greeter class after expand")
	}

	sem := semantic.New(main)
	sem.Check(expanded)
	if len(sem.Errors()) > 0 {
		t.Fatalf("semantic errors: %v", sem.Errors())
	}
}

func TestExpandCircularImport(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.sol")
	b := filepath.Join(dir, "b.sol")

	parseFile(t, a, `orbit "b.sol";`)
	parseFile(t, b, `orbit "a.sol";`)

	prog := parser.New(lexer.New(`orbit "b.sol";`), a).Parse()
	_, errs := Expand(prog, a)
	if len(errs) == 0 {
		t.Fatal("expected circular import error")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "circular import") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected circular import message, got %v", errs)
	}
}

func TestExpandMissingFile(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "main.sol")
	mainSrc := `orbit "missing.sol";`
	parseFile(t, main, mainSrc)

	prog := parser.New(lexer.New(mainSrc), main).Parse()
	_, errs := Expand(prog, main)
	if len(errs) == 0 {
		t.Fatal("expected missing file error")
	}
}

func TestExpandImportOrder(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.sol")
	main := filepath.Join(dir, "main.sol")

	parseFile(t, lib, `Console.print("from-lib");`)
	mainSrc := `orbit "lib.sol";
Console.print("from-main");`
	parseFile(t, main, mainSrc)

	prog := parser.New(lexer.New(mainSrc), main).Parse()
	expanded, errs := Expand(prog, main)
	if len(errs) > 0 {
		t.Fatalf("expand errors: %v", errs)
	}
	if len(expanded.Decls) != 2 {
		t.Fatalf("expected 2 top-level stmts, got %d", len(expanded.Decls))
	}
}
