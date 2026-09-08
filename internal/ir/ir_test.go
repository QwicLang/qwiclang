package ir

import (
	"strings"
	"testing"

	"qwiclang/internal/types"
)

func TestBuildFunctionVariablesBinaryCallAndReturn(t *testing.T) {
	module := buildValid(t, `func add(a: int, b: int): int {
    const result: int = a + b
    return result
}

public func main() {
    print(add(20, 22))
}
`)

	if len(module.Functions) != 2 {
		t.Fatalf("function count = %d, want 2", len(module.Functions))
	}
	add := module.Functions[0]
	if add.Name != "add" || add.ReturnType.Kind != types.Int {
		t.Fatalf("add function = %#v, want int-returning add", add)
	}
	if len(add.Parameters) != 2 {
		t.Fatalf("parameter count = %d, want 2", len(add.Parameters))
	}

	debug := module.DebugString()
	for _, want := range []string{
		"Function name=add",
		"const result: int",
		"load a: int",
		"binary Plus",
		"store result",
		"return",
		"call add",
		"call print",
	} {
		if !strings.Contains(debug, want) {
			t.Fatalf("IR debug output missing %q:\n%s", want, debug)
		}
	}
}

func TestBuildBranchBlocksForIfElse(t *testing.T) {
	module := buildValid(t, `func main() {
    const ready: bool = true
    if ready {
        print("yes")
    } else {
        print("no")
    }
}
`)

	debug := module.DebugString()
	for _, want := range []string{
		"branch",
		"Block name=if.then",
		"Block name=if.else",
		"Block name=if.end",
	} {
		if !strings.Contains(debug, want) {
			t.Fatalf("IR debug output missing %q:\n%s", want, debug)
		}
	}
}

func TestBuildBranchBlocksForWhile(t *testing.T) {
	module := buildValid(t, `func main() {
    let i: int = 0
    while i < 3 {
        i = i + 1
    }
}
`)

	debug := module.DebugString()
	for _, want := range []string{
		"jump while.cond",
		"Block name=while.cond",
		"Block name=while.body",
		"Block name=while.end",
		"binary Less",
		"binary Plus",
		"store i",
	} {
		if !strings.Contains(debug, want) {
			t.Fatalf("IR debug output missing %q:\n%s", want, debug)
		}
	}
}

func TestBuildSourceReturnsSemanticDiagnostics(t *testing.T) {
	_, diagnostics := BuildSource("bad.qw", `func main() {
    unknownFunction()
}
`)
	if len(diagnostics) == 0 {
		t.Fatal("expected semantic diagnostic")
	}
	if !strings.Contains(diagnostics[0].Message, "unknown function") {
		t.Fatalf("diagnostic = %q, want unknown function", diagnostics[0].Message)
	}
}

func buildValid(t *testing.T, source string) Module {
	t.Helper()

	module, diagnostics := BuildSource("test.qw", source)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	return module
}
