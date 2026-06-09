package arraymethods

import "testing"

func TestLookupPush(t *testing.T) {
	m, ok := Lookup("push")
	if !ok || !m.Mutates || m.ReturnType != "void" {
		t.Fatalf("push: %+v ok=%v", m, ok)
	}
}

func TestLookupMissing(t *testing.T) {
	_, ok := Lookup("sort")
	if ok {
		t.Fatal("expected missing method")
	}
}
