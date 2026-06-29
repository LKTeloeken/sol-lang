package vm

import "testing"

func TestValueString(t *testing.T) {
	cases := []struct {
		v    Value
		want string
	}{
		{Int(42), "42"},
		{Float(3.5), "3.5"},
		{Bool(true), "true"},
		{Str("hi"), "hi"},
		{Null(), "null"},
	}
	for _, c := range cases {
		if got := c.v.String(); got != c.want {
			t.Errorf("String() = %q, want %q", got, c.want)
		}
	}
}

func TestValueAsFloat(t *testing.T) {
	if Int(7).AsFloat() != 7 {
		t.Fatal("int AsFloat")
	}
	if Bool(true).AsFloat() != 1 {
		t.Fatal("bool AsFloat")
	}
}
