package sema

import (
	"fmt"
	"strings"

	"qwiclang/internal/ast"
	"qwiclang/internal/diagnostic"
	"qwiclang/internal/parser"
	"qwiclang/internal/stdlib"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Diagnostic = diagnostic.Diagnostic

type Result struct {
	Program         *ast.Program
	Programs        []*ast.Program
	Functions       map[string]FunctionSymbol
	FunctionNames   map[*ast.FunctionDeclaration]string
	Types           map[string]TypeSymbol
	TypeNames       map[*ast.TypeDeclaration]string
	ExpressionTypes map[ast.Expression]types.Type
}

type FunctionSymbol struct {
	Name         string
	Module       string
	Qualified    string
	Visibility   ast.Visibility
	Turbo        bool
	Parameters   []ParameterSymbol
	ReturnType   types.Type
	Builtin      bool
	Variadic     bool
	Owner        string
	Receiver     types.Type
	ReceiverName string
	Static       bool
	Pos          token.Position
}

type TypeSymbol struct {
	Name       string
	Module     string
	Qualified  string
	Visibility ast.Visibility
	Fields     map[string]FieldSymbol
	Pos        token.Position
}

type FieldSymbol struct {
	Name string
	Type types.Type
	Pos  token.Position
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
	diagnostics     []Diagnostic
	functions       map[string]FunctionSymbol
	functionNames   map[*ast.FunctionDeclaration]string
	types           map[string]TypeSymbol
	typeNames       map[*ast.TypeDeclaration]string
	expressionTypes map[ast.Expression]types.Type
	files           map[*ast.Program]fileInfo
	moduleOverrides map[*ast.Program]string
	currentFunc     *FunctionSymbol
	currentModule   string
	currentImports  map[string]token.Position
}

type SourceFile struct {
	Filename string
	Source   string
	// Module assigns package identity when a package contains multiple source
	// files that do not repeat a module declaration in every file.
	Module string
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
			Program:         firstProgram(programs),
			Programs:        programs,
			Functions:       map[string]FunctionSymbol{},
			FunctionNames:   map[*ast.FunctionDeclaration]string{},
			Types:           map[string]TypeSymbol{},
			TypeNames:       map[*ast.TypeDeclaration]string{},
			ExpressionTypes: map[ast.Expression]types.Type{},
		}, diagnostics
	}
	checker := NewChecker()
	for index, program := range programs {
		checker.moduleOverrides[program] = files[index].Module
	}
	return checker.CheckPrograms(programs)
}

func firstProgram(programs []*ast.Program) *ast.Program {
	if len(programs) == 0 {
		return &ast.Program{}
	}
	return programs[0]
}

func NewChecker() *Checker {
	checker := &Checker{
		functions:       map[string]FunctionSymbol{},
		functionNames:   map[*ast.FunctionDeclaration]string{},
		types:           map[string]TypeSymbol{},
		typeNames:       map[*ast.TypeDeclaration]string{},
		expressionTypes: map[ast.Expression]types.Type{},
		files:           map[*ast.Program]fileInfo{},
		moduleOverrides: map[*ast.Program]string{},
	}
	checker.functions["print"] = FunctionSymbol{
		Name:       "print",
		Qualified:  "print",
		Parameters: []ParameterSymbol{{Name: "value", Type: types.AnyType}},
		ReturnType: types.VoidType,
		Builtin:    true,
		Pos:        token.Position{Line: 1, Column: 1},
	}
	for _, function := range stdlib.Functions() {
		parameters := make([]ParameterSymbol, 0, len(function.Parameters))
		for _, parameter := range function.Parameters {
			parameters = append(parameters, ParameterSymbol{Name: parameter.Name, Type: parameter.Type})
		}
		qualified := function.Package + "." + function.Name
		checker.functions[qualified] = FunctionSymbol{
			Name:       function.Name,
			Module:     function.Package,
			Qualified:  qualified,
			Visibility: ast.VisibilityPublic,
			Parameters: parameters,
			ReturnType: function.ReturnType,
			Builtin:    true,
			Pos:        token.Position{Line: 1, Column: 1},
		}
	}
	checker.addReceiverFunction("string", "contains", "strings.contains")
	checker.addReceiverFunction("string", "split", "strings.split")
	checker.addReceiverFunction("string", "starts_with", "strings.starts_with")
	checker.addReceiverFunction("string", "ends_with", "strings.ends_with")
	checker.addReceiverFunction("string", "substring", "strings.substring")
	checker.addReceiverFunction("string", "to_int", "strings.to_int")
	checker.addReceiverFunction("string", "to_string", "values.to_string")
	checker.addReceiverFunction("int", "to_string", "values.to_string")
	checker.addReceiverFunction("float", "to_string", "values.to_string")
	checker.addReceiverFunction("nano", "to_string", "values.to_string")
	checker.addReceiverFunction("bool", "to_string", "values.to_string")
	checker.addReceiverFunction("any", "to_string", "values.to_string")
	checker.addReceiverFunction("dictionary", "keys", "dictionaries.keys")
	return checker
}

func (checker *Checker) addReceiverFunction(owner, method, qualified string) {
	function, ok := checker.functions[qualified]
	if !ok || len(function.Parameters) == 0 {
		return
	}
	function.Owner = owner
	function.Receiver = function.Parameters[0].Type
	function.ReceiverName = "value"
	function.Parameters = append([]ParameterSymbol(nil), function.Parameters[1:]...)
	checker.functions[owner+"."+method] = function
}

func (checker *Checker) CheckProgram(program *ast.Program) (*Result, []Diagnostic) {
	return checker.CheckPrograms([]*ast.Program{program})
}

func (checker *Checker) CheckPrograms(programs []*ast.Program) (*Result, []Diagnostic) {
	for _, program := range programs {
		checker.collectFileInfo(program)
	}
	for _, program := range programs {
		checker.collectTypeNames(program)
	}
	for _, program := range programs {
		checker.collectTypeFields(program)
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
			checker.checkFunction(function, checker.files[program])
		}
	}

	return &Result{
		Program:         firstProgram(programs),
		Programs:        programs,
		Functions:       checker.functions,
		FunctionNames:   checker.functionNames,
		Types:           checker.types,
		TypeNames:       checker.typeNames,
		ExpressionTypes: checker.expressionTypes,
	}, checker.diagnostics
}

func (checker *Checker) collectTypeNames(program *ast.Program) {
	moduleName := checker.files[program].module
	for _, declaration := range program.Declarations {
		typeDeclaration, ok := declaration.(*ast.TypeDeclaration)
		if !ok {
			continue
		}
		qualified := qualify(moduleName, typeDeclaration.Name)
		if _, exists := checker.types[qualified]; exists {
			checker.errorAt(typeDeclaration.Position(), "type %q is already declared", qualified)
			continue
		}
		checker.types[qualified] = TypeSymbol{
			Name: typeDeclaration.Name, Module: moduleName, Qualified: qualified,
			Visibility: typeDeclaration.Visibility, Fields: map[string]FieldSymbol{}, Pos: typeDeclaration.Position(),
		}
		checker.typeNames[typeDeclaration] = qualified
	}
}

func (checker *Checker) collectTypeFields(program *ast.Program) {
	moduleName := checker.files[program].module
	for _, declaration := range program.Declarations {
		typeDeclaration, ok := declaration.(*ast.TypeDeclaration)
		if !ok {
			continue
		}
		qualified := checker.typeNames[typeDeclaration]
		symbol := checker.types[qualified]
		for _, field := range typeDeclaration.Fields {
			if existing, exists := symbol.Fields[field.Name]; exists {
				checker.errorAt(field.Pos, "field %q is already declared at %d:%d", field.Name, existing.Pos.Line, existing.Pos.Column)
				continue
			}
			symbol.Fields[field.Name] = FieldSymbol{Name: field.Name, Type: checker.resolveTypeInModule(field.TypeName, moduleName, field.Pos), Pos: field.Pos}
		}
		checker.types[qualified] = symbol
	}
}

func (checker *Checker) collectFileInfo(program *ast.Program) {
	info := fileInfo{
		module:  checker.moduleOverrides[program],
		imports: map[string]token.Position{},
	}
	declaredModule := false
	for _, declaration := range program.Declarations {
		switch node := declaration.(type) {
		case *ast.ModuleDeclaration:
			if declaredModule {
				checker.errorAt(node.Position(), "module is already declared as %q", info.module)
				continue
			}
			declaredModule = true
			if info.module != "" && info.module != node.Name {
				checker.errorAt(node.Position(), "module %q does not match package name %q", node.Name, info.module)
				continue
			}
			if info.module == "" {
				info.module = node.Name
			}
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

		localName := function.Name
		if function.Owner != "" {
			localName = function.Owner + "." + function.Name
		}
		qualified := qualify(moduleName, localName)
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
			parameterType := checker.resolveTypeInModule(parameter.TypeName, moduleName, parameter.Pos)
			parameters = append(parameters, ParameterSymbol{
				Name: parameter.Name,
				Type: parameterType,
				Pos:  parameter.Pos,
			})
		}

		receiver := types.InvalidType
		if function.Owner != "" {
			receiver = checker.resolveTypeInModule(function.Owner, moduleName, function.Position())
			if receiver.Kind != types.Invalid && receiver.Kind != types.Struct {
				checker.errorAt(function.Position(), "method owner %q must be a user-defined type", function.Owner)
			}
		}
		checker.functions[qualified] = FunctionSymbol{
			Name:         function.Name,
			Module:       moduleName,
			Qualified:    qualified,
			Visibility:   function.Visibility,
			Turbo:        function.Turbo,
			Parameters:   parameters,
			ReturnType:   checker.resolveTypeInModule(function.ReturnType, moduleName, function.Position()),
			Owner:        function.Owner,
			Receiver:     receiver,
			ReceiverName: function.Receiver,
			Static:       function.Owner != "" && function.Receiver == "",
			Pos:          function.Position(),
		}
		checker.functionNames[function] = qualified
	}
}

func (checker *Checker) checkFunction(function *ast.FunctionDeclaration, info fileInfo) {
	symbol := checker.functions[checker.functionNames[function]]
	checker.currentFunc = &symbol
	checker.currentModule = info.module
	checker.currentImports = info.imports
	defer func() {
		checker.currentFunc = nil
		checker.currentModule = ""
		checker.currentImports = nil
	}()

	functionScope := newScope(nil)
	if symbol.Owner != "" && !symbol.Static {
		functionScope.declare(variableSymbol{name: symbol.ReceiverName, typ: symbol.Receiver, mutable: false, pos: function.Position()})
	}
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
	case *ast.ForStatement:
		iterableType := checker.inferExpression(node.Iterable, activeScope)
		if iterableType.Kind != types.List && iterableType.Kind != types.Set && iterableType.Kind != types.Dictionary && iterableType.Kind != types.Range {
			checker.errorAt(node.Iterable.Position(), "can only loop over collection types (list, set, dict), got %q", iterableType)
		}

		// The loop variable is implicit and scoped to the body
		// We create a child scope specifically for the for-loop body
		loopScope := &scope{
			parent:    activeScope,
			variables: make(map[string]variableSymbol),
		}
		loopScope.declare(variableSymbol{
			name:    node.Variable,
			typ:     collectionElementType(iterableType),
			mutable: false,
			pos:     node.Pos,
		})
		checker.checkBlock(node.Body, loopScope, true)
	case *ast.TryStatement:
		checker.checkBlock(node.TryBlock, activeScope, true)
		catchScope := newScope(activeScope)
		catchScope.declare(variableSymbol{name: node.CatchVariable, typ: types.StringType, mutable: false, pos: node.Position()})
		checker.checkBlock(node.CatchBlock, catchScope, false)
	case *ast.ThrowStatement:
		valueType := checker.inferExpression(node.Value, activeScope)
		if valueType.Kind != types.Invalid && valueType.Kind != types.String {
			checker.errorAt(node.Value.Position(), "throw value must be string, got %q", valueType)
		}
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
	valueType := checker.inferExpression(statement.Value, activeScope)
	switch target := statement.Target.(type) {
	case *ast.IdentifierExpression:
		variable, ok := activeScope.lookup(target.Name)
		if !ok {
			checker.errorAt(target.Position(), "unknown variable %q", target.Name)
			return
		}
		if !variable.mutable {
			checker.errorAt(target.Position(), "cannot assign to const variable %q", target.Name)
		}
		if !types.Compatible(variable.typ, valueType) {
			checker.errorAt(statement.Value.Position(), "cannot assign value of type %q to variable of type %q", valueType, variable.typ)
		}
	case *ast.SelectorExpression:
		fieldType := checker.inferFieldSelector(target, activeScope)
		if !types.Compatible(fieldType, valueType) {
			checker.errorAt(statement.Value.Position(), "cannot assign value of type %q to field of type %q", valueType, fieldType)
		}
	case *ast.IndexExpression:
		targetType := checker.inferExpression(target, activeScope)
		if !types.Compatible(targetType, valueType) {
			checker.errorAt(statement.Value.Position(), "cannot assign value of type %q to indexed value of type %q", valueType, targetType)
		}
	default:
		checker.errorAt(statement.Target.Position(), "invalid assignment target")
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

func (checker *Checker) inferExpression(expression ast.Expression, activeScope *scope) (result types.Type) {
	defer func() {
		checker.expressionTypes[expression] = result
	}()
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		if node.Name == "<error>" {
			return types.InvalidType
		}
		if variable, ok := activeScope.lookup(node.Name); ok {
			return variable.typ
		}
		functionName := node.Name
		if _, ok := checker.functions[functionName]; !ok {
			functionName = qualify(checker.currentModule, node.Name)
		}
		if function, ok := checker.functions[functionName]; ok {
			return functionType(function)
		}
		checker.errorAt(node.Position(), "unknown variable %q", node.Name)
		return types.InvalidType
	case *ast.SelectorExpression:
		return checker.inferFieldSelector(node, activeScope)
	case *ast.LiteralExpression:
		return literalType(node)
	case *ast.IndexExpression:
		leftType := checker.inferExpression(node.Left, activeScope)
		indexType := checker.inferExpression(node.Index, activeScope)

		if leftType.Kind == types.List || leftType.Kind == types.Tuple {
			if indexType.Kind != types.Int {
				checker.errorAt(node.Index.Position(), "index must be an integer, got %q", indexType)
				return types.InvalidType
			}
			return collectionElementType(leftType)
		}
		if leftType.Kind == types.Dictionary {
			if indexType.Kind != types.String {
				checker.errorAt(node.Index.Position(), "dictionary index must be a string, got %q", indexType)
				return types.InvalidType
			}
			return types.AnyType
		}
		if leftType.Kind == types.Any {
			return types.AnyType
		}
		checker.errorAt(node.Position(), "cannot index into type %q", leftType)
		return types.InvalidType
	case *ast.SliceExpression:
		leftType := checker.inferExpression(node.Left, activeScope)
		startType := checker.inferExpression(node.Start, activeScope)
		endType := checker.inferExpression(node.End, activeScope)

		if leftType.Kind != types.List && leftType.Kind != types.String {
			checker.errorAt(node.Position(), "can only slice lists or strings, got %q", leftType)
			return types.InvalidType
		}
		if startType.Kind != types.Int || endType.Kind != types.Int {
			checker.errorAt(node.Position(), "slice boundaries must be integers")
			return types.InvalidType
		}
		return leftType
	case *ast.ArrayLiteralExpression:
		if len(node.Elements) == 0 {
			return types.ListType
		}
		firstType := checker.inferExpression(node.Elements[0], activeScope)
		allSame := true
		for _, elem := range node.Elements[1:] {
			elemType := checker.inferExpression(elem, activeScope)
			if !types.Compatible(firstType, elemType) {
				allSame = false
			}
		}
		if !allSame {
			return types.AnyType // Or a specialized ListAny type if available
		}
		return types.Type{Kind: types.List, Parameters: []types.Type{firstType}}
	case *ast.DictionaryLiteralExpression:
		for _, pair := range node.Pairs {
			keyType := checker.inferExpression(pair.Key, activeScope)
			if keyType.Kind != types.String && keyType.Kind != types.Any {
				checker.errorAt(pair.Key.Position(), "dictionary keys must be strings, got %q", keyType)
			}
			checker.inferExpression(pair.Value, activeScope)
		}
		return types.DictType
	case *ast.TupleLiteralExpression:
		for _, elem := range node.Elements {
			checker.inferExpression(elem, activeScope)
		}
		return types.TupleType
	case *ast.RangeExpression:
		startType := checker.inferExpression(node.Start, activeScope)
		endType := checker.inferExpression(node.End, activeScope)
		if startType.Kind != types.Int || endType.Kind != types.Int {
			checker.errorAt(node.Position(), "range boundaries must be int, got %q and %q", startType, endType)
		}
		return types.RangeType
	case *ast.TypeLiteralExpression:
		return checker.inferTypeLiteral(node, activeScope)
	case *ast.LambdaExpression:
		return checker.inferLambda(node, activeScope)
	case *ast.InterpolatedStringExpression:
		for _, part := range node.Parts {
			if part.Expression == nil {
				continue
			}
			partType := checker.inferExpression(part.Expression, activeScope)
			if partType.Kind == types.Invalid {
				continue
			}
			if partType.Kind != types.String && partType.Kind != types.Int && partType.Kind != types.Float && partType.Kind != types.Nano && partType.Kind != types.Bool && partType.Kind != types.Any {
				checker.errorAt(part.Expression.Position(), "cannot format value of type %q in f-string", partType)
			}
		}
		return types.StringType
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

func (checker *Checker) inferFieldSelector(expression *ast.SelectorExpression, activeScope *scope) types.Type {
	leftType := checker.inferExpression(expression.Left, activeScope)
	if expression.Name == "length" {
		switch leftType.Kind {
		case types.String, types.List, types.Dictionary, types.Tuple, types.Set:
			return types.IntType
		}
	}
	if leftType.Kind == types.Any {
		if expression.Name == "type" {
			return types.StringType
		}
		return types.AnyType
	}
	if leftType.Kind != types.Struct {
		if leftType.Kind != types.Invalid {
			checker.errorAt(expression.Position(), "cannot select field %q on type %q", expression.Name, leftType)
		}
		return types.InvalidType
	}
	typeSymbol, ok := checker.types[leftType.Name]
	if !ok {
		checker.errorAt(expression.Position(), "unknown type %q", leftType.Name)
		return types.InvalidType
	}
	field, ok := typeSymbol.Fields[expression.Name]
	if !ok {
		checker.errorAt(expression.Position(), "type %q has no field %q", typeSymbol.Name, expression.Name)
		return types.InvalidType
	}
	return field.Type
}

func (checker *Checker) inferTypeLiteral(expression *ast.TypeLiteralExpression, activeScope *scope) types.Type {
	typeName, position, ok := checker.resolveTypeReference(expression.Type)
	if !ok {
		checker.errorAt(expression.Position(), "invalid type literal target")
		return types.InvalidType
	}
	symbol, ok := checker.types[typeName]
	if !ok {
		checker.errorAt(position, "unknown type %q", typeName)
		return types.InvalidType
	}
	if symbol.Module != "" && symbol.Module != checker.currentModule && symbol.Visibility != ast.VisibilityPublic {
		checker.errorAt(position, "type %q is private to module %q", symbol.Name, symbol.Module)
	}
	seen := map[string]bool{}
	for _, value := range expression.Fields {
		field, exists := symbol.Fields[value.Name]
		if !exists {
			checker.errorAt(value.Pos, "type %q has no field %q", symbol.Name, value.Name)
			checker.inferExpression(value.Value, activeScope)
			continue
		}
		if seen[value.Name] {
			checker.errorAt(value.Pos, "field %q is initialized more than once", value.Name)
		}
		seen[value.Name] = true
		valueType := checker.inferExpression(value.Value, activeScope)
		if !types.Compatible(field.Type, valueType) {
			checker.errorAt(value.Value.Position(), "cannot initialize field %q of type %q with %q", value.Name, field.Type, valueType)
		}
	}
	for name := range symbol.Fields {
		if !seen[name] {
			checker.errorAt(expression.Position(), "missing value for field %q of type %q", name, symbol.Name)
		}
	}
	return types.StructType(symbol.Qualified)
}

func (checker *Checker) inferLambda(expression *ast.LambdaExpression, activeScope *scope) types.Type {
	parameterTypes := make([]types.Type, 0, len(expression.Parameters))
	lambdaScope := newScope(activeScope)
	for _, parameter := range expression.Parameters {
		parameterType := checker.resolveType(parameter.TypeName, parameter.Pos)
		parameterTypes = append(parameterTypes, parameterType)
		if _, exists := lambdaScope.variables[parameter.Name]; exists {
			checker.errorAt(parameter.Pos, "lambda parameter %q is already declared", parameter.Name)
			continue
		}
		lambdaScope.declare(variableSymbol{name: parameter.Name, typ: parameterType, mutable: false, pos: parameter.Pos})
	}
	returnType := checker.resolveType(expression.ReturnType, expression.Position())
	previousFunction := checker.currentFunc
	checker.currentFunc = &FunctionSymbol{Name: "<lambda>", ReturnType: returnType}
	checker.checkBlock(expression.Body, lambdaScope, false)
	checker.currentFunc = previousFunction
	if returnType.Kind != types.Void && !blockHasReturn(expression.Body) {
		checker.errorAt(expression.Position(), "lambda must return %s", returnType)
	}
	return types.FunctionType(parameterTypes, returnType)
}

func (checker *Checker) inferBinaryExpression(expression *ast.BinaryExpression, activeScope *scope) types.Type {
	leftType := checker.inferExpression(expression.Left, activeScope)
	rightType := checker.inferExpression(expression.Right, activeScope)

	switch expression.Operator {
	case token.Plus, token.Minus, token.Star, token.Slash, token.Percent:
		if leftType.Kind == types.Invalid || rightType.Kind == types.Invalid {
			return types.InvalidType
		}
		if expression.Operator == token.Plus && leftType.Kind == types.List && rightType.Kind == types.List {
			return types.ListType
		}
		if expression.Operator == token.Plus && leftType.Kind == types.String && rightType.Kind == types.String {
			return types.StringType
		}
		if expression.Operator == token.Plus && (leftType.Kind == types.String || rightType.Kind == types.String) && (leftType.Kind == types.Any || rightType.Kind == types.Any) {
			return types.StringType
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
	function, _, calleePos, ok := checker.resolveCallee(expression.Callee, activeScope)
	if !ok {
		if identifier, isIdentifier := expression.Callee.(*ast.IdentifierExpression); isIdentifier {
			if _, isVariable := activeScope.lookup(identifier.Name); !isVariable {
				checker.errorAt(calleePos, "unknown function %q", identifier.Name)
				for _, argument := range expression.Arguments {
					checker.inferExpression(argument, activeScope)
				}
				return types.InvalidType
			}
		}
		calleeType := checker.inferExpression(expression.Callee, activeScope)
		if calleeType.Kind == types.Any {
			for _, argument := range expression.Arguments {
				checker.inferExpression(argument, activeScope)
			}
			return types.AnyType
		}
		if calleeType.Kind == types.Function {
			checker.checkCallArguments("lambda", calleeType.Parameters, expression.Arguments, activeScope, expression.Position())
			if calleeType.ReturnType == nil {
				return types.VoidType
			}
			return *calleeType.ReturnType
		}
		for _, argument := range expression.Arguments {
			checker.inferExpression(argument, activeScope)
		}
		checker.errorAt(calleePos, "expression of type %q is not callable", calleeType)
		return types.InvalidType
	}
	if function.Module != "" && function.Module != checker.currentModule && function.Visibility != ast.VisibilityPublic {
		checker.errorAt(calleePos, "function %q is private to module %q", function.Name, function.Module)
	}
	parameterTypes := make([]types.Type, 0, len(function.Parameters))
	for _, parameter := range function.Parameters {
		parameterTypes = append(parameterTypes, parameter.Type)
	}
	if !function.Variadic && len(expression.Arguments) != len(parameterTypes) {
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

func (checker *Checker) checkCallArguments(name string, parameters []types.Type, arguments []ast.Expression, activeScope *scope, position token.Position) {
	if len(arguments) != len(parameters) {
		checker.errorAt(position, "%s expects %d argument(s), got %d", name, len(parameters), len(arguments))
	}
	for index, argument := range arguments {
		argumentType := checker.inferExpression(argument, activeScope)
		if index < len(parameters) && !types.Compatible(parameters[index], argumentType) {
			checker.errorAt(argument.Position(), "cannot pass argument of type %q to parameter of type %q", argumentType, parameters[index])
		}
	}
}

func (checker *Checker) resolveCallee(expression ast.Expression, activeScope *scope) (FunctionSymbol, ast.Expression, token.Position, bool) {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		if function, ok := checker.functions[node.Name]; ok {
			return function, nil, node.Position(), true
		}
		qualified := qualify(checker.currentModule, node.Name)
		if function, ok := checker.functions[qualified]; ok {
			return function, nil, node.Position(), true
		}
		return FunctionSymbol{}, nil, node.Position(), false
	case *ast.SelectorExpression:
		parts, pathOK := selectorPath(node)
		if pathOK {
			candidates := []string{strings.Join(parts, ".")}
			if checker.currentModule != "" {
				candidates = append(candidates, qualify(checker.currentModule, strings.Join(parts, ".")))
			}
			for _, candidate := range candidates {
				if function, exists := checker.functions[candidate]; exists {
					if function.Module != "" && function.Module != checker.currentModule {
						if _, imported := checker.currentImports[function.Module]; !imported {
							checker.errorAt(node.Position(), "module %q is not imported", function.Module)
						}
					}
					return function, nil, node.Position(), true
				}
			}
		}

		leftType := checker.inferExpression(node.Left, activeScope)
		candidate := receiverOwner(leftType) + "." + node.Name
		if function, exists := checker.functions[candidate]; exists && !function.Static {
			return function, node.Left, node.Position(), true
		}
		return FunctionSymbol{}, nil, node.Position(), false
	default:
		return FunctionSymbol{}, nil, expression.Position(), false
	}
}

func receiverOwner(typ types.Type) string {
	if typ.Kind == types.Struct {
		return typ.Name
	}
	return typ.String()
}

func collectionElementType(typ types.Type) types.Type {
	if len(typ.Parameters) > 0 {
		return typ.Parameters[0]
	}
	if typ.Kind == types.Range {
		return types.IntType
	}
	return types.AnyType
}

func selectorPath(expression ast.Expression) ([]string, bool) {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		return []string{node.Name}, true
	case *ast.SelectorExpression:
		parts, ok := selectorPath(node.Left)
		if !ok {
			return nil, false
		}
		return append(parts, node.Name), true
	default:
		return nil, false
	}
}

func (checker *Checker) resolveType(name string, position token.Position) types.Type {
	typ, ok := types.Lookup(name)
	if ok {
		return typ
	}
	if symbol, exists := checker.types[name]; exists {
		if symbol.Module != "" && symbol.Module != checker.currentModule {
			if _, imported := checker.currentImports[symbol.Module]; !imported {
				checker.errorAt(position, "module %q is not imported", symbol.Module)
			}
			if symbol.Visibility != ast.VisibilityPublic {
				checker.errorAt(position, "type %q is private to module %q", symbol.Name, symbol.Module)
			}
		}
		return types.StructType(name)
	}
	qualified := qualify(checker.currentModule, name)
	if _, exists := checker.types[qualified]; exists {
		return types.StructType(qualified)
	}
	checker.errorAt(position, "unknown type %q", name)
	return types.InvalidType
}

func (checker *Checker) resolveTypeInModule(name, module string, position token.Position) types.Type {
	if typ, ok := types.Lookup(name); ok {
		return typ
	}
	if _, exists := checker.types[name]; exists {
		return types.StructType(name)
	}
	qualified := qualify(module, name)
	if _, exists := checker.types[qualified]; exists {
		return types.StructType(qualified)
	}
	checker.errorAt(position, "unknown type %q", name)
	return types.InvalidType
}

func (checker *Checker) resolveTypeReference(expression ast.Expression) (string, token.Position, bool) {
	parts, ok := selectorPath(expression)
	if !ok || len(parts) == 0 {
		return "", expression.Position(), false
	}
	if len(parts) == 1 {
		return qualify(checker.currentModule, parts[0]), expression.Position(), true
	}
	moduleName := parts[0]
	if moduleName != checker.currentModule {
		if _, imported := checker.currentImports[moduleName]; !imported {
			checker.errorAt(expression.Position(), "module %q is not imported", moduleName)
		}
	}
	return strings.Join(parts, "."), expression.Position(), true
}

func functionType(function FunctionSymbol) types.Type {
	parameters := make([]types.Type, 0, len(function.Parameters))
	for _, parameter := range function.Parameters {
		parameters = append(parameters, parameter.Type)
	}
	return types.FunctionType(parameters, function.ReturnType)
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
		case *ast.ThrowStatement:
			return true
		case *ast.IfStatement:
			if node.ElseBranch != nil && blockHasReturn(node.ThenBranch) && blockHasReturn(node.ElseBranch) {
				return true
			}
		case *ast.TryStatement:
			if blockHasReturn(node.TryBlock) && blockHasReturn(node.CatchBlock) {
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
