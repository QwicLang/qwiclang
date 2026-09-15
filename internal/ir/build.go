package ir

import (
	"fmt"
	"sort"
	"strings"

	"qwiclang/internal/ast"
	"qwiclang/internal/diagnostic"
	"qwiclang/internal/sema"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Diagnostic = diagnostic.Diagnostic

type Builder struct {
	functions          map[string]sema.FunctionSymbol
	names              map[*ast.FunctionDeclaration]string
	types              map[string]sema.TypeSymbol
	typeNames          map[*ast.TypeDeclaration]string
	expressionTypes    map[ast.Expression]types.Type
	diagnostics        []Diagnostic
	function           *Function
	moduleName         string
	blockIndex         int
	tempIndex          int
	lambdaIndex        int
	scopes             []map[string]types.Type
	activeTryFrames    []string
	generatedFunctions []Function
}

func Build(result *sema.Result) (Module, []Diagnostic) {
	return NewBuilder(result.Functions, result.FunctionNames, result.Types, result.TypeNames, result.ExpressionTypes).BuildPrograms(result.Programs)
}

func BuildSource(filename, source string) (Module, []Diagnostic) {
	result, diagnostics := sema.Check(filename, source)
	if len(diagnostics) > 0 {
		return Module{}, diagnostics
	}
	return Build(result)
}

func NewBuilder(
	functions map[string]sema.FunctionSymbol,
	names map[*ast.FunctionDeclaration]string,
	typeSymbols map[string]sema.TypeSymbol,
	typeNames map[*ast.TypeDeclaration]string,
	expressionTypes map[ast.Expression]types.Type,
) *Builder {
	return &Builder{
		functions:       functions,
		names:           names,
		types:           typeSymbols,
		typeNames:       typeNames,
		expressionTypes: expressionTypes,
	}
}

func (builder *Builder) BuildProgram(program *ast.Program) (Module, []Diagnostic) {
	return builder.BuildPrograms([]*ast.Program{program})
}

func (builder *Builder) BuildPrograms(programs []*ast.Program) (Module, []Diagnostic) {
	module := Module{}
	builder.generatedFunctions = nil
	for _, program := range programs {
		for _, declaration := range program.Declarations {
			typeDeclaration, ok := declaration.(*ast.TypeDeclaration)
			if !ok {
				continue
			}
			qualifiedName := builder.typeNames[typeDeclaration]
			symbol := builder.types[qualifiedName]
			definition := TypeDefinition{Name: qualifiedName}
			for _, field := range typeDeclaration.Fields {
				definition.Fields = append(definition.Fields, Field{Name: field.Name, Type: symbol.Fields[field.Name].Type})
			}
			module.Types = append(module.Types, definition)
		}
	}
	for _, program := range programs {
		moduleName := moduleName(program)
		for _, declaration := range program.Declarations {
			function, ok := declaration.(*ast.FunctionDeclaration)
			if !ok {
				continue
			}
			module.Functions = append(module.Functions, builder.buildFunction(function, moduleName))
		}
	}
	module.Functions = append(module.Functions, builder.generatedFunctions...)
	return module, builder.diagnostics
}

func (builder *Builder) buildFunction(declaration *ast.FunctionDeclaration, moduleName string) Function {
	qualifiedName := builder.names[declaration]
	if qualifiedName == "" {
		qualifiedName = qualify(moduleName, declaration.Name)
	}
	symbol := builder.functions[qualifiedName]
	if symbol.Module != "" {
		moduleName = symbol.Module
	}
	function := Function{
		Name:       qualifiedName,
		Visibility: declaration.Visibility,
		Turbo:      declaration.Turbo,
		ReturnType: symbol.ReturnType,
	}
	if symbol.Owner != "" && !symbol.Static {
		function.Parameters = append(function.Parameters, Parameter{Name: symbol.ReceiverName, Type: symbol.Receiver})
	}
	for _, parameter := range symbol.Parameters {
		function.Parameters = append(function.Parameters, Parameter{
			Name: parameter.Name,
			Type: parameter.Type,
		})
	}

	previousFunction := builder.function
	previousModuleName := builder.moduleName
	builder.function = &function
	builder.moduleName = moduleName
	builder.blockIndex = 0
	builder.tempIndex = 0
	builder.scopes = nil
	builder.activeTryFrames = nil
	builder.pushScope()
	if symbol.Owner != "" && !symbol.Static {
		builder.declare(symbol.ReceiverName, symbol.Receiver)
	}
	for _, parameter := range symbol.Parameters {
		builder.declare(parameter.Name, parameter.Type)
	}

	builder.appendBlock("entry")
	builder.buildBlock(declaration.Body)
	if len(builder.currentBlock().Instructions) == 0 || !isTerminator(builder.currentBlock().Instructions[len(builder.currentBlock().Instructions)-1]) {
		if symbol.ReturnType.Kind == types.Void {
			builder.emit(&Return{})
		}
	}

	function = *builder.function
	builder.popScope()
	builder.function = previousFunction
	builder.moduleName = previousModuleName
	return function
}

func (builder *Builder) buildBlock(block *ast.BlockStatement) {
	builder.pushScope()
	defer builder.popScope()

	for _, statement := range block.Statements {
		builder.buildStatement(statement)
	}
}

func (builder *Builder) buildStatement(statement ast.Statement) {
	switch node := statement.(type) {
	case *ast.VariableDeclaration:
		value := builder.buildExpression(node.Value)
		variableType := value.Type
		if node.TypeName != "" {
			variableType = builder.lookupType(node.TypeName, node.Position())
		}
		builder.emit(&Variable{Name: node.Name, Type: variableType, Mutable: node.Mutable})
		builder.emit(&Store{Target: node.Name, Value: value.Name})
		builder.declare(node.Name, variableType)
	case *ast.AssignmentStatement:
		value := builder.buildExpression(node.Value)
		builder.buildAssignment(node.Target, value)
	case *ast.ReturnStatement:
		builder.endActiveTryFrames()
		if node.Value == nil {
			builder.emit(&Return{})
			return
		}
		value := builder.buildExpression(node.Value)
		builder.emit(&Return{Value: value.Name})
	case *ast.ExpressionStatement:
		builder.buildExpression(node.Expression)
	case *ast.IfStatement:
		builder.buildIfStatement(node)
	case *ast.WhileStatement:
		builder.buildWhileStatement(node)
	case *ast.ForStatement:
		builder.buildForStatement(node)
	case *ast.TryStatement:
		builder.buildTryStatement(node)
	case *ast.ThrowStatement:
		value := builder.buildExpression(node.Value)
		builder.emit(&Throw{Value: value.Name})
	default:
		builder.errorAt(statement.Position(), "unsupported statement %T", statement)
	}
}

func (builder *Builder) buildAssignment(target ast.Expression, value valueRef) {
	switch node := target.(type) {
	case *ast.IdentifierExpression:
		builder.emit(&Store{Target: node.Name, Value: value.Name})
	case *ast.SelectorExpression:
		object := builder.buildExpression(node.Left)
		builder.emit(&StoreField{Object: object.Name, Field: node.Name, Value: value.Name})
	case *ast.IndexExpression:
		object := builder.buildExpression(node.Left)
		index := builder.buildExpression(node.Index)
		if object.Type.Kind == types.Any {
			builder.emit(&Call{Function: "values.set_index", Args: []string{object.Name, index.Name, value.Name}, Type: types.VoidType})
			return
		}
		if object.Type.Kind != types.Dictionary {
			builder.errorAt(node.Position(), "index assignment is only supported for dictionaries")
			return
		}
		builder.emit(&Call{Function: "dictionaries.set", Args: []string{object.Name, index.Name, value.Name}, Type: types.VoidType})
	default:
		builder.errorAt(target.Position(), "invalid assignment target")
	}
}

func (builder *Builder) buildTryStatement(statement *ast.TryStatement) {
	frame := fmt.Sprintf("try_frame_%d", builder.blockIndex+1)
	tryBlock := builder.newBlockName("try.body")
	catchBlock := builder.newBlockName("try.catch")
	afterBlock := builder.newBlockName("try.end")
	outerFrames := append([]string(nil), builder.activeTryFrames...)

	builder.emit(&TryBegin{Frame: frame, TryBlock: tryBlock, CatchBlock: catchBlock})
	builder.appendBlock(tryBlock)
	builder.activeTryFrames = append(outerFrames, frame)
	builder.buildBlock(statement.TryBlock)
	if !builder.currentBlockTerminated() {
		builder.emit(&TryEnd{Frame: frame})
		builder.emit(&Jump{Target: afterBlock})
	}

	builder.appendBlock(catchBlock)
	builder.activeTryFrames = outerFrames
	builder.pushScope()
	builder.emit(&Catch{Target: statement.CatchVariable, Frame: frame})
	builder.declare(statement.CatchVariable, types.StringType)
	for _, child := range statement.CatchBlock.Statements {
		builder.buildStatement(child)
	}
	builder.popScope()
	builder.emitJumpIfNeeded(afterBlock)
	builder.appendBlock(afterBlock)
	builder.activeTryFrames = outerFrames
}

func (builder *Builder) endActiveTryFrames() {
	for index := len(builder.activeTryFrames) - 1; index >= 0; index-- {
		builder.emit(&TryEnd{Frame: builder.activeTryFrames[index]})
	}
}

func (builder *Builder) buildForStatement(node *ast.ForStatement) {
	if _, ok := node.Iterable.(*ast.RangeExpression); ok {
		builder.buildRangeForStatement(node)
		return
	}
	iterable := builder.buildExpression(node.Iterable)

	// We desugar 'for var in iterable' into:
	// let i = 0
	// while i < lists.length(iterable) {
	//     var = lists.get(iterable, i)
	//     ... body ...
	//     i = i + 1
	// }

	// 1. Use unique names for the index to avoid collisions.
	indexVar := fmt.Sprintf("for_idx_%d", builder.tempIndex+1)
	builder.tempIndex++

	// Explicitly declare index and loop variable
	builder.emit(&Variable{Name: indexVar, Type: types.IntType, Mutable: true})
	elementType := collectionElementType(iterable.Type)
	builder.emit(&Variable{Name: node.Variable, Type: elementType, Mutable: true})

	condBlock := builder.newBlockName("for.cond")
	bodyBlock := builder.newBlockName("for.body")
	endBlock := builder.newBlockName("for.end")

	// Condition: index < lists.length(iterable)
	builder.emit(&Jump{Target: condBlock})
	builder.appendBlock(condBlock)

	lenCall := builder.newTemp()
	builder.emit(&Call{Target: lenCall, Function: "lists.length", Args: []string{iterable.Name}, Type: types.IntType})

	condTemp := builder.newTemp()
	builder.emit(&BinaryOperation{Target: condTemp, Operator: token.Less, Left: indexVar, Right: lenCall, Type: types.BoolType})
	builder.emit(&Branch{Condition: condTemp, ThenBlock: bodyBlock, ElseBlock: endBlock})

	builder.appendBlock(bodyBlock)

	// Get current item: var = lists.get(iterable, index)
	itemTemp := builder.newTemp()
	builder.emit(&Call{Target: itemTemp, Function: "lists.get", Args: []string{iterable.Name, indexVar}, Type: elementType})
	builder.emit(&Store{Target: node.Variable, Value: itemTemp})

	// Build body with loop variable declared in scope
	builder.pushScope()
	builder.declare(node.Variable, elementType)
	for _, statement := range node.Body.Statements {
		builder.buildStatement(statement)
	}
	builder.popScope()

	// Increment index: i = i + 1
	oneTemp := builder.newTemp()
	builder.emit(&Constant{Target: oneTemp, Type: types.IntType, Value: "1"})

	incTemp := builder.newTemp()
	builder.emit(&BinaryOperation{Target: incTemp, Operator: token.Plus, Left: indexVar, Right: oneTemp, Type: types.IntType})
	builder.emit(&Store{Target: indexVar, Value: incTemp})

	builder.emitJumpIfNeeded(condBlock)
	builder.appendBlock(endBlock)
}

func (builder *Builder) buildRangeForStatement(node *ast.ForStatement) {
	rangeExpression := node.Iterable.(*ast.RangeExpression)
	start := builder.buildExpression(rangeExpression.Start)
	end := builder.buildExpression(rangeExpression.End)
	builder.emit(&Variable{Name: node.Variable, Type: types.IntType, Mutable: true})
	builder.emit(&Store{Target: node.Variable, Value: start.Name})

	conditionBlock := builder.newBlockName("range.cond")
	bodyBlock := builder.newBlockName("range.body")
	afterBlock := builder.newBlockName("range.end")
	builder.emit(&Jump{Target: conditionBlock})
	builder.appendBlock(conditionBlock)
	condition := builder.newTemp()
	builder.emit(&BinaryOperation{Target: condition, Operator: token.LessEqual, Left: node.Variable, Right: end.Name, Type: types.BoolType})
	builder.emit(&Branch{Condition: condition, ThenBlock: bodyBlock, ElseBlock: afterBlock})

	builder.appendBlock(bodyBlock)
	builder.pushScope()
	builder.declare(node.Variable, types.IntType)
	for _, statement := range node.Body.Statements {
		builder.buildStatement(statement)
	}
	builder.popScope()
	one := builder.newTemp()
	builder.emit(&Constant{Target: one, Type: types.IntType, Value: "1"})
	next := builder.newTemp()
	builder.emit(&BinaryOperation{Target: next, Operator: token.Plus, Left: node.Variable, Right: one, Type: types.IntType})
	builder.emit(&Store{Target: node.Variable, Value: next})
	builder.emitJumpIfNeeded(conditionBlock)
	builder.appendBlock(afterBlock)
}

func (builder *Builder) buildIfStatement(statement *ast.IfStatement) {
	condition := builder.buildExpression(statement.Condition)
	thenBlock := builder.newBlockName("if.then")
	elseBlock := builder.newBlockName("if.else")
	afterBlock := builder.newBlockName("if.end")
	if statement.ElseBranch == nil {
		elseBlock = afterBlock
	}

	builder.emit(&Branch{Condition: condition.Name, ThenBlock: thenBlock, ElseBlock: elseBlock})
	builder.appendBlock(thenBlock)
	builder.buildBlock(statement.ThenBranch)
	builder.emitJumpIfNeeded(afterBlock)

	if statement.ElseBranch != nil {
		builder.appendBlock(elseBlock)
		builder.buildBlock(statement.ElseBranch)
		builder.emitJumpIfNeeded(afterBlock)
	}

	builder.appendBlock(afterBlock)
}

func (builder *Builder) buildWhileStatement(statement *ast.WhileStatement) {
	conditionBlock := builder.newBlockName("while.cond")
	bodyBlock := builder.newBlockName("while.body")
	afterBlock := builder.newBlockName("while.end")

	builder.emit(&Jump{Target: conditionBlock})
	builder.appendBlock(conditionBlock)
	condition := builder.buildExpression(statement.Condition)
	builder.emit(&Branch{Condition: condition.Name, ThenBlock: bodyBlock, ElseBlock: afterBlock})

	builder.appendBlock(bodyBlock)
	builder.buildBlock(statement.Body)
	builder.emitJumpIfNeeded(conditionBlock)

	builder.appendBlock(afterBlock)
}

type valueRef struct {
	Name string
	Type types.Type
}

func (builder *Builder) buildExpression(expression ast.Expression) valueRef {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		variableType := builder.lookupVariable(node.Name)
		if variableType.Kind == types.Invalid {
			if functionName, function, ok := builder.lookupFunction(node.Name); ok {
				target := builder.newTemp()
				functionType := functionType(function)
				builder.emit(&FunctionReference{Target: target, Function: functionName, Type: functionType})
				return valueRef{Name: target, Type: functionType}
			}
		}
		target := builder.newTemp()
		builder.emit(&Load{Target: target, Source: node.Name, Type: variableType})
		return valueRef{Name: target, Type: variableType}
	case *ast.SelectorExpression:
		object := builder.buildExpression(node.Left)
		resultType := builder.expressionType(expression)
		target := builder.newTemp()
		if function := propertyFunction(object.Type, node.Name); function != "" {
			builder.emit(&Call{Target: target, Function: function, Args: []string{object.Name}, Type: resultType})
			return valueRef{Name: target, Type: resultType}
		}
		builder.emit(&LoadField{Target: target, Object: object.Name, Field: node.Name, Type: resultType})
		return valueRef{Name: target, Type: resultType}
	case *ast.LiteralExpression:
		target := builder.newTemp()
		literalType := builder.literalType(node)
		builder.emit(&Constant{Target: target, Type: literalType, Value: node.Value})
		return valueRef{Name: target, Type: literalType}
	case *ast.IndexExpression:
		left := builder.buildExpression(node.Left)
		index := builder.buildExpression(node.Index)
		target := builder.newTemp()

		var funcName string
		switch left.Type.Kind {
		case types.List:
			funcName = "lists.get"
		case types.Dictionary:
			funcName = "dictionaries.get"
		case types.Any:
			funcName = "values.index"
		case types.Tuple:
			// For tuples, we map index 0 to first, 1 to second, etc.
			if lit, ok := node.Index.(*ast.LiteralExpression); ok && lit.Kind == ast.LiteralInteger {
				switch lit.Value {
				case "0":
					funcName = "tuples.first"
				case "1":
					funcName = "tuples.second"
				default:
					builder.errorAt(node.Position(), "tuple index out of range (only 0 and 1 supported in v0)")
					funcName = "tuples.first"
				}
			} else {
				builder.errorAt(node.Position(), "dynamic indexing of tuples not yet supported")
				funcName = "tuples.first"
			}
			builder.emit(&Call{Target: target, Function: funcName, Args: []string{left.Name}, Type: types.StringType})
			return valueRef{Name: target, Type: types.StringType}
		default:
			builder.errorAt(node.Position(), "cannot index into type %q", left.Type)
			funcName = "lists.get"
		}

		resultType := builder.expressionType(expression)
		builder.emit(&Call{Target: target, Function: funcName, Args: []string{left.Name, index.Name}, Type: resultType})
		return valueRef{Name: target, Type: resultType}
	case *ast.SliceExpression:
		left := builder.buildExpression(node.Left)
		start := builder.buildExpression(node.Start)
		end := builder.buildExpression(node.End)
		target := builder.newTemp()

		var funcName string
		if left.Type.Kind == types.List {
			funcName = "lists.slice"
		} else if left.Type.Kind == types.String {
			funcName = "strings.slice"
		} else {
			builder.errorAt(node.Position(), "unsupported slice type %q", left.Type)
			funcName = "lists.slice"
		}

		builder.emit(&Call{Target: target, Function: funcName, Args: []string{left.Name, start.Name, end.Name}, Type: left.Type})
		return valueRef{Name: target, Type: left.Type}
	case *ast.ArrayLiteralExpression:
		target := builder.newTemp()
		listType := builder.expressionType(expression)
		builder.emit(&Call{Target: target, Function: "lists.new", Args: []string{}, Type: listType})
		for _, elem := range node.Elements {
			val := builder.buildExpression(elem)
			builder.emit(&Call{Target: "", Function: "lists.push", Args: []string{target, val.Name}, Type: types.VoidType})
		}
		return valueRef{Name: target, Type: listType}
	case *ast.DictionaryLiteralExpression:
		target := builder.newTemp()
		builder.emit(&Call{Target: target, Function: "dictionaries.new", Args: []string{}, Type: types.DictType})
		for _, pair := range node.Pairs {
			key := builder.buildExpression(pair.Key)
			val := builder.buildExpression(pair.Value)
			builder.emit(&Call{Target: "", Function: "dictionaries.set", Args: []string{target, key.Name, val.Name}, Type: types.VoidType})
		}
		return valueRef{Name: target, Type: types.DictType}
	case *ast.TupleLiteralExpression:
		// Simple desugaring to tuples.newN based on number of elements
		count := len(node.Elements)
		args := make([]string, 0, count)
		for _, elem := range node.Elements {
			args = append(args, builder.buildExpression(elem).Name)
		}
		target := builder.newTemp()
		funcName := fmt.Sprintf("tuples.new%d", count)
		builder.emit(&Call{Target: target, Function: funcName, Args: args, Type: types.TupleType})
		return valueRef{Name: target, Type: types.TupleType}
	case *ast.RangeExpression:
		builder.errorAt(node.Position(), "range expressions are only supported by for loops")
		return valueRef{Name: "<range>", Type: types.RangeType}
	case *ast.TypeLiteralExpression:
		resultType := builder.expressionType(expression)
		target := builder.newTemp()
		builder.emit(&NewStruct{Target: target, Type: resultType})
		for _, field := range node.Fields {
			value := builder.buildExpression(field.Value)
			builder.emit(&StoreField{Object: target, Field: field.Name, Value: value.Name})
		}
		return valueRef{Name: target, Type: resultType}
	case *ast.LambdaExpression:
		functionName, functionType, captures := builder.buildLambda(node)
		target := builder.newTemp()
		builder.emit(&FunctionReference{Target: target, Function: functionName, Type: functionType, Captures: captures})
		return valueRef{Name: target, Type: functionType}
	case *ast.InterpolatedStringExpression:
		target := builder.newTemp()
		parts := make([]FormatPart, 0, len(node.Parts))
		for _, part := range node.Parts {
			if part.Expression == nil {
				parts = append(parts, FormatPart{Text: part.Text})
				continue
			}
			value := builder.buildExpression(part.Expression)
			parts = append(parts, FormatPart{Value: value.Name, Type: value.Type})
		}
		builder.emit(&FormatString{Target: target, Parts: parts, Type: types.StringType})
		return valueRef{Name: target, Type: types.StringType}
	case *ast.UnaryExpression:
		right := builder.buildExpression(node.Right)
		zero := valueRef{}
		if node.Operator == token.Minus {
			zero = builder.emitZero(right.Type)
			target := builder.newTemp()
			builder.emit(&BinaryOperation{Target: target, Operator: token.Minus, Left: zero.Name, Right: right.Name, Type: right.Type})
			return valueRef{Name: target, Type: right.Type}
		}
		target := builder.newTemp()
		builder.emit(&BinaryOperation{Target: target, Operator: node.Operator, Left: right.Name, Type: types.BoolType})
		return valueRef{Name: target, Type: types.BoolType}
	case *ast.BinaryExpression:
		left := builder.buildExpression(node.Left)
		right := builder.buildExpression(node.Right)
		resultType := left.Type
		switch node.Operator {
		case token.Equal, token.NotEqual, token.Less, token.LessEqual, token.Greater, token.GreaterEqual, token.And, token.Or:
			resultType = types.BoolType
		case token.Plus, token.Minus, token.Star, token.Slash, token.Percent:
			if left.Type.Kind == types.Nano || right.Type.Kind == types.Nano {
				resultType = types.NanoType
			}
		}
		target := builder.newTemp()
		builder.emit(&BinaryOperation{Target: target, Operator: node.Operator, Left: left.Name, Right: right.Name, Type: resultType})
		return valueRef{Name: target, Type: resultType}
	case *ast.CallExpression:
		calleeName, function, receiver, direct := builder.resolveCallee(node.Callee)
		args := make([]string, 0, len(node.Arguments)+1)
		if receiver != nil {
			args = append(args, builder.buildExpression(receiver).Name)
		}
		for _, argument := range node.Arguments {
			args = append(args, builder.buildExpression(argument).Name)
		}
		if !direct {
			callee := builder.buildExpression(node.Callee)
			returnType := types.AnyType
			if callee.Type.Kind == types.Function && callee.Type.ReturnType != nil {
				returnType = *callee.Type.ReturnType
			}
			target := ""
			if returnType.Kind != types.Void {
				target = builder.newTemp()
			}
			builder.emit(&IndirectCall{Target: target, Callee: callee.Name, Args: args, Type: returnType})
			return valueRef{Name: target, Type: returnType}
		}
		target := ""
		if function.ReturnType.Kind != types.Void {
			target = builder.newTemp()
		}
		builder.emit(&Call{Target: target, Function: calleeName, Args: args, Type: function.ReturnType})
		return valueRef{Name: target, Type: function.ReturnType}
	default:
		builder.errorAt(expression.Position(), "unsupported expression %T", expression)
		return valueRef{Name: "<invalid>", Type: types.InvalidType}
	}
}

func (builder *Builder) resolveCallee(expression ast.Expression) (string, sema.FunctionSymbol, ast.Expression, bool) {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		if name, function, ok := builder.lookupFunction(node.Name); ok {
			return name, function, nil, true
		}
		return "", sema.FunctionSymbol{}, nil, false
	case *ast.SelectorExpression:
		if parts, ok := selectorPath(node); ok {
			path := strings.Join(parts, ".")
			if name, function, exists := builder.lookupFunction(path); exists {
				return name, function, nil, true
			}
		}
		leftType := builder.expressionType(node.Left)
		name := receiverOwner(leftType) + "." + node.Name
		if function, ok := builder.functions[name]; ok && !function.Static {
			qualified := function.Qualified
			if qualified == "" {
				qualified = name
			}
			return qualified, function, node.Left, true
		}
		return "", sema.FunctionSymbol{}, nil, false
	default:
		return "", sema.FunctionSymbol{}, nil, false
	}
}

func (builder *Builder) lookupFunction(name string) (string, sema.FunctionSymbol, bool) {
	if function, ok := builder.functions[name]; ok {
		return name, function, true
	}
	qualified := qualify(builder.moduleName, name)
	function, ok := builder.functions[qualified]
	return qualified, function, ok
}

func (builder *Builder) expressionType(expression ast.Expression) types.Type {
	if typ, ok := builder.expressionTypes[expression]; ok {
		return typ
	}
	return types.InvalidType
}

func (builder *Builder) buildLambda(expression *ast.LambdaExpression) (string, types.Type, []string) {
	lambdaType := builder.expressionType(expression)
	builder.lambdaIndex++
	name := fmt.Sprintf("%s.$lambda%d", builder.function.Name, builder.lambdaIndex)
	function := Function{
		Name:       name,
		Visibility: ast.VisibilityPrivate,
		ReturnType: types.VoidType,
		Lambda:     true,
	}
	captureTypes := builder.visibleVariables()
	captureNames := make([]string, 0, len(captureTypes))
	for captureName := range captureTypes {
		captureNames = append(captureNames, captureName)
	}
	sort.Strings(captureNames)
	for _, captureName := range captureNames {
		function.Captures = append(function.Captures, Parameter{Name: captureName, Type: captureTypes[captureName]})
	}
	if lambdaType.ReturnType != nil {
		function.ReturnType = *lambdaType.ReturnType
	}
	for index, parameter := range expression.Parameters {
		parameterType := types.InvalidType
		if index < len(lambdaType.Parameters) {
			parameterType = lambdaType.Parameters[index]
		}
		function.Parameters = append(function.Parameters, Parameter{Name: parameter.Name, Type: parameterType})
	}

	previousFunction := builder.function
	previousModuleName := builder.moduleName
	previousBlockIndex := builder.blockIndex
	previousTempIndex := builder.tempIndex
	previousScopes := builder.scopes
	previousTryFrames := builder.activeTryFrames

	builder.function = &function
	builder.blockIndex = 0
	builder.tempIndex = 0
	builder.scopes = nil
	builder.activeTryFrames = nil
	builder.pushScope()
	for _, capture := range function.Captures {
		builder.declare(capture.Name, capture.Type)
	}
	for _, parameter := range function.Parameters {
		builder.declare(parameter.Name, parameter.Type)
	}
	builder.appendBlock("entry")
	builder.buildBlock(expression.Body)
	if !builder.currentBlockTerminated() && function.ReturnType.Kind == types.Void {
		builder.emit(&Return{})
	}
	function = *builder.function
	builder.popScope()

	builder.function = previousFunction
	builder.moduleName = previousModuleName
	builder.blockIndex = previousBlockIndex
	builder.tempIndex = previousTempIndex
	builder.scopes = previousScopes
	builder.activeTryFrames = previousTryFrames
	builder.generatedFunctions = append(builder.generatedFunctions, function)
	return name, lambdaType, captureNames
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

func functionType(function sema.FunctionSymbol) types.Type {
	parameters := make([]types.Type, 0, len(function.Parameters))
	for _, parameter := range function.Parameters {
		parameters = append(parameters, parameter.Type)
	}
	return types.FunctionType(parameters, function.ReturnType)
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
	return types.AnyType
}

func propertyFunction(typ types.Type, name string) string {
	if name != "length" {
		if typ.Kind == types.Any && name == "type" {
			return "values.type"
		}
		return ""
	}
	switch typ.Kind {
	case types.String:
		return "strings.length"
	case types.List:
		return "lists.length"
	case types.Dictionary:
		return "dictionaries.length"
	case types.Tuple:
		return "tuples.length"
	case types.Set:
		return "sets.length"
	default:
		return ""
	}
}

func (builder *Builder) visibleVariables() map[string]types.Type {
	visible := map[string]types.Type{}
	for index := len(builder.scopes) - 1; index >= 0; index-- {
		for name, typ := range builder.scopes[index] {
			if _, exists := visible[name]; !exists {
				visible[name] = typ
			}
		}
	}
	return visible
}

func (builder *Builder) emitZero(typ types.Type) valueRef {
	target := builder.newTemp()
	value := "0"
	if typ.Kind == types.Float || typ.Kind == types.Nano {
		value = "0.0"
	}
	builder.emit(&Constant{Target: target, Type: typ, Value: value})
	return valueRef{Name: target, Type: typ}
}

func (builder *Builder) literalType(literal *ast.LiteralExpression) types.Type {
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

func (builder *Builder) lookupType(name string, position token.Position) types.Type {
	typ, ok := types.Lookup(name)
	if ok {
		return typ
	}
	if _, exists := builder.types[name]; exists {
		return types.StructType(name)
	}
	qualified := qualify(builder.moduleName, name)
	if _, exists := builder.types[qualified]; exists {
		return types.StructType(qualified)
	}
	builder.errorAt(position, "unknown type %q", name)
	return types.InvalidType
}

func (builder *Builder) lookupVariable(name string) types.Type {
	for i := len(builder.scopes) - 1; i >= 0; i-- {
		if typ, ok := builder.scopes[i][name]; ok {
			return typ
		}
	}
	return types.InvalidType
}

func (builder *Builder) declare(name string, typ types.Type) {
	builder.scopes[len(builder.scopes)-1][name] = typ
}

func (builder *Builder) pushScope() {
	builder.scopes = append(builder.scopes, map[string]types.Type{})
}

func (builder *Builder) popScope() {
	builder.scopes = builder.scopes[:len(builder.scopes)-1]
}

func (builder *Builder) appendBlock(name string) {
	builder.function.Blocks = append(builder.function.Blocks, Block{Name: name})
}

func (builder *Builder) currentBlock() *Block {
	return &builder.function.Blocks[len(builder.function.Blocks)-1]
}

func (builder *Builder) emit(instruction Instruction) {
	builder.currentBlock().Instructions = append(builder.currentBlock().Instructions, instruction)
}

func (builder *Builder) emitJumpIfNeeded(target string) {
	if builder.currentBlockTerminated() {
		return
	}
	builder.emit(&Jump{Target: target})
}

func (builder *Builder) currentBlockTerminated() bool {
	block := builder.currentBlock()
	return len(block.Instructions) > 0 && isTerminator(block.Instructions[len(block.Instructions)-1])
}

func (builder *Builder) newTemp() string {
	builder.tempIndex++
	return fmt.Sprintf("%%%d", builder.tempIndex)
}

func (builder *Builder) newBlockName(prefix string) string {
	builder.blockIndex++
	return fmt.Sprintf("%s.%d", prefix, builder.blockIndex)
}

func (builder *Builder) errorAt(position token.Position, format string, args ...any) {
	builder.diagnostics = append(builder.diagnostics, diagnostic.Error(position, fmt.Sprintf(format, args...)))
}

func isTerminator(instruction Instruction) bool {
	switch instruction.(type) {
	case *Return, *Branch, *Jump, *TryBegin, *Throw:
		return true
	default:
		return false
	}
}

func moduleName(program *ast.Program) string {
	for _, declaration := range program.Declarations {
		module, ok := declaration.(*ast.ModuleDeclaration)
		if ok {
			return module.Name
		}
	}
	return ""
}

func qualify(moduleName, symbolName string) string {
	if moduleName == "" {
		return symbolName
	}
	return moduleName + "." + symbolName
}
