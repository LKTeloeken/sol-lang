package compiler

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHello(t *testing.T) {
	res, err := CompileFile("../../examples/hello.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
}

func TestRunForEach(t *testing.T) {
	res, err := CompileFile("../../examples/for_each.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
}

func TestRunControlFlow(t *testing.T) {
	res, err := CompileFile("../../examples/control_flow.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
}

func TestRunScript(t *testing.T) {
	res, err := CompileFile("../../examples/script.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
}

func TestRunContaBancaria(t *testing.T) {
	res, err := CompileFile("../../examples/conta_bancaria.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
	saldo, err := res.VM.GetField("conta", "saldo")
	if err != nil {
		t.Fatal(err)
	}
	if saldo.AsFloat() != 1150 {
		t.Fatalf("expected saldo 1150, got %v", saldo.AsFloat())
	}
}

func TestRunSimpleGetX(t *testing.T) {
	res, err := CompileFile("../../examples/simple.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
	y, ok := res.VM.Global("y")
	if !ok {
		t.Fatal("expected global y")
	}
	if y.AsFloat() != 42 {
		t.Fatalf("expected y=42, got %v", y.AsFloat())
	}
}

func TestRunFlareUncaught(t *testing.T) {
	src := `
shine X {
	public ray boom() { flare "boom"; }
}
var x X = new X();
x.boom();
`
	res, err := Compile(src, "test.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("compile errors: %v", res.Errors)
	}
	if res.RunErr == nil {
		t.Fatal("expected flare error")
	}
}

func TestRunHeranca(t *testing.T) {
	res, err := CompileFile("../../examples/heranca.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
	saldo, err := res.VM.GetField("especial", "saldo")
	if err != nil {
		t.Fatal(err)
	}
	if saldo.AsFloat() != 450 {
		t.Fatalf("expected saldo 450, got %v", saldo.AsFloat())
	}
}

func TestStringConcat(t *testing.T) {
	src := `var s string = "n=" + 42;`
	res, err := Compile(src, "test.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
	s, ok := res.VM.Global("s")
	if !ok {
		t.Fatal("expected global s")
	}
	if s.StrVal != "n=42" {
		t.Fatalf("expected n=42, got %q", s.StrVal)
	}
}

func TestBuildHello(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not available")
	}

	res, err := CompileFile("../../examples/hello.sol", PhaseEmitIR)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}

	dir := t.TempDir()
	llPath := filepath.Join(dir, "hello.ll")
	binPath := filepath.Join(dir, "hello")
	if err := os.WriteFile(llPath, []byte(res.IR), 0644); err != nil {
		t.Fatal(err)
	}

	rtPath, err := filepath.Abs("../../runtime/solrt.c")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("clang", llPath, rtPath, "-o", binPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clang failed: %v\n%s", err, out)
	}

	run := exec.Command(binPath)
	var stdout bytes.Buffer
	run.Stdout = &stdout
	if err := run.Run(); err != nil {
		t.Fatalf("binary failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Hello, SOL") {
		t.Fatalf("expected Hello, SOL on stdout, got %q", stdout.String())
	}
}

func TestConsolePrintOutput(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not available")
	}
	_ = io.Discard
	src := `Console.print("Hello, SOL");`
	res, err := Compile(src, "test.sol", PhaseRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
}
