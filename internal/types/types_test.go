package types

import "testing"

func TestLookupBuiltins(t *testing.T) {
	for _, name := range []string{"void", "bool", "int", "float", "nano", "string"} {
		if _, ok := Lookup(name); !ok {
			t.Fatalf("expected builtin type %q", name)
		}
	}
}

func TestCompatibleAllowsExactMatchesAndNanoFloatLiterals(t *testing.T) {
	if !Compatible(IntType, IntType) {
		t.Fatal("expected int to be compatible with int")
	}
	if !Compatible(NanoType, FloatType) {
		t.Fatal("expected float literal type to be compatible with nano")
	}
	if Compatible(IntType, StringType) {
		t.Fatal("did not expect string to be compatible with int")
	}
}
