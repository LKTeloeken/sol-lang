package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/unisc/compiladores/sol/internal/ast"
	"github.com/unisc/compiladores/sol/internal/diag"
	"github.com/unisc/compiladores/sol/internal/lexer"
	"github.com/unisc/compiladores/sol/internal/parser"
	"github.com/unisc/compiladores/sol/internal/semantic"
	"github.com/unisc/compiladores/sol/internal/tac"
	"github.com/unisc/compiladores/sol/internal/token"
	"github.com/unisc/compiladores/sol/internal/vm"
)

type Phase int

const (
	PhaseLex Phase = iota
	PhaseParse
	PhaseCheck
	PhaseCompile
	PhaseRun
	PhaseEmitIR
	PhaseBuild
)

type Result struct {
	Tokens   []token.Token
	Program  *ast.Program
	TAC      string
	IR       string
	Errors   []diag.Error
	Analyzer *semantic.Analyzer
	VM       *vm.VM
	RunErr   error
}

func CompileFile(file string, phase Phase) (*Result, error) {
	src, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return Compile(string(src), file, phase)
}

func Compile(src, file string, phase Phase) (*Result, error) {
	res := &Result{}
	l := lexer.New(src)
	if phase == PhaseLex {
		res.Tokens = lexer.Tokenize(src)
		return res, nil
	}
	p := parser.New(l, file)
	prog := p.Parse()
	res.Errors = append(res.Errors, p.Errors()...)
	res.Program = prog
	if phase == PhaseParse {
		return res, nil
	}
	sem := semantic.New(file)
	sem.Check(prog)
	res.Errors = append(res.Errors, sem.Errors()...)
	res.Analyzer = sem
	if phase == PhaseCheck {
		return res, nil
	}
	if len(res.Errors) > 0 {
		return res, nil
	}
	gen := tac.New(sem.Classes())
	res.TAC = gen.Generate(prog)
	if phase == PhaseCompile {
		return res, nil
	}
	if phase == PhaseRun {
		machine := vm.New(gen.Instructions(), sem.Classes())
		res.VM = machine
		res.RunErr = machine.Run()
		return res, nil
	}
	irGen := NewIRGen(sem.Classes())
	res.IR = irGen.Generate(prog)
	return res, nil
}

func FormatTokens(tokens []token.Token) string {
	var b strings.Builder
	for _, t := range tokens {
		if t.Type == token.EOF {
			b.WriteString("EOF\n")
			break
		}
		if t.Type == token.ILLEGAL {
			fmt.Fprintf(&b, "ILLEGAL(%q) [%d:%d]\n", t.Lexeme, t.Line, t.Column)
			continue
		}
		fmt.Fprintf(&b, "%s(%q) [%d:%d]\n", t.Type, t.Lexeme, t.Line, t.Column)
	}
	return b.String()
}

func PrintErrors(errors []diag.Error) {
	for _, e := range errors {
		fmt.Fprintln(os.Stderr, e.Error())
	}
}
