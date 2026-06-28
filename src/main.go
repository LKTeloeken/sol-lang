package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unisc/compiladores/sol/src/compiler"
)

func main() {
	lex := flag.Bool("lex", false, "run lexer only")
	parse := flag.Bool("parse", false, "run parser only")
	check := flag.Bool("check", false, "run semantic analysis")
	tac := flag.Bool("tac", false, "generate three-address code (TAC)")
	out := flag.String("o", "", "output file for -tac (default: stdout)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: sollang [options] <file> [args...]")
		os.Exit(1)
	}

	file := flag.Arg(0)
	phase := compiler.PhaseRun
	switch {
	case *lex:
		phase = compiler.PhaseLex
	case *parse:
		phase = compiler.PhaseParse
	case *check:
		phase = compiler.PhaseCheck
	case *tac:
		phase = compiler.PhaseCompile
	}

	scriptArgs := []string{}
	if flag.NArg() > 1 {
		scriptArgs = flag.Args()[1:]
	}

	res, err := compiler.RunFileWithOptions(file, phase, compiler.RunOptions{ScriptArgs: scriptArgs})
	if err != nil {
		fmt.Fprintf(os.Stderr, "sollang: %v\n", err)
		os.Exit(1)
	}

	if len(res.Errors) > 0 {
		compiler.PrintErrors(res.Errors)
		os.Exit(1)
	}

	switch phase {
	case compiler.PhaseLex:
		fmt.Print(compiler.FormatTokens(res.Tokens))

	case compiler.PhaseParse:
		if res.Program != nil {
			fmt.Printf("Program with %d top-level decls\n", len(res.Program.Decls))
		}

	case compiler.PhaseCheck:
		fmt.Println("semantic analysis OK")

	case compiler.PhaseCompile:
		if *out != "" {
			if err := os.WriteFile(*out, []byte(res.TAC), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "sollang: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Print(res.TAC)
		}

	case compiler.PhaseRun:
		if res.RunErr != nil {
			fmt.Fprintf(os.Stderr, "sollang: %v\n", res.RunErr)
			os.Exit(1)
		}
	}
}
