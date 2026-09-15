package lsp

import (
	"testing"
)

func TestCollectDiagnosticsFindsErrors(t *testing.T) {
	source := `public func main() {
    let x: int = "string"
}`
	diags := collectDiagnostics("test.qw", source)
	if len(diags) == 0 {
		t.Fatal("expected diagnostic errors for type mismatch, got none")
	}
	if diags[0].Message == "" {
		t.Fatal("expected diagnostic message")
	}
}

func TestCollectDiagnosticsCleanForValidSource(t *testing.T) {
	source := `public func main() {
    print("hello")
}`
	diags := collectDiagnostics("test.qw", source)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for valid code, got %v", diags)
	}
}

func TestGetWordBeforePosition(t *testing.T) {
	content := "import math\nmath.sq"
	prefix, isDot, pkg := getWordBeforePosition(content, 1, 7)
	if isDot || pkg != "" || prefix != "sq" {
		t.Fatalf("expected prefix=sq, got prefix=%q isDot=%v pkg=%q", prefix, isDot, pkg)
	}

	contentDot := "import math\nmath."
	prefix, isDot, pkg = getWordBeforePosition(contentDot, 1, 5)
	if !isDot || pkg != "math" {
		t.Fatalf("expected isDot=true pkg=math, got isDot=%v pkg=%q prefix=%q", isDot, pkg, prefix)
	}
}

func TestGetIdentifierAtPosition(t *testing.T) {
	content := "import math\npublic func main() {\n    math.abs(-42.0)\n}"
	word, qual := getIdentifierAtPosition(content, 2, 9)
	if word != "abs" || qual != "math.abs" {
		t.Fatalf("expected word=abs, qual=math.abs; got word=%q qual=%q", word, qual)
	}
}

func TestServerProtocolHandling(t *testing.T) {
	server := NewServer("v0.1.0-alpha.1")
	docURI := "file:///test.qw"
	docText := "import math\npublic func main() {\n    math.abs(-5.0)\n}"

	server.documents.Store(docURI, docText)

	// Test completion
	items := server.complete(docURI, Position{Line: 2, Character: 9})
	if len(items) == 0 {
		t.Fatal("expected completion items for math., got 0")
	}

	// Test hover
	hover := server.hover(docURI, Position{Line: 2, Character: 10})
	if hover == nil {
		t.Fatal("expected hover for math.abs, got nil")
	}
	markup, ok := hover.Contents.(MarkupContent)
	if !ok || len(markup.Value) == 0 {
		t.Fatalf("expected valid hover markup, got %v", hover.Contents)
	}

	// Test formatting
	edits := server.formatDocument(docURI)
	// Even if already formatted or edits returned, no panic occurs
	_ = edits

	// Test definition
	loc := server.definition(docURI, Position{Line: 1, Character: 12}) // "main"
	if loc == nil {
		t.Fatal("expected definition location for main function, got nil")
	}
}
