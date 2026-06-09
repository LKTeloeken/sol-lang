package compiler

import (
	"testing"
)

func TestCompileLex(t *testing.T) {
	res, err := RunFile("../../examples/simple.sol", PhaseLex)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tokens) == 0 {
		t.Fatal("expected tokens")
	}
}

func TestCompileParse(t *testing.T) {
	res, err := RunFile("../../examples/simple.sol", PhaseParse)
	if err != nil {
		t.Fatal(err)
	}
	if res.Program == nil || len(res.Program.Decls) == 0 {
		t.Fatal("expected program decls")
	}
}

func TestCompileCheck(t *testing.T) {
	res, err := RunFile("../../examples/conta_bancaria.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
}

func TestCompileInvalid(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/type_mismatch.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected semantic errors")
	}
}

func TestCompileOrbitModules(t *testing.T) {
	res, err := RunFile("../../examples/modules/main.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
}

func TestCompileOrbitMissing(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/orbit_missing.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected import errors")
	}
}

func TestCompileOrbitCircular(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/orbit_cycle_a.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected circular import error")
	}
}

func TestCheckReservedStdlibClass(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/reserved_class_time.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected reserved class name error")
	}
}

func TestCheckBuiltinWrongArgs(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/builtin_wrong_args.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected builtin arity error")
	}
}

func TestCheckArrayPrefixSyntax(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/array_prefix_syntax.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected parse errors for prefix array syntax")
	}
}

func TestCheckTypeAliasCycle(t *testing.T) {
	res, err := RunFile("../../testdata/invalid/type_alias_cycle.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected circular type alias error")
	}
}

func TestCompileRealTest(t *testing.T) {
	res, err := RunFile("../../examples/real-test/main.sol", PhaseCheck)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
}
