package compiler

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unisc/compiladores/sol/src/lexer"
	"github.com/unisc/compiladores/sol/src/parser"
	"github.com/unisc/compiladores/sol/src/semantic"
	"github.com/unisc/compiladores/sol/src/tac"
	"github.com/unisc/compiladores/sol/src/vm"
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

func TestRunForRange(t *testing.T) {
	res, err := CompileFile("../../examples/for_range.sol", PhaseRun)
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

func TestBreakContinue(t *testing.T) {
	src := `
var sum int = 0;
for i in 0..10 {
    sum = sum + i;
}
`
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
	sum, ok := res.VM.Global("sum")
	if !ok {
		t.Fatal("expected global sum")
	}
	if sum.AsFloat() != 45 {
		t.Fatalf("expected sum=45, got %v", sum.AsFloat())
	}
}

func TestBreakContinueAdvanced(t *testing.T) {
	src := `
var sum int = 0;
for i in 0..100 {
    if (i == 50) {
        break;
    }
    if (i % 2 == 0) {
        continue;
    }
    sum = sum + i;
}
`
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
	sum, ok := res.VM.Global("sum")
	if !ok {
		t.Fatal("expected global sum")
	}
	if sum.AsFloat() != 625 {
		t.Fatalf("expected sum=625, got %v", sum.AsFloat())
	}
}

func TestConsoleReadLine(t *testing.T) {
	machine := runSourceWithStdin(t, `var s string = Console.readLine("> ");`, "Lucas\n")
	s, ok := machine.Global("s")
	if !ok {
		t.Fatal("expected global s")
	}
	if s.StrVal != "Lucas" {
		t.Fatalf("expected Lucas, got %q", s.StrVal)
	}
}

func TestConsoleReadInt(t *testing.T) {
	machine := runSourceWithStdin(t, `var n int = Console.readInt();`, "42\n")
	n, ok := machine.Global("n")
	if !ok {
		t.Fatal("expected global n")
	}
	if n.Int != 42 {
		t.Fatalf("expected 42, got %d", n.Int)
	}
}

func TestTimeNow(t *testing.T) {
	src := `var t int = Time.now();
if (t <= 0) {
    Console.print("bad time");
}`
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
	tv, ok := res.VM.Global("t")
	if !ok || tv.Int <= 0 {
		t.Fatalf("expected positive unix time, got %v", tv)
	}
}

func TestArgsCLI(t *testing.T) {
	src := `
var n int = Args.count();
var a string = Args.at(0);
var b string = Args.at(1);
`
	res, err := CompileWithOptions(src, "test.sol", PhaseRun, RunOptions{ScriptArgs: []string{"hello", "world"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("errors: %v", res.Errors)
	}
	if res.RunErr != nil {
		t.Fatalf("run error: %v", res.RunErr)
	}
	n, _ := res.VM.Global("n")
	if n.Int != 2 {
		t.Fatalf("expected 2 args, got %d", n.Int)
	}
	a, _ := res.VM.Global("a")
	if a.StrVal != "hello" {
		t.Fatalf("expected hello, got %q", a.StrVal)
	}
	b, _ := res.VM.Global("b")
	if b.StrVal != "world" {
		t.Fatalf("expected world, got %q", b.StrVal)
	}
}

func TestStringMath(t *testing.T) {
	src := `
var len int = String.length("abc");
var trimmed string = String.trim("  x  ");
var parts [string] = String.split("a,b", ",");
var has bool = String.contains("hello", "ell");
var sub string = String.substring("hello", 1, 4);
var r float = Math.random();
var m float = Math.max(1.0, 2.0);
var f int = Math.floor(3.7);
`
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
	len, _ := res.VM.Global("len")
	if len.Int != 3 {
		t.Fatalf("length: %d", len.Int)
	}
	trimmed, _ := res.VM.Global("trimmed")
	if trimmed.StrVal != "x" {
		t.Fatalf("trim: %q", trimmed.StrVal)
	}
	has, _ := res.VM.Global("has")
	if !has.Bool {
		t.Fatal("contains expected true")
	}
	sub, _ := res.VM.Global("sub")
	if sub.StrVal != "ell" {
		t.Fatalf("substring: %q", sub.StrVal)
	}
	f, _ := res.VM.Global("f")
	if f.Int != 3 {
		t.Fatalf("floor: %d", f.Int)
	}
	m, _ := res.VM.Global("m")
	if m.Float != 2.0 {
		t.Fatalf("max: %v", m.Float)
	}
}

func TestFileAppendExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")
	src := fmt.Sprintf(`
File.write("%s", "a");
File.append("%s", "b");
var ok bool = File.exists("%s");
var s string = File.read("%s");
`, path, path, path, path)
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
	ok, _ := res.VM.Global("ok")
	if !ok.Bool {
		t.Fatal("expected exists true")
	}
	s, _ := res.VM.Global("s")
	if s.StrVal != "ab" {
		t.Fatalf("expected ab, got %q", s.StrVal)
	}
}

func TestEmitIRMainArgs(t *testing.T) {
	res, err := Compile(`Console.print("x");`, "test.sol", PhaseEmitIR)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.IR, "define i32 @main(i32 %argc, i8** %argv)") {
		t.Fatal("expected main with argc/argv")
	}
	if !strings.Contains(res.IR, "sol_args_init") {
		t.Fatal("expected sol_args_init call")
	}
	if !strings.Contains(res.IR, "sol_time_now") {
		t.Fatal("expected sol_time_now declare")
	}
}

func TestFileIO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	src := fmt.Sprintf(`
File.write("%s", "hello file");
var s string = File.read("%s");
`, path, path)
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
	if s.StrVal != "hello file" {
		t.Fatalf("expected hello file, got %q", s.StrVal)
	}
}

func runSourceWithStdin(t *testing.T, src, stdin string) *vm.VM {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l, "test.sol")
	prog := p.Parse()
	sem := semantic.New("test.sol")
	sem.Check(prog)
	if len(sem.Errors()) > 0 {
		t.Fatalf("semantic errors: %v", sem.Errors())
	}
	gen := tac.New(sem.Classes())
	gen.Build(prog)
	machine := vm.New(gen.Instructions(), sem.Classes())
	machine.SetStdin(strings.NewReader(stdin))
	if err := machine.Run(); err != nil {
		t.Fatalf("run error: %v", err)
	}
	return machine
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
rise X {
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
