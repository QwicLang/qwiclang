package format

import "testing"

func TestSourceFormatsBasicProgram(t *testing.T) {
	formatted, diagnostics := Source("test.qw", `public func main(){print("Hello, Qwic");if true{print(1)}}`)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}

	want := `public func main() {
    print("Hello, Qwic")
    if true {
        print(1)
    }
}
`
	if formatted != want {
		t.Fatalf("formatted source:\n%s\nwant:\n%s", formatted, want)
	}
}

func TestSourceDoesNotDoubleSpaceTypeAnnotationsOrArguments(t *testing.T) {
	formatted, diagnostics := Source("test.qw", `func add(a:int,b:int):int{return a+b}`)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}

	want := `func add(a: int, b: int): int {
    return a + b
}
`
	if formatted != want {
		t.Fatalf("formatted source:\n%s\nwant:\n%s", formatted, want)
	}
}

func TestSourceReportsLexerDiagnostics(t *testing.T) {
	_, diagnostics := Source("test.qw", `"unterminated`)
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic")
	}
}
