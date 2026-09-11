package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qwiclang/internal/ir"
)

func TestGenerateCHelloWorld(t *testing.T) {
	module := buildIR(t, `public func main() {
    print("Hello, Qwic")
}
`)

	source, diagnostics := GenerateC(module)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	for _, want := range []string{
		`#include "qwic_runtime.h"`,
		"int main(void)",
		"qwic_runtime_init();",
		"qwic_print_string(qw_tmp_1);",
		`const char * qw_tmp_1 = "Hello, Qwic";`,
		"return qwic_runtime_exit_code();",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated C missing %q:\n%s", want, source)
		}
	}
}

func TestBuildExecutableRunsHelloWorld(t *testing.T) {
	module := buildIR(t, `public func main() {
    print("Hello, Qwic")
}
`)
	outputPath := filepath.Join(t.TempDir(), "hello")
	if diagnostics := BuildExecutable(module, Options{OutputPath: outputPath, RuntimePath: runtimePath(t)}); len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}

	exitCode, output, err := RunExecutable(outputPath)
	if err != nil {
		t.Fatalf("run executable: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; output: %s", exitCode, output)
	}
	if output != "Hello, Qwic\n" {
		t.Fatalf("output = %q, want hello line", output)
	}
}

func TestBuildExecutableRunsComputation(t *testing.T) {
	module := buildIR(t, `func add(a: int, b: int): int {
    return a + b
}

public func main() {
    print(add(20, 22))
}
`)
	outputPath := filepath.Join(t.TempDir(), "answer")
	if diagnostics := BuildExecutable(module, Options{OutputPath: outputPath, RuntimePath: runtimePath(t)}); len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}

	exitCode, output, err := RunExecutable(outputPath)
	if err != nil {
		t.Fatalf("run executable: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; output: %s", exitCode, output)
	}
	if output != "42\n" {
		t.Fatalf("output = %q, want 42 line", output)
	}
}

func TestGenerateCTurboFunction(t *testing.T) {
	module := buildIR(t, `turbo func calculate(value: int): int {
    return value * 2
}

public func main() {
    print(calculate(10))
}
`)

	source, diagnostics := GenerateC(module)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	want := "static inline __attribute__((always_inline, hot, optimize(\"O3\")))"
	if !strings.Contains(source, want) {
		t.Fatalf("generated C missing turbo optimization attribute %q:\n%s", want, source)
	}
}

func TestBuildExecutableCanKeepGeneratedC(t *testing.T) {
	module := buildIR(t, `public func main() {
    print("Hello, Qwic")
}
`)
	workDir := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "hello")
	if diagnostics := BuildExecutable(module, Options{OutputPath: outputPath, WorkDir: workDir, KeepC: true, RuntimePath: runtimePath(t)}); len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	if _, err := os.Stat(filepath.Join(workDir, "main.c")); err != nil {
		t.Fatalf("expected generated C to be kept: %v", err)
	}
}

func TestBuildExecutableUsesRuntimeSource(t *testing.T) {
	module := buildIR(t, `public func main() {
    print(true)
}
`)
	outputPath := filepath.Join(t.TempDir(), "program")
	if diagnostics := BuildExecutable(module, Options{OutputPath: outputPath, RuntimePath: runtimePath(t)}); len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
}

func TestGenerateCInterpolatedString(t *testing.T) {
	module := buildIR(t, `public func main() {
    const name: string = "Qwic"
    const count: int = 2
    print(f"Hello, {name}: {count + 1}")
}
`)

	source, diagnostics := GenerateC(module)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	for _, want := range []string{
		`snprintf(NULL, 0, "Hello, %s: %lld"`,
		`qwic_alloc`,
		`qwic_print_string`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated C missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateCStandardStringsCall(t *testing.T) {
	module := buildIR(t, `import strings

public func main() {
    print(strings.trim("  QwicLang  "))
}
`)

	source, diagnostics := GenerateC(module)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	if !strings.Contains(source, "qwic_strings_trim") {
		t.Fatalf("generated C missing stdlib runtime call:\n%s", source)
	}
}

func TestGenerateCDataStructureCalls(t *testing.T) {
	module := buildIR(t, `import lists
import dictionaries

public func main() {
    const names: list = lists.new()
    lists.push(names, "Qwic")
    const values: dictionary = dictionaries.new()
    dictionaries.set(values, "name", lists.get(names, 0))
    print(dictionaries.get(values, "name"))
}
`)

	source, diagnostics := GenerateC(module)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	for _, want := range []string{"qwic_lists_new", "qwic_lists_push", "qwic_dictionaries_set"} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated C missing %q:\n%s", want, source)
		}
	}
}

func runtimePath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "runtime")
}

func buildIR(t *testing.T, source string) ir.Module {
	t.Helper()

	module, diagnostics := ir.BuildSource("test.qw", source)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	return module
}
