package sema

import (
	"strings"
	"testing"

	"qwiclang/internal/types"
)

func TestCheckAcceptsValidFunctionCallsAndReturn(t *testing.T) {
	result := checkValid(t, `func add(a: int, b: int): int {
    return a + b
}

public func main() {
    const result: int = add(20, 22)
    print(result)
}
`)

	add := result.Functions["add"]
	if add.ReturnType.Kind != types.Int {
		t.Fatalf("add return type = %s, want int", add.ReturnType)
	}
	if len(add.Parameters) != 2 {
		t.Fatalf("add parameter count = %d, want 2", len(add.Parameters))
	}
	main := result.Functions["main"]
	if main.Visibility.String() != "public" {
		t.Fatalf("main visibility = %s, want public", main.Visibility)
	}
}

func TestCheckPreservesTurboFunctionMarker(t *testing.T) {
	result := checkValid(t, `turbo func calculate(value: int): int {
    return value * 2
}
`)

	if !result.Functions["calculate"].Turbo {
		t.Fatal("expected turbo marker on function symbol")
	}
}

func TestCheckRejectsInvalidAssignmentType(t *testing.T) {
	assertDiagnostic(t, `func main() {
    let x: int = "hello"
}
`, "cannot assign value of type \"string\" to variable of type \"int\"")
}

func TestCheckRejectsConstAssignment(t *testing.T) {
	assertDiagnostic(t, `func main() {
    const x: int = 10
    x = 20
}
`, "cannot assign to const variable \"x\"")
}

func TestCheckRejectsUnknownFunction(t *testing.T) {
	assertDiagnostic(t, `func main() {
    unknownFunction()
}
`, "unknown function \"unknownFunction\"")
}

func TestCheckRejectsUnknownVariable(t *testing.T) {
	assertDiagnostic(t, `func main() {
    print(user)
}
`, "unknown variable \"user\"")
}

func TestCheckRejectsInvalidReturnType(t *testing.T) {
	assertDiagnostic(t, `func answer(): int {
    return "forty two"
}
`, "cannot return value of type \"string\" from function returning \"int\"")
}

func TestCheckRejectsWrongArgumentType(t *testing.T) {
	assertDiagnostic(t, `func takesInt(value: int) {
}

func main() {
    takesInt("no")
}
`, "cannot pass argument of type \"string\" to parameter \"value\" of type \"int\"")
}

func TestCheckRejectsNonBoolCondition(t *testing.T) {
	assertDiagnostic(t, `func main() {
    if 1 {
        print("bad")
    }
}
`, "condition must be bool, got \"int\"")
}

func TestCheckAllowsNanoDeclarationFromFloatLiteral(t *testing.T) {
	checkValid(t, `func main() {
    const delta: nano = 0.016666
}
`)
}

func TestCheckAllowsNanoArithmeticWithIntegerScalar(t *testing.T) {
	checkValid(t, `func main() {
    const delta: nano = 0.016666
    const value: nano = delta * 2
}
`)
}

func TestCheckAllowsPublicImportedModuleFunction(t *testing.T) {
	result, diagnostics := CheckFiles([]SourceFile{
		{Filename: "main.qw", Source: `import users

public func main() {
    users.createUser()
}
`},
		{Filename: "users.qw", Source: `module users

public func createUser() {
    print("created")
}
`},
	})
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	if _, ok := result.Functions["users.createUser"]; !ok {
		t.Fatalf("expected qualified users.createUser symbol, got %#v", result.Functions)
	}
}

func TestCheckRejectsPrivateImportedModuleFunction(t *testing.T) {
	_, diagnostics := CheckFiles([]SourceFile{
		{Filename: "main.qw", Source: `import users

public func main() {
    users.validateUser()
}
`},
		{Filename: "users.qw", Source: `module users

private func validateUser() {
}
`},
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected private access diagnostic")
	}
	assertContainsDiagnostic(t, diagnostics, "function \"validateUser\" is private to module \"users\"")
}

func TestCheckRejectsUnimportedModuleFunction(t *testing.T) {
	_, diagnostics := CheckFiles([]SourceFile{
		{Filename: "main.qw", Source: `public func main() {
    users.createUser()
}
`},
		{Filename: "users.qw", Source: `module users

public func createUser() {
}
`},
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected missing import diagnostic")
	}
	assertContainsDiagnostic(t, diagnostics, "module \"users\" is not imported")
}

func TestCheckDiagnosticsContainSourceLocation(t *testing.T) {
	_, diagnostics := Check("bad.qw", `func main() {
    unknownFunction()
}
`)
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic")
	}
	if diagnostics[0].Position.Filename != "bad.qw" {
		t.Fatalf("filename = %q, want bad.qw", diagnostics[0].Position.Filename)
	}
	if diagnostics[0].Position.Line == 0 || diagnostics[0].Position.Column == 0 {
		t.Fatalf("diagnostic has invalid position: %#v", diagnostics[0].Position)
	}
}

func checkValid(t *testing.T, source string) *Result {
	t.Helper()

	result, diagnostics := Check("test.qw", source)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	return result
}

func assertDiagnostic(t *testing.T, source string, want string) {
	t.Helper()

	_, diagnostics := Check("test.qw", source)
	if len(diagnostics) == 0 {
		t.Fatalf("expected diagnostic containing %q", want)
	}
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Message, want) {
			return
		}
	}
	t.Fatalf("expected diagnostic containing %q, got %v", want, diagnostics)
}

func assertContainsDiagnostic(t *testing.T, diagnostics []Diagnostic, want string) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Message, want) {
			return
		}
	}
	t.Fatalf("expected diagnostic containing %q, got %v", want, diagnostics)
}
