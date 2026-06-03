package token

import "testing"

func TestLookupIdentKeywords(t *testing.T) {
	tests := map[string]Type{
		"rise": RISE, "ray": RAY, "glow": GLOW, "enlights": ENLIGHTS,
		"emit": EMIT, "flare": FLARE, "int": INT_TYPE, "float": FLOAT_TYPE,
		"bool": BOOL_TYPE, "string": STRING_TYPE, "void": VOID_TYPE,
		"foobar": IDENT,
	}
	for lit, want := range tests {
		if got := LookupIdent(lit); got != want {
			t.Errorf("LookupIdent(%q) = %v, want %v", lit, got, want)
		}
	}
}

func TestTypeString(t *testing.T) {
	if RISE.String() != "rise" {
		t.Errorf("RISE.String() = %q", RISE.String())
	}
	if ASSIGN.String() != "=" {
		t.Errorf("ASSIGN.String() = %q", ASSIGN.String())
	}
}
