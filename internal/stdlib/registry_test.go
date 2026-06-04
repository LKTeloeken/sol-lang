package stdlib

import "testing"

func TestIsBuiltin(t *testing.T) {
	if !IsBuiltin("Time") {
		t.Fatal("Time should be builtin")
	}
	if IsBuiltin("Greeter") {
		t.Fatal("Greeter should not be builtin")
	}
}

func TestLookup(t *testing.T) {
	m, ok := Lookup("File", "append")
	if !ok || m.ReturnType.Base != "void" {
		t.Fatalf("File.append: %+v ok=%v", m, ok)
	}
	_, ok = Lookup("File", "missing")
	if ok {
		t.Fatal("expected missing method")
	}
}

func TestBuiltinID(t *testing.T) {
	if BuiltinID("Math", "random") != "Math.random" {
		t.Fatal("unexpected id")
	}
}
