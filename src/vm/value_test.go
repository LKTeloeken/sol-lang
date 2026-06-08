package vm

import "testing"

func TestParseLiteral(t *testing.T) {
	v, err := parseLiteral(`"hello"`)
	if err != nil || v.StrVal != "hello" {
		t.Fatalf("got %v err %v", v, err)
	}
	v, err = parseLiteral("42")
	if err != nil || v.Int != 42 {
		t.Fatalf("got %v err %v", v, err)
	}
}
