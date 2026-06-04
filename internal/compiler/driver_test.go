package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileLex(t *testing.T) {
	res, err := CompileFile("../../examples/simple.sol", PhaseLex)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tokens) == 0 {
		t.Fatal("expected tokens")
	}
}

func TestCompileParse(t *testing.T) {
	res, err := CompileFile("../../examples/simple.sol", PhaseParse)
	if err != nil {
		t.Fatal(err)
	}
	if res.Program == nil || len(res.Program.Decls) == 0 {
		t.Fatal("expected program decls")
	}
}

func TestCompileCheck(t *testing.T) {
	res, err := CompileFile("../../examples/conta_bancaria.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
}

func TestCompileTAC(t *testing.T) {
	res, err := CompileFile("../../examples/conta_bancaria.sol", PhaseCompile)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if !strings.Contains(res.TAC, "ContaBancaria") {
		t.Fatalf("expected ContaBancaria in TAC")
	}
}

func TestCompileInvalid(t *testing.T) {
	res, err := CompileFile("../../testdata/invalid/type_mismatch.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected semantic errors")
	}
}

func TestCompileOrbitModules(t *testing.T) {
	res, err := CompileFile("../../examples/modules/main.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
}

func TestCompileOrbitMissing(t *testing.T) {
	res, err := CompileFile("../../testdata/invalid/orbit_missing.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected import errors")
	}
}

func TestCompileOrbitCircular(t *testing.T) {
	res, err := CompileFile("../../testdata/invalid/orbit_cycle_a.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected circular import error")
	}
}

func TestCheckReservedStdlibClass(t *testing.T) {
	res, err := CompileFile("../../testdata/invalid/reserved_class_time.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected reserved class name error")
	}
}

func TestCheckBuiltinWrongArgs(t *testing.T) {
	res, err := CompileFile("../../testdata/invalid/builtin_wrong_args.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected builtin arity error")
	}
}

func TestEmitIR(t *testing.T) {
	res, err := CompileFile("../../examples/simple.sol", PhaseEmitIR)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if !strings.Contains(res.IR, "define i32 @main") {
		t.Fatalf("expected main in IR")
	}
}

func TestRuntimePath(t *testing.T) {
	wd, _ := os.Getwd()
	rt := filepath.Join(wd, "..", "..", "runtime", "solrt.c")
	if _, err := os.Stat(rt); err != nil {
		t.Fatalf("runtime not found at %s: %v", rt, err)
	}
}
