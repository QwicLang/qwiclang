package lexer

import (
	"fmt"
	"testing"
)

type expectedToken struct {
	kind   string
	lexeme string
}

func TestLexHelloWorld(t *testing.T) {
	assertLexes(t, `public func main() {
    print("Hello, Qwic")
}
`, []expectedToken{
		{"Public", "public"},
		{"Func", "func"},
		{"Identifier", "main"},
		{"LParen", "("},
		{"RParen", ")"},
		{"LBrace", "{"},
		{"Newline", "\n"},
		{"Identifier", "print"},
		{"LParen", "("},
		{"String", `"Hello, Qwic"`},
		{"RParen", ")"},
		{"Newline", "\n"},
		{"RBrace", "}"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexVariables(t *testing.T) {
	assertLexes(t, `const name: string = "Qwic"
let count: int = 0
count = count + 1
`, []expectedToken{
		{"Const", "const"},
		{"Identifier", "name"},
		{"Colon", ":"},
		{"Identifier", "string"},
		{"Assign", "="},
		{"String", `"Qwic"`},
		{"Newline", "\n"},
		{"Let", "let"},
		{"Identifier", "count"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Assign", "="},
		{"Integer", "0"},
		{"Newline", "\n"},
		{"Identifier", "count"},
		{"Assign", "="},
		{"Identifier", "count"},
		{"Plus", "+"},
		{"Integer", "1"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexFunctionDeclarations(t *testing.T) {
	assertLexes(t, `private turbo func add(a: int, b: int): int {
    return a + b
}
`, []expectedToken{
		{"Private", "private"},
		{"Turbo", "turbo"},
		{"Func", "func"},
		{"Identifier", "add"},
		{"LParen", "("},
		{"Identifier", "a"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Comma", ","},
		{"Identifier", "b"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"RParen", ")"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"LBrace", "{"},
		{"Newline", "\n"},
		{"Return", "return"},
		{"Identifier", "a"},
		{"Plus", "+"},
		{"Identifier", "b"},
		{"Newline", "\n"},
		{"RBrace", "}"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexNumbers(t *testing.T) {
	assertLexes(t, `const whole: int = 42
const ratio: float = 3.14
const delta: nano = 0.016666
`, []expectedToken{
		{"Const", "const"},
		{"Identifier", "whole"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Assign", "="},
		{"Integer", "42"},
		{"Newline", "\n"},
		{"Const", "const"},
		{"Identifier", "ratio"},
		{"Colon", ":"},
		{"Identifier", "float"},
		{"Assign", "="},
		{"Float", "3.14"},
		{"Newline", "\n"},
		{"Const", "const"},
		{"Identifier", "delta"},
		{"Colon", ":"},
		{"Identifier", "nano"},
		{"Assign", "="},
		{"Float", "0.016666"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexStrings(t *testing.T) {
	assertLexes(t, `"plain"
"with \"quotes\""
"line\nbreak"
`, []expectedToken{
		{"String", `"plain"`},
		{"Newline", "\n"},
		{"String", `"with \"quotes\""`},
		{"Newline", "\n"},
		{"String", `"line\nbreak"`},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexOperators(t *testing.T) {
	assertLexes(t, `a + b - c * d / e % f
a == b != c < d <= e > f >= g
ready && enabled || fallback
!ready
`, []expectedToken{
		{"Identifier", "a"},
		{"Plus", "+"},
		{"Identifier", "b"},
		{"Minus", "-"},
		{"Identifier", "c"},
		{"Star", "*"},
		{"Identifier", "d"},
		{"Slash", "/"},
		{"Identifier", "e"},
		{"Percent", "%"},
		{"Identifier", "f"},
		{"Newline", "\n"},
		{"Identifier", "a"},
		{"Equal", "=="},
		{"Identifier", "b"},
		{"NotEqual", "!="},
		{"Identifier", "c"},
		{"Less", "<"},
		{"Identifier", "d"},
		{"LessEqual", "<="},
		{"Identifier", "e"},
		{"Greater", ">"},
		{"Identifier", "f"},
		{"GreaterEqual", ">="},
		{"Identifier", "g"},
		{"Newline", "\n"},
		{"Identifier", "ready"},
		{"And", "&&"},
		{"Identifier", "enabled"},
		{"Or", "||"},
		{"Identifier", "fallback"},
		{"Newline", "\n"},
		{"Not", "!"},
		{"Identifier", "ready"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexComments(t *testing.T) {
	assertLexes(t, `// leading comment
const x: int = 1 // trailing comment
/*
    block comment
*/
let y: int = x
`, []expectedToken{
		{"Newline", "\n"},
		{"Const", "const"},
		{"Identifier", "x"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Assign", "="},
		{"Integer", "1"},
		{"Newline", "\n"},
		{"Newline", "\n"},
		{"Newline", "\n"},
		{"Newline", "\n"},
		{"Let", "let"},
		{"Identifier", "y"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Assign", "="},
		{"Identifier", "x"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexNewlinesAndSemicolons(t *testing.T) {
	assertLexes(t, "const x: int = 10;\nconst y: int = 20\n\nprint(x);\n", []expectedToken{
		{"Const", "const"},
		{"Identifier", "x"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Assign", "="},
		{"Integer", "10"},
		{"Semicolon", ";"},
		{"Newline", "\n"},
		{"Const", "const"},
		{"Identifier", "y"},
		{"Colon", ":"},
		{"Identifier", "int"},
		{"Assign", "="},
		{"Integer", "20"},
		{"Newline", "\n"},
		{"Newline", "\n"},
		{"Identifier", "print"},
		{"LParen", "("},
		{"Identifier", "x"},
		{"RParen", ")"},
		{"Semicolon", ";"},
		{"Newline", "\n"},
		{"EOF", ""},
	})
}

func TestLexKeywords(t *testing.T) {
	assertLexes(t, `public private const let func return if else while struct import module turbo true false null`, []expectedToken{
		{"Public", "public"},
		{"Private", "private"},
		{"Const", "const"},
		{"Let", "let"},
		{"Func", "func"},
		{"Return", "return"},
		{"If", "if"},
		{"Else", "else"},
		{"While", "while"},
		{"Struct", "struct"},
		{"Import", "import"},
		{"Module", "module"},
		{"Turbo", "turbo"},
		{"True", "true"},
		{"False", "false"},
		{"Null", "null"},
		{"EOF", ""},
	})
}

func TestLexReportsUnterminatedString(t *testing.T) {
	l := New("test.qw", `"unterminated`)
	_, diagnostics := l.Lex()
	if len(diagnostics) == 0 {
		t.Fatal("expected at least one diagnostic for unterminated string")
	}
}

func assertLexes(t *testing.T, source string, expected []expectedToken) {
	t.Helper()

	l := New("test.qw", source)
	tokens, diagnostics := l.Lex()
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %d: %v", len(diagnostics), diagnostics)
	}

	if len(tokens) != len(expected) {
		t.Fatalf("token count mismatch: got %d, want %d\nactual: %#v", len(tokens), len(expected), tokens)
	}

	for i, want := range expected {
		gotKind := fmt.Sprint(tokens[i].Kind)
		if gotKind != want.kind {
			t.Fatalf("token %d kind mismatch: got %q, want %q", i, gotKind, want.kind)
		}
		if tokens[i].Lexeme != want.lexeme {
			t.Fatalf("token %d lexeme mismatch: got %q, want %q", i, tokens[i].Lexeme, want.lexeme)
		}
	}
}
