package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/diag"
	"github.com/unisc/compiladores/sol/src/parser"
)

// Expand resolves orbit imports in prog, inlining imported declarations in place.
func Expand(prog *ast.Program, entryFile string) (*ast.Program, []diag.Error) {
	if prog == nil {
		return prog, nil
	}
	loading := make(map[string]bool)
	var chain []string
	return expandProgram(prog, entryFile, loading, chain)
}

func expandProgram(prog *ast.Program, currentFile string, loading map[string]bool, chain []string) (*ast.Program, []diag.Error) {
	var errs []diag.Error
	var flat []ast.TopLevelDecl

	for _, decl := range prog.Decls {
		imp, ok := decl.(*ast.ImportDecl)
		if !ok {
			flat = append(flat, decl)
			continue
		}

		resolved, err := resolveImportPath(currentFile, imp.Path)
		if err != nil {
			errs = append(errs, diag.Error{
				File: currentFile, Line: imp.Pos().Line, Column: imp.Pos().Column,
				Message: err.Error(),
			})
			continue
		}

		abs, err := filepath.Abs(resolved)
		if err != nil {
			errs = append(errs, diag.Error{
				File: currentFile, Line: imp.Pos().Line, Column: imp.Pos().Column,
				Message: fmt.Sprintf("cannot resolve import path %q: %v", imp.Path, err),
			})
			continue
		}
		abs = filepath.Clean(abs)

		if loading[abs] {
			errs = append(errs, diag.Error{
				File: currentFile, Line: imp.Pos().Line, Column: imp.Pos().Column,
				Message: formatCircularImport(chain, abs),
			})
			continue
		}

		src, err := os.ReadFile(abs)
		if err != nil {
			errs = append(errs, diag.Error{
				File: currentFile, Line: imp.Pos().Line, Column: imp.Pos().Column,
				Message: fmt.Sprintf("cannot read import %q: %v", imp.Path, err),
			})
			continue
		}

		pp := parser.New(string(src), abs)
		sub := pp.Parse()
		errs = append(errs, pp.Errors()...)

		loading[abs] = true
		chain = append(chain, abs)
		expanded, subErrs := expandProgram(sub, abs, loading, chain)
		delete(loading, abs)
		chain = chain[:len(chain)-1]
		errs = append(errs, subErrs...)
		flat = append(flat, expanded.Decls...)
	}

	return &ast.Program{Decls: flat}, errs
}

func resolveImportPath(currentFile, importPath string) (string, error) {
	if filepath.IsAbs(importPath) {
		if _, err := os.Stat(importPath); err != nil {
			return "", fmt.Errorf("import file not found: %q", importPath)
		}
		return importPath, nil
	}
	base := filepath.Dir(currentFile)
	if base == "" || base == "." {
		base, _ = os.Getwd()
	}
	resolved := filepath.Clean(filepath.Join(base, importPath))
	if _, err := os.Stat(resolved); err != nil {
		return "", fmt.Errorf("import file not found: %q", importPath)
	}
	return resolved, nil
}

func formatCircularImport(chain []string, target string) string {
	msg := "circular import:"
	for _, p := range chain {
		msg += " " + filepath.Base(p) + " ->"
	}
	msg += " " + filepath.Base(target)
	return msg
}
