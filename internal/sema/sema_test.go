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

func TestCheckAllowsInterpolatedStringExpressions(t *testing.T) {
	checkValid(t, `func main() {
    const name: string = "Qwic"
    const count: int = 2
    const message: string = f"{name} has {count + 1} values"
    print(message)
}
`)
}

func TestCheckRejectsUnknownInterpolatedVariable(t *testing.T) {
	assertDiagnostic(t, `func main() {
    print(f"Hello, {missing}")
}
`, "unknown variable \"missing\"")
}

func TestCheckAllowsImportedStandardStringsPackage(t *testing.T) {
	checkValid(t, `import strings

func main() {
    const clean: string = strings.trim("  QwicLang  ")
    const length: int = strings.length(clean)
    const hasLang: bool = strings.contains(clean, "Lang")
    print(length)
    print(hasLang)
}
`)
}

func TestCheckRejectsStandardPackageWithoutImport(t *testing.T) {
	assertDiagnostic(t, `func main() {
    print(strings.trim("  QwicLang  "))
}
`, "module \"strings\" is not imported")
}

func TestCheckRejectsWrongStandardPackageArgumentType(t *testing.T) {
	assertDiagnostic(t, `import strings

func main() {
    print(strings.length(42))
}
`, "cannot pass argument of type \"int\" to parameter \"value\" of type \"string\"")
}

func TestCheckAllowsImportedDataStructurePackages(t *testing.T) {
	checkValid(t, `import lists
import sets
import dictionaries
import tuples

func main() {
    const names: list = lists.new()
    lists.push(names, "Qwic")
    const first: string = lists.get(names, 0)

    const unique: set = sets.new()
    sets.add(unique, first)
    const hasQwic: bool = sets.contains(unique, "Qwic")

    const values: dictionary = dictionaries.new()
    dictionaries.set(values, "name", first)
    const name: string = dictionaries.get(values, "name")

    const pair: tuple = tuples.new2(name, "Lang")
    print(tuples.first(pair))
    print(hasQwic)
}
`)
}

func TestCheckRejectsDataStructurePackageWithoutImport(t *testing.T) {
	assertDiagnostic(t, `func main() {
    const names: list = lists.new()
}
`, "module \"lists\" is not imported")
}

func TestCheckRejectsWrongDataStructureValueType(t *testing.T) {
	assertDiagnostic(t, `import lists

func main() {
    const values: list = lists.new()
    lists.push(values, 42)
}
`, "cannot pass argument of type \"int\" to parameter \"value\" of type \"string\"")
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
