package vm

import "testing"

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
