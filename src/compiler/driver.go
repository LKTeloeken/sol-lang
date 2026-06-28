package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/diag"
	"github.com/unisc/compiladores/sol/src/lexer"
	"github.com/unisc/compiladores/sol/src/modules"
	"github.com/unisc/compiladores/sol/src/parser"
	"github.com/unisc/compiladores/sol/src/semantic"
	"github.com/unisc/compiladores/sol/src/tac"
	"github.com/unisc/compiladores/sol/src/token"
	"github.com/unisc/compiladores/sol/src/vm"
)

type Phase int

const (
	PhaseLex Phase = iota
	PhaseParse
	PhaseCheck
	PhaseCompile
	PhaseRun
)

// RunOptions configures VM execution.
type RunOptions struct {
	ScriptArgs []string
}

type Result struct {
	Tokens   []token.Token
	Program  *ast.Program
	Errors   []diag.Error
	Analyzer *semantic.Analyzer
	TAC      string
	VM       *vm.VM
	RunErr   error
}

func RunFile(file string, phase Phase) (*Result, error) {
	return RunFileWithOptions(file, phase, RunOptions{})
}

func RunFileWithOptions(file string, phase Phase, opts RunOptions) (*Result, error) {
	src, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return RunWithOptions(string(src), file, phase, opts)
}

func Run(src, file string, phase Phase) (*Result, error) {
	return RunWithOptions(src, file, phase, RunOptions{})
}

func RunWithOptions(src, file string, phase Phase, opts RunOptions) (*Result, error) {
	res := &Result{}
	if phase == PhaseLex {
		res.Tokens = lexer.Tokenize(src)
		return res, nil
	}
	p := parser.New(src, file)
	prog := p.Parse()
	res.Errors = append(res.Errors, p.Errors()...)
	res.Program = prog
	if phase == PhaseParse {
		return res, nil
	}
	if phase >= PhaseCheck {
		var expandErrs []diag.Error
		prog, expandErrs = modules.Expand(prog, file)
		res.Errors = append(res.Errors, expandErrs...)
		res.Program = prog
	}
	if len(res.Errors) > 0 && phase == PhaseCheck {
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
	gen.Build(prog)
	if phase == PhaseCompile {
		res.TAC = gen.Format()
		return res, nil
	}
	machine := vm.New(gen.Instructions(), sem.Classes())
	machine.SetScriptArgs(opts.ScriptArgs)
	res.VM = machine
	res.RunErr = machine.Run()
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
