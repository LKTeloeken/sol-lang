package vm

import (
	"testing"

	"github.com/unisc/compiladores/sol/src/tac"
)

func TestStringConcatValues(t *testing.T) {
	left := Str("n=")
	right := Int(42)
	got := Str(formatPrintable(left) + formatPrintable(right))
	if got.StrVal != "n=42" {
		t.Fatalf("expected n=42, got %q", got.StrVal)
	}
}

func TestArrayLength(t *testing.T) {
	vm := New(nil, nil)
	vm.globals["nums"] = Arr(Int(1), Int(2), Int(3))
	arr, err := vm.resolveOperand(tac.Name("nums"))
	if err != nil {
		t.Fatal(err)
	}
	if arr.Kind != KindArray || len(arr.Array) != 3 {
		t.Fatalf("expected array of length 3, got %v", arr)
	}
}

func TestArrayIndex(t *testing.T) {
	vm := New(nil, nil)
	vm.globals["nums"] = Arr(Int(10), Int(20))
	v, err := vm.loadIndex(tac.Name("nums"), tac.ConstInt(1))
	if err != nil {
		t.Fatal(err)
	}
	if v.Int != 20 {
		t.Fatalf("expected 20, got %d", v.Int)
	}
}

func TestArrayPushWriteBack(t *testing.T) {
	vm := New(nil, nil)
	vm.globals["items"] = Arr(Str("a"))

	ins := tac.Instr{Op: tac.OpArrayCall, Obj: tac.Name("items"), Sym: "push", NArgs: 1}
	vm.params = []Value{Str("b")}
	if err := vm.doArrayCall(ins); err != nil {
		t.Fatal(err)
	}
	updated, ok := vm.Global("items")
	if !ok {
		t.Fatal("items not found")
	}
	if len(updated.Array) != 2 || updated.Array[1].StrVal != "b" {
		t.Fatalf("expected [a b], got %v", updated.Array)
	}
}
