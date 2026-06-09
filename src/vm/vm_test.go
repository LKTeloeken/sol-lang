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
	arr := Arr(Int(1), Int(2), Int(3))
	vm := New(nil, nil)
	vm.globals["nums"] = arr
	v, err := vm.resolveValue("nums.length")
	if err != nil {
		t.Fatal(err)
	}
	if v.Int != 3 {
		t.Fatalf("expected length 3, got %d", v.Int)
	}
}

func TestArrayIndex(t *testing.T) {
	arr := Arr(Int(10), Int(20))
	vm := New(nil, nil)
	vm.globals["nums"] = arr
	v, err := vm.resolveValue("nums[1]")
	if err != nil {
		t.Fatal(err)
	}
	if v.Int != 20 {
		t.Fatalf("expected 20, got %d", v.Int)
	}
}

func TestArrayPushWriteBack(t *testing.T) {
	obj := Obj("Box")
	obj.Object.Fields["items"] = Arr(Str("a"))
	vm := New(nil, nil)
	vm.globals["box"] = obj

	ins := tac.Instr{Op: tac.OpArrayCall, Arg1: "box.items", Arg2: "push", Arg3: "1"}
	vm.params = []Value{Str("b")}
	if err := vm.doArrayCall(ins); err != nil {
		t.Fatal(err)
	}
	updated, err := vm.GetField("box", "items")
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Array) != 2 || updated.Array[1].StrVal != "b" {
		t.Fatalf("expected [a b], got %v", updated.Array)
	}
}
