package integration_test

import (
	"path/filepath"
	"testing"

	"qwiclang/internal/codegen"
	"qwiclang/internal/ir"
	"qwiclang/internal/sema"
)

func TestHelloProgramBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    print("Hello, Qwic")
}
`, "Hello, Qwic\n")
}

func TestComputationProgramBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `func add(a: int, b: int): int {
    return a + b
}

public func main() {
    print(add(20, 22))
}
`, "42\n")
}

func TestPhaseSixExpressionsFunctionsAndAssignments(t *testing.T) {
	assertProgramOutput(t, `func add(a: int, b: int): int {
    return a + b
}

func double(value: int): int {
    return value * 2
}

public func main() {
    let total: int = add(10, 20)
    total = total + double(6)
    print(total)
    print(total == 42 && true)
}
`, "42\ntrue\n")
}

func TestPhaseSevenIfElseBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    const value: int = 12

    if value > 10 {
        print("large")
    } else {
        print("small")
    }
}
`, "large\n")
}

func TestPhaseSevenWhileBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    let i: int = 0

    while i < 5 {
        print(i)
        i = i + 1
    }
}
`, "0\n1\n2\n3\n4\n")
}

func TestPhaseEightNanoBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    const delta: nano = 0.016666
    const value: nano = delta * 2

    print(value)
}
`, "0.033332\n")
}

func TestPhaseNineModulesBuildAndRun(t *testing.T) {
	files := []sema.SourceFile{
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
	}

	result, diagnostics := sema.CheckFiles(files)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no frontend diagnostics, got %v", diagnostics)
	}
	module, diagnostics := ir.Build(result)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no IR diagnostics, got %v", diagnostics)
	}
	assertModuleOutput(t, module, "created\n")
}

func TestRuntimeBackedPrintingBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    print(7)
    print(1.500000)
    print(true)
    print("runtime")
}
`, "7\n1.500000\ntrue\nruntime\n")
}

func TestInterpolatedStringsBuildAndRun(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    const name: string = "Qwic"
    const count: int = 2
    const ratio: float = 1.500000
    const ready: bool = true

    print(f"Hello, {name}: {count + 1}, {ratio}, {ready}")
}
`, "Hello, Qwic: 3, 1.500000, true\n")
}

func TestStandardStringsPackageBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `import strings

public func main() {
    const raw: string = "  QwicLang  "
    const clean: string = strings.trim(raw)

    print(strings.upper(clean))
    print(strings.length(clean))
    print(strings.contains(clean, "Lang"))
    print(strings.startsWith(clean, "Qwic"))
    print(strings.endsWith(clean, "Lang"))
    print(strings.indexOf(clean, "Lang"))
}
`, "QWICLANG\n8\ntrue\ntrue\ntrue\n4\n")
}

func TestDataStructurePackagesBuildAndRun(t *testing.T) {
	assertProgramOutput(t, `import lists
import sets
import dictionaries
import tuples

public func main() {
    const names: list = lists.new()
    lists.push(names, "Qwic")
    lists.push(names, "Lang")

    const unique: set = sets.new()
    sets.add(unique, "Qwic")
    sets.add(unique, "Qwic")

    const values: dictionary = dictionaries.new()
    dictionaries.set(values, "name", lists.get(names, 0))
    dictionaries.set(values, "kind", "language")

    const pair: tuple = tuples.new2(dictionaries.get(values, "name"), dictionaries.get(values, "kind"))

    print(lists.length(names))
    print(lists.get(names, 1))
    print(sets.length(unique))
    print(sets.contains(unique, "Qwic"))
    print(dictionaries.length(values))
    print(dictionaries.get(values, "kind"))
    print(f"{tuples.first(pair)}:{tuples.second(pair)}:{tuples.length(pair)}")
}
`, "2\nLang\n1\ntrue\n2\nlanguage\nQwic:language:2\n")
}

func TestForLoopBuildsAndRuns(t *testing.T) {
	assertProgramOutput(t, `import lists

public func main() {
    const items = ["alpha", "beta", "gamma"]
    for item in items {
        print(item)
    }
}
`, "alpha\nbeta\ngamma\n")
}

func TestPrintDataStructures(t *testing.T) {
	assertProgramOutput(t, `public func main() {
    const numbers = ["1", "2", "3"]
    print(numbers)

    const values = {"a": "1", "b": "2"}
    print(values)

    const pair = ("foo", "bar")
    print(pair)
}
`, "[1, 2, 3]\n{a: 1, b: 2}\n(foo, bar)\n")
}


func assertProgramOutput(t *testing.T, source string, want string) {
	t.Helper()

	module, diagnostics := ir.BuildSource("test.qw", source)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no frontend diagnostics, got %v", diagnostics)
	}
	assertModuleOutput(t, module, want)
}

func assertModuleOutput(t *testing.T, module ir.Module, want string) {
	t.Helper()

	outputPath := filepath.Join(t.TempDir(), "program")
	if diagnostics := codegen.BuildExecutable(module, codegen.Options{OutputPath: outputPath, RuntimePath: filepath.Join("..", "..", "runtime")}); len(diagnostics) > 0 {
		t.Fatalf("expected no codegen diagnostics, got %v", diagnostics)
	}

	exitCode, output, err := codegen.RunExecutable(outputPath)
	if err != nil {
		t.Fatalf("run executable: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; output: %s", exitCode, output)
	}
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}
