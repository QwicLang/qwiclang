package sema

import (
	"fmt"

	"qwiclang/internal/ast"
	"qwiclang/internal/diagnostic"
	"qwiclang/internal/parser"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Diagnostic = diagnostic.Diagnostic

type Result struct {
	Program       *ast.Program
	Programs      []*ast.Program
	Functions     map[string]FunctionSymbol
	FunctionNames map[*ast.FunctionDeclaration]string
}

type FunctionSymbol struct {
	Name       string
	Module     string
	Qualified  string
	Visibility ast.Visibility
	Turbo      bool
	Parameters []ParameterSymbol
	ReturnType types.Type
	Builtin    bool
	Variadic   bool
	Pos        token.Position
}

type ParameterSymbol struct {
	Name string
	Type types.Type
	Pos  token.Position
}

type variableSymbol struct {
	name    string
	typ     types.Type
	mutable bool
	pos     token.Position
}

type scope struct {
	parent    *scope
	variables map[string]variableSymbol
}

type Checker struct {
	diagnostics   []Diagnostic
	functions     map[string]FunctionSymbol
	functionNames map[*ast.FunctionDeclaration]string
	files         map[*ast.Program]fileInfo
	currentFunc   *FunctionSymbol
	currentModule string
}

type SourceFile struct {
	Filename string
	Source   string
}

type fileInfo struct {
	module  string
	imports map[string]token.Position
}

func Check(filename, source string) (*Result, []Diagnostic) {
	return CheckFiles([]SourceFile{{Filename: filename, Source: source}})
}

func CheckFiles(files []SourceFile) (*Result, []Diagnostic) {
	programs := make([]*ast.Program, 0, len(files))
	diagnostics := []Diagnostic{}
	for _, file := range files {
		program, parseDiagnostics := parser.Parse(file.Filename, file.Source)
		programs = append(programs, program)
		for _, diagnostic := range parseDiagnostics {
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	if len(diagnostics) > 0 {
		return &Result{
			Program:       firstProgram(programs),
			Programs:      programs,
			Functions:     map[string]FunctionSymbol{},
			FunctionNames: map[*ast.FunctionDeclaration]string{},
		}, diagnostics
	}
	return NewChecker().CheckPrograms(programs)
}

func firstProgram(programs []*ast.Program) *ast.Program {
	if len(programs) == 0 {
		return &ast.Program{}
	}
	return programs[0]
}

func NewChecker() *Checker {
	checker := &Checker{
		functions:     map[string]FunctionSymbol{},
		functionNames: map[*ast.FunctionDeclaration]string{},
		files:         map[*ast.Program]fileInfo{},
	}
	checker.functions["print"] = FunctionSymbol{
		Name:       "print",
		Qualified:  "print",
		Parameters: []ParameterSymbol{{Name: "value", Type: types.AnyType}},
		ReturnType: types.VoidType,
		Builtin:    true,
		Pos:        token.Position{Line: 1, Column: 1},
	}
	return checker
}

func (checker *Checker) CheckProgram(program *ast.Program) (*Result, []Diagnostic) {
	return checker.CheckPrograms([]*ast.Program{program})
}

func (checker *Checker) CheckPrograms(programs []*ast.Program) (*Result, []Diagnostic) {
	for _, program := range programs {
		checker.collectFileInfo(program)
	}
	for _, program := range programs {
		checker.collectFunctions(program)
	}
	for _, program := range programs {
		for _, declaration := range program.Declarations {
			function, ok := declaration.(*ast.FunctionDeclaration)
			if !ok {
				continue
			}
			checker.checkFunction(function, checker.files[program].module)
		}
	}

	return &Result{
		Program:       firstProgram(programs),
		Programs:      programs,
		Functions:     checker.functions,
		FunctionNames: checker.functionNames,
	}, checker.diagnostics
}

func (checker *Checker) collectFileInfo(program *ast.Program) {
	info := fileInfo{imports: map[string]token.Position{}}
	for _, declaration := range program.Declarations {
		switch node := declaration.(type) {
		case *ast.ModuleDeclaration:
			if info.module != "" {
				checker.errorAt(node.Position(), "module is already declared as %q", info.module)
				continue
			}
			info.module = node.Name
		case *ast.ImportDeclaration:
			info.imports[node.Name] = node.Position()
		}
	}
	checker.files[program] = info
}

func (checker *Checker) collectFunctions(program *ast.Program) {
	moduleName := checker.files[program].module
	for _, declaration := range program.Declarations {
		function, ok := declaration.(*ast.FunctionDeclaration)
		if !ok {
			continue
		}

		qualified := qualify(moduleName, function.Name)
		if existing, exists := checker.functions[qualified]; exists && !existing.Builtin {
			checker.errorAt(function.Position(), "function %q is already declared", qualified)
			continue
		}
		if existing, exists := checker.functions[qualified]; exists && existing.Builtin {
			checker.errorAt(function.Position(), "function %q conflicts with a builtin function", function.Name)
			continue
		}

		parameters := make([]ParameterSymbol, 0, len(function.Parameters))
		for _, parameter := range function.Parameters {
			parameterType := checker.resolveType(parameter.TypeName, parameter.Pos)
			parameters = append(parameters, ParameterSymbol{
				Name: parameter.Name,
				Type: parameterType,
				Pos:  parameter.Pos,
			})
		}

		checker.functions[qualified] = FunctionSymbol{
			Name:       function.Name,
			Module:     moduleName,
			Qualified:  qualified,
			Visibility: function.Visibility,
			Turbo:      function.Turbo,
			Parameters: parameters,
			ReturnType: checker.resolveType(function.ReturnType, function.Position()),
			Pos:        function.Position(),
		}
		checker.functionNames[function] = qualified
	}
}

func (checker *Checker) checkFunction(function *ast.FunctionDeclaration, moduleName string) {
	symbol := checker.functions[checker.functionNames[function]]
	checker.currentFunc = &symbol
	checker.currentModule = moduleName
	defer func() {
		checker.currentFunc = nil
		checker.currentModule = ""
	}()

	functionScope := newScope(nil)
	seenParameters := map[string]token.Position{}
	for _, parameter := range symbol.Parameters {
		if previous, exists := seenParameters[parameter.Name]; exists {
			checker.errorAt(parameter.Pos, "parameter %q is already declared at %d:%d", parameter.Name, previous.Line, previous.Column)
			continue
		}
		seenParameters[parameter.Name] = parameter.Pos
		functionScope.declare(variableSymbol{
			name:    parameter.Name,
			typ:     parameter.Type,
			mutable: false,
			pos:     parameter.Pos,
		})
	}

	checker.checkBlock(function.Body, functionScope, false)
	if symbol.ReturnType.Kind != types.Void && !blockHasReturn(function.Body) {
		checker.errorAt(function.Position(), "function %q must return %s", function.Name, symbol.ReturnType)
	}
}

func (checker *Checker) checkBlock(block *ast.BlockStatement, parent *scope, createChild bool) {
	activeScope := parent
	if createChild {
		activeScope = newScope(parent)
	}
	for _, statement := range block.Statements {
		checker.checkStatement(statement, activeScope)
	}
}

func (checker *Checker) checkStatement(statement ast.Statement, activeScope *scope) {
	switch node := statement.(type) {
	case *ast.BlockStatement:
		checker.checkBlock(node, activeScope, true)
	case *ast.VariableDeclaration:
		checker.checkVariableDeclaration(node, activeScope)
	case *ast.AssignmentStatement:
		checker.checkAssignment(node, activeScope)
	case *ast.ReturnStatement:
		checker.checkReturn(node, activeScope)
	case *ast.ExpressionStatement:
		checker.inferExpression(node.Expression, activeScope)
	case *ast.IfStatement:
		checker.checkCondition(node.Condition, activeScope)
		checker.checkBlock(node.ThenBranch, activeScope, true)
		if node.ElseBranch != nil {
			checker.checkBlock(node.ElseBranch, activeScope, true)
		}
	case *ast.WhileStatement:
		checker.checkCondition(node.Condition, activeScope)
		checker.checkBlock(node.Body, activeScope, true)
	default:
		checker.errorAt(statement.Position(), "unsupported statement %T", statement)
	}
}

func (checker *Checker) checkVariableDeclaration(statement *ast.VariableDeclaration, activeScope *scope) {
	valueType := checker.inferExpression(statement.Value, activeScope)
	declaredType := valueType
	if statement.TypeName != "" {
		declaredType = checker.resolveType(statement.TypeName, statement.Position())
		if !types.Compatible(declaredType, valueType) {
			checker.errorAt(statement.Value.Position(), "cannot assign value of type %q to variable of type %q", valueType, declaredType)
		}
	}
	if declaredType.Kind == types.Void {
		checker.errorAt(statement.Position(), "variable %q cannot have type void", statement.Name)
	}
	if previous, exists := activeScope.variables[statement.Name]; exists {
		checker.errorAt(statement.Position(), "variable %q is already declared at %d:%d", statement.Name, previous.pos.Line, previous.pos.Column)
		return
	}
	activeScope.declare(variableSymbol{
		name:    statement.Name,
		typ:     declaredType,
		mutable: statement.Mutable,
		pos:     statement.Position(),
	})
}

func (checker *Checker) checkAssignment(statement *ast.AssignmentStatement, activeScope *scope) {
	variable, ok := activeScope.lookup(statement.Name)
	if !ok {
		checker.errorAt(statement.Position(), "unknown variable %q", statement.Name)
		checker.inferExpression(statement.Value, activeScope)
		return
	}
	if !variable.mutable {
		checker.errorAt(statement.Position(), "cannot assign to const variable %q", statement.Name)
	}
	valueType := checker.inferExpression(statement.Value, activeScope)
	if !types.Compatible(variable.typ, valueType) {
		checker.errorAt(statement.Value.Position(), "cannot assign value of type %q to variable of type %q", valueType, variable.typ)
	}
}

func (checker *Checker) checkReturn(statement *ast.ReturnStatement, activeScope *scope) {
	if checker.currentFunc == nil {
		checker.errorAt(statement.Position(), "return statement outside function")
		return
	}
	expected := checker.currentFunc.ReturnType
	if statement.Value == nil {
		if expected.Kind != types.Void {
			checker.errorAt(statement.Position(), "cannot return void from function returning %q", expected)
		}
		return
	}

	valueType := checker.inferExpression(statement.Value, activeScope)
	if expected.Kind == types.Void {
		checker.errorAt(statement.Position(), "cannot return value from function returning void")
		return
	}
	if !types.Compatible(expected, valueType) {
		checker.errorAt(statement.Value.Position(), "cannot return value of type %q from function returning %q", valueType, expected)
	}
}

func (checker *Checker) checkCondition(expression ast.Expression, activeScope *scope) {
	conditionType := checker.inferExpression(expression, activeScope)
	if conditionType.Kind != types.Invalid && conditionType.Kind != types.Bool {
		checker.errorAt(expression.Position(), "condition must be bool, got %q", conditionType)
	}
}

func (checker *Checker) inferExpression(expression ast.Expression, activeScope *scope) types.Type {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		if node.Name == "<error>" {
			return types.InvalidType
		}
		variable, ok := activeScope.lookup(node.Name)
		if !ok {
			checker.errorAt(node.Position(), "unknown variable %q", node.Name)
			return types.InvalidType
		}
		return variable.typ
	case *ast.LiteralExpression:
		return literalType(node)
	case *ast.UnaryExpression:
		rightType := checker.inferExpression(node.Right, activeScope)
		switch node.Operator {
		case token.Not:
			if rightType.Kind != types.Invalid && rightType.Kind != types.Bool {
				checker.errorAt(node.Position(), "operator ! requires bool, got %q", rightType)
				return types.InvalidType
			}
			return types.BoolType
		case token.Minus:
			if rightType.Kind != types.Invalid && !rightType.IsNumeric() {
				checker.errorAt(node.Position(), "operator - requires numeric operand, got %q", rightType)
				return types.InvalidType
			}
			return rightType
		default:
			checker.errorAt(node.Position(), "unsupported unary operator %s", node.Operator)
			return types.InvalidType
		}
	case *ast.BinaryExpression:
		return checker.inferBinaryExpression(node, activeScope)
	case *ast.CallExpression:
		return checker.inferCallExpression(node, activeScope)
	default:
		checker.errorAt(expression.Position(), "unsupported expression %T", expression)
		return types.InvalidType
	}
}

func (checker *Checker) inferBinaryExpression(expression *ast.BinaryExpression, activeScope *scope) types.Type {
	leftType := checker.inferExpression(expression.Left, activeScope)
	rightType := checker.inferExpression(expression.Right, activeScope)

	switch expression.Operator {
	case token.Plus, token.Minus, token.Star, token.Slash, token.Percent:
		if leftType.Kind == types.Invalid || rightType.Kind == types.Invalid {
			return types.InvalidType
		}
		if !leftType.IsNumeric() || !rightType.IsNumeric() {
			checker.errorAt(expression.Position(), "operator %s requires numeric operands, got %q and %q", expression.Operator, leftType, rightType)
			return types.InvalidType
		}
		if !numericCompatible(leftType, rightType) {
			checker.errorAt(expression.Position(), "operator %s requires matching numeric operands, got %q and %q", expression.Operator, leftType, rightType)
			return types.InvalidType
		}
		if leftType.Kind == types.Nano || rightType.Kind == types.Nano {
			return types.NanoType
		}
		return leftType
	case token.Less, token.LessEqual, token.Greater, token.GreaterEqual:
		if leftType.Kind == types.Invalid || rightType.Kind == types.Invalid {
			return types.InvalidType
		}
		if !leftType.IsNumeric() || !rightType.IsNumeric() {
			checker.errorAt(expression.Position(), "operator %s requires numeric operands, got %q and %q", expression.Operator, leftType, rightType)
			return types.InvalidType
		}
		if !numericCompatible(leftType, rightType) {
			checker.errorAt(expression.Position(), "operator %s requires matching numeric operands, got %q and %q", expression.Operator, leftType, rightType)
		}
		return types.BoolType
	case token.Equal, token.NotEqual:
		if !types.Compatible(leftType, rightType) && !types.Compatible(rightType, leftType) {
			checker.errorAt(expression.Position(), "operator %s cannot compare %q and %q", expression.Operator, leftType, rightType)
		}
		return types.BoolType
	case token.And, token.Or:
		if leftType.Kind != types.Invalid && leftType.Kind != types.Bool {
			checker.errorAt(expression.Left.Position(), "operator %s requires bool operands, got %q", expression.Operator, leftType)
		}
		if rightType.Kind != types.Invalid && rightType.Kind != types.Bool {
			checker.errorAt(expression.Right.Position(), "operator %s requires bool operands, got %q", expression.Operator, rightType)
		}
		return types.BoolType
	default:
		checker.errorAt(expression.Position(), "unsupported binary operator %s", expression.Operator)
		return types.InvalidType
	}
}

func (checker *Checker) inferCallExpression(expression *ast.CallExpression, activeScope *scope) types.Type {
	calleeName, calleePos, ok := checker.resolveCallee(expression.Callee)
	if !ok {
		checker.errorAt(expression.Position(), "function call target must be an identifier or module selector")
		return types.InvalidType
	}

	function, ok := checker.functions[calleeName]
	if !ok {
		checker.errorAt(calleePos, "unknown function %q", calleeName)
		for _, argument := range expression.Arguments {
			checker.inferExpression(argument, activeScope)
		}
		return types.InvalidType
	}
	if function.Module != "" && function.Module != checker.currentModule && function.Visibility != ast.VisibilityPublic {
		checker.errorAt(calleePos, "function %q is private to module %q", function.Name, function.Module)
	}

	if !function.Variadic && len(expression.Arguments) != len(function.Parameters) {
		checker.errorAt(expression.Position(), "function %q expects %d argument(s), got %d", function.Name, len(function.Parameters), len(expression.Arguments))
	}

	for i, argument := range expression.Arguments {
		argumentType := checker.inferExpression(argument, activeScope)
		if i >= len(function.Parameters) {
			continue
		}
		parameterType := function.Parameters[i].Type
		if !types.Compatible(parameterType, argumentType) {
			checker.errorAt(argument.Position(), "cannot pass argument of type %q to parameter %q of type %q", argumentType, function.Parameters[i].Name, parameterType)
		}
	}

	return function.ReturnType
}

func (checker *Checker) resolveCallee(expression ast.Expression) (string, token.Position, bool) {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		if _, ok := checker.functions[node.Name]; ok {
			return node.Name, node.Position(), true
		}
		qualified := qualify(checker.currentModule, node.Name)
		if _, ok := checker.functions[qualified]; ok {
			return qualified, node.Position(), true
		}
		return node.Name, node.Position(), true
	case *ast.SelectorExpression:
		module, ok := node.Left.(*ast.IdentifierExpression)
		if !ok {
			return "", node.Position(), false
		}
		info := checker.fileInfoForCurrentModule()
		if _, imported := info.imports[module.Name]; !imported && module.Name != checker.currentModule {
			checker.errorAt(module.Position(), "module %q is not imported", module.Name)
		}
		return qualify(module.Name, node.Name), node.Position(), true
	default:
		return "", expression.Position(), false
	}
}

func (checker *Checker) fileInfoForCurrentModule() fileInfo {
	for _, info := range checker.files {
		if info.module == checker.currentModule {
			return info
		}
	}
	return fileInfo{imports: map[string]token.Position{}}
}

func (checker *Checker) resolveType(name string, position token.Position) types.Type {
	typ, ok := types.Lookup(name)
	if !ok {
		checker.errorAt(position, "unknown type %q", name)
		return types.InvalidType
	}
	return typ
}

func (checker *Checker) errorAt(position token.Position, format string, args ...any) {
	checker.diagnostics = append(checker.diagnostics, diagnostic.Error(position, fmt.Sprintf(format, args...)))
}

func literalType(literal *ast.LiteralExpression) types.Type {
	switch literal.Kind {
	case ast.LiteralInteger:
		return types.IntType
	case ast.LiteralFloat:
		return types.FloatType
	case ast.LiteralString:
		return types.StringType
	case ast.LiteralBool:
		return types.BoolType
	case ast.LiteralNull:
		return types.NullType
	default:
		return types.InvalidType
	}
}

func numericCompatible(left, right types.Type) bool {
	if left.Kind == right.Kind {
		return true
	}
	return (left.Kind == types.Nano && right.Kind == types.Int) || (left.Kind == types.Int && right.Kind == types.Nano)
}

func blockHasReturn(block *ast.BlockStatement) bool {
	for _, statement := range block.Statements {
		switch node := statement.(type) {
		case *ast.ReturnStatement:
			return true
		case *ast.IfStatement:
			if node.ElseBranch != nil && blockHasReturn(node.ThenBranch) && blockHasReturn(node.ElseBranch) {
				return true
			}
		}
	}
	return false
}

func qualify(moduleName, symbolName string) string {
	if moduleName == "" {
		return symbolName
	}
	return moduleName + "." + symbolName
}

func newScope(parent *scope) *scope {
	return &scope{
		parent:    parent,
		variables: map[string]variableSymbol{},
	}
}

func (activeScope *scope) declare(symbol variableSymbol) {
	activeScope.variables[symbol.name] = symbol
}

func (activeScope *scope) lookup(name string) (variableSymbol, bool) {
	for current := activeScope; current != nil; current = current.parent {
		if symbol, ok := current.variables[name]; ok {
			return symbol, true
		}
	}
	return variableSymbol{}, false
}
