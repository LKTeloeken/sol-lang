package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	solbuiltin "github.com/unisc/compiladores/sol/src/builtin"
	"github.com/unisc/compiladores/sol/src/compiler"
)

func main() {
	lex := flag.Bool("lex", false, "run lexer only")
	parse := flag.Bool("parse", false, "run parser only")
	check := flag.Bool("check", false, "run semantic analysis")
	run := flag.Bool("run", false, "compile and execute with TAC interpreter")
	compile := flag.Bool("compile", false, "emit TAC")
	emitIR := flag.Bool("emit-ir", false, "emit LLVM IR")
	build := flag.Bool("build", false, "compile to native binary via LLVM")
	out := flag.String("o", "", "output path")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: solc [options] <file> [args...]")
		os.Exit(1)
	}

	file := flag.Arg(0)
	phase := compiler.PhaseCompile
	switch {
	case *lex:
		phase = compiler.PhaseLex

	case *parse:
		phase = compiler.PhaseParse

	case *check:
		phase = compiler.PhaseCheck

	case *run:
		phase = compiler.PhaseRun

	case *emitIR, *build:
		phase = compiler.PhaseEmitIR

	case *compile:
		phase = compiler.PhaseCompile

	default:
		*compile = true
	}

	scriptArgs := []string{}
	if flag.NArg() > 1 {
		scriptArgs = flag.Args()[1:]
	}

	res, err := compiler.CompileFileWithOptions(file, phase, compiler.RunOptions{ScriptArgs: scriptArgs})
	if err != nil {
		fmt.Fprintf(os.Stderr, "solc: %v\n", err)
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
			fmt.Print(dumpProgram(res))
		}

	case compiler.PhaseCheck:
		fmt.Println("semantic analysis OK")

	case compiler.PhaseCompile:
		outPath := *out
		if outPath == "" {
			outPath = "output.tac"
		}

		if err := os.WriteFile(outPath, []byte(res.TAC), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "solc: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("wrote TAC to %s\n", outPath)

	case compiler.PhaseRun:
		if res.RunErr != nil {
			fmt.Fprintf(os.Stderr, "solc: %v\n", res.RunErr)
			os.Exit(1)
		}

	case compiler.PhaseEmitIR:
		outPath := *out
		if outPath == "" {
			outPath = "output.ll"
		}

		if err := os.WriteFile(outPath, []byte(res.IR), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "solc: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("wrote LLVM IR to %s\n", outPath)
		if *build {
			binPath := "program"
			if *out != "" {
				binPath = stringsTrimExt(*out, ".ll")
			}

			if err := buildBinary(outPath, binPath); err != nil {
				fmt.Fprintf(os.Stderr, "solc build: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("wrote binary to %s\n", binPath)
		}
	}
}

func dumpProgram(res *compiler.Result) string {
	if res.Program == nil {
		return ""
	}

	return fmt.Sprintf("Program with %d top-level decls\n", len(res.Program.Decls))
}

func stringsTrimExt(path, ext string) string {
	if filepath.Ext(path) == ext {
		return path[:len(path)-len(ext)]
	}

	return path
}

func buildBinary(llPath, binPath string) error {
	tmp, err := os.CreateTemp("", "solrt-*.c")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(solbuiltin.SolRT); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	cmd := exec.Command("clang", llPath, tmp.Name(), "-o", binPath, "-lm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
