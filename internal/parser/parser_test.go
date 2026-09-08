package parser

import (
	"strings"
	"testing"

	"qwiclang/internal/ast"
)

func TestParsePhaseTwoCompletionExample(t *testing.T) {
	program := parseValid(t, `public func main() {
    const x: int = 10

    if x > 5 {
        print(x)
    }
}
`)

	if len(program.Declarations) != 1 {
		t.Fatalf("declaration count = %d, want 1", len(program.Declarations))
	}

	mainFunction, ok := program.Declarations[0].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("declaration type = %T, want *ast.FunctionDeclaration", program.Declarations[0])
	}
	if mainFunction.Name != "main" {
		t.Fatalf("function name = %q, want main", mainFunction.Name)
	}
	if mainFunction.Visibility != ast.VisibilityPublic {
		t.Fatalf("function visibility = %s, want public", mainFunction.Visibility)
	}
	if mainFunction.ReturnType != "void" {
		t.Fatalf("return type = %q, want void", mainFunction.ReturnType)
	}
	if len(mainFunction.Body.Statements) != 2 {
		t.Fatalf("statement count = %d, want 2", len(mainFunction.Body.Statements))
	}

	variable, ok := mainFunction.Body.Statements[0].(*ast.VariableDeclaration)
	if !ok {
		t.Fatalf("first statement type = %T, want *ast.VariableDeclaration", mainFunction.Body.Statements[0])
	}
	if variable.Mutable {
		t.Fatal("const declaration parsed as mutable")
	}
	if variable.Name != "x" || variable.TypeName != "int" {
		t.Fatalf("variable declaration = %#v, want name x type int", variable)
	}

	condition, ok := mainFunction.Body.Statements[1].(*ast.IfStatement)
	if !ok {
		t.Fatalf("second statement type = %T, want *ast.IfStatement", mainFunction.Body.Statements[1])
	}
	binary, ok := condition.Condition.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("condition type = %T, want *ast.BinaryExpression", condition.Condition)
	}
	if binary.Operator.String() != "Greater" {
		t.Fatalf("condition operator = %s, want Greater", binary.Operator)
	}
	if len(condition.ThenBranch.Statements) != 1 {
		t.Fatalf("then statement count = %d, want 1", len(condition.ThenBranch.Statements))
	}
}

func TestParseFunctionDeclarationWithParametersReturnTypeAndTurbo(t *testing.T) {
	program := parseValid(t, `private turbo func add(a: int, b: int): int {
    return a + b
}
`)

	function := program.Declarations[0].(*ast.FunctionDeclaration)
	if function.Visibility != ast.VisibilityPrivate {
		t.Fatalf("visibility = %s, want private", function.Visibility)
	}
	if !function.Turbo {
		t.Fatal("expected turbo modifier")
	}
	if function.ReturnType != "int" {
		t.Fatalf("return type = %q, want int", function.ReturnType)
	}
	if len(function.Parameters) != 2 {
		t.Fatalf("parameter count = %d, want 2", len(function.Parameters))
	}
	if function.Parameters[0].Name != "a" || function.Parameters[0].TypeName != "int" {
		t.Fatalf("first parameter = %#v, want a: int", function.Parameters[0])
	}
	if _, ok := function.Body.Statements[0].(*ast.ReturnStatement); !ok {
		t.Fatalf("body statement type = %T, want *ast.ReturnStatement", function.Body.Statements[0])
	}
}

func TestParseAssignmentCallWhileAndElse(t *testing.T) {
	program := parseValid(t, `func main() {
    let count: int = 0;
    while count < 10 {
        count = count + 1
    }
    if count == 10 {
        print("done")
    } else {
        print("miss")
    }
}
`)

	function := program.Declarations[0].(*ast.FunctionDeclaration)
	if len(function.Body.Statements) != 3 {
		t.Fatalf("statement count = %d, want 3", len(function.Body.Statements))
	}
	if _, ok := function.Body.Statements[0].(*ast.VariableDeclaration); !ok {
		t.Fatalf("first statement type = %T, want *ast.VariableDeclaration", function.Body.Statements[0])
	}
	loop, ok := function.Body.Statements[1].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("second statement type = %T, want *ast.WhileStatement", function.Body.Statements[1])
	}
	if _, ok := loop.Body.Statements[0].(*ast.AssignmentStatement); !ok {
		t.Fatalf("loop statement type = %T, want *ast.AssignmentStatement", loop.Body.Statements[0])
	}
	branch := function.Body.Statements[2].(*ast.IfStatement)
	if branch.ElseBranch == nil {
		t.Fatal("expected else branch")
	}
	thenCall := branch.ThenBranch.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	if thenCall.Callee.(*ast.IdentifierExpression).Name != "print" {
		t.Fatalf("call callee = %#v, want print identifier", thenCall.Callee)
	}
}

func TestParseExpressionPrecedence(t *testing.T) {
	program := parseValid(t, `func main() {
    const value: int = 1 + 2 * 3 == 7 && true
}
`)

	function := program.Declarations[0].(*ast.FunctionDeclaration)
	variable := function.Body.Statements[0].(*ast.VariableDeclaration)
	root, ok := variable.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.BinaryExpression", variable.Value)
	}
	if root.Operator.String() != "And" {
		t.Fatalf("root operator = %s, want And", root.Operator)
	}
	equality := root.Left.(*ast.BinaryExpression)
	if equality.Operator.String() != "Equal" {
		t.Fatalf("left operator = %s, want Equal", equality.Operator)
	}
	addition := equality.Left.(*ast.BinaryExpression)
	if addition.Operator.String() != "Plus" {
		t.Fatalf("addition operator = %s, want Plus", addition.Operator)
	}
	multiply := addition.Right.(*ast.BinaryExpression)
	if multiply.Operator.String() != "Star" {
		t.Fatalf("multiply operator = %s, want Star", multiply.Operator)
	}
}

func TestParseModuleImportAndQualifiedCall(t *testing.T) {
	program := parseValid(t, `module app
import users

public func main() {
    users.createUser()
}
`)

	if len(program.Declarations) != 3 {
		t.Fatalf("declaration count = %d, want 3", len(program.Declarations))
	}
	if module, ok := program.Declarations[0].(*ast.ModuleDeclaration); !ok || module.Name != "app" {
		t.Fatalf("first declaration = %#v, want module app", program.Declarations[0])
	}
	if imported, ok := program.Declarations[1].(*ast.ImportDeclaration); !ok || imported.Name != "users" {
		t.Fatalf("second declaration = %#v, want import users", program.Declarations[1])
	}
	function := program.Declarations[2].(*ast.FunctionDeclaration)
	statement := function.Body.Statements[0].(*ast.ExpressionStatement)
	call := statement.Expression.(*ast.CallExpression)
	selector := call.Callee.(*ast.SelectorExpression)
	if selector.Name != "createUser" {
		t.Fatalf("selector name = %q, want createUser", selector.Name)
	}
	if selector.Left.(*ast.IdentifierExpression).Name != "users" {
		t.Fatalf("selector module = %#v, want users", selector.Left)
	}
}

func TestParseReportsSourceLocatedSyntaxError(t *testing.T) {
	_, diagnostics := Parse("broken.qw", `func main( {
}
`)
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diagnostics[0].Position.Filename != "broken.qw" {
		t.Fatalf("filename = %q, want broken.qw", diagnostics[0].Position.Filename)
	}
	if !strings.Contains(diagnostics[0].Message, "expected parameter name") {
		t.Fatalf("diagnostic = %q, want parameter-name error", diagnostics[0].Message)
	}
}

func TestProgramDebugString(t *testing.T) {
	program := parseValid(t, `public func main() {
    print("Hello, Qwic")
}
`)

	debug := program.DebugString()
	for _, text := range []string{
		"Program",
		"FunctionDeclaration name=main visibility=public",
		"CallExpression",
		"LiteralExpression kind=string value=\"Hello, Qwic\"",
	} {
		if !strings.Contains(debug, text) {
			t.Fatalf("debug output missing %q:\n%s", text, debug)
		}
	}
}

func parseValid(t *testing.T, source string) *ast.Program {
	t.Helper()

	program, diagnostics := Parse("test.qw", source)
	if len(diagnostics) > 0 {
		t.Fatalf("expected no diagnostics, got %v", diagnostics)
	}
	return program
}
