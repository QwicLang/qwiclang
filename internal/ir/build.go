package ir

import (
	"fmt"

	"qwiclang/internal/ast"
	"qwiclang/internal/diagnostic"
	"qwiclang/internal/sema"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Diagnostic = diagnostic.Diagnostic

type Builder struct {
	functions   map[string]sema.FunctionSymbol
	names       map[*ast.FunctionDeclaration]string
	diagnostics []Diagnostic
	function    *Function
	moduleName  string
	blockIndex  int
	tempIndex   int
	scopes      []map[string]types.Type
}

func Build(result *sema.Result) (Module, []Diagnostic) {
	return NewBuilder(result.Functions, result.FunctionNames).BuildPrograms(result.Programs)
}

func BuildSource(filename, source string) (Module, []Diagnostic) {
	result, diagnostics := sema.Check(filename, source)
	if len(diagnostics) > 0 {
		return Module{}, diagnostics
	}
	return Build(result)
}

func NewBuilder(functions map[string]sema.FunctionSymbol, names map[*ast.FunctionDeclaration]string) *Builder {
	return &Builder{functions: functions, names: names}
}

func (builder *Builder) BuildProgram(program *ast.Program) (Module, []Diagnostic) {
	return builder.BuildPrograms([]*ast.Program{program})
}

func (builder *Builder) BuildPrograms(programs []*ast.Program) (Module, []Diagnostic) {
	module := Module{}
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
	return module, builder.diagnostics
}

func (builder *Builder) buildFunction(declaration *ast.FunctionDeclaration, moduleName string) Function {
	qualifiedName := builder.names[declaration]
	if qualifiedName == "" {
		qualifiedName = qualify(moduleName, declaration.Name)
	}
	symbol := builder.functions[qualifiedName]
	function := Function{
		Name:       qualifiedName,
		Visibility: declaration.Visibility,
		Turbo:      declaration.Turbo,
		ReturnType: symbol.ReturnType,
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
	builder.pushScope()
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
		builder.emit(&Store{Target: node.Name, Value: value.Name})
	case *ast.ReturnStatement:
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
	default:
		builder.errorAt(statement.Position(), "unsupported statement %T", statement)
	}
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
		target := builder.newTemp()
		builder.emit(&Load{Target: target, Source: node.Name, Type: variableType})
		return valueRef{Name: target, Type: variableType}
	case *ast.LiteralExpression:
		target := builder.newTemp()
		literalType := builder.literalType(node)
		builder.emit(&Constant{Target: target, Type: literalType, Value: node.Value})
		return valueRef{Name: target, Type: literalType}
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
		calleeName, ok := builder.resolveCallee(node.Callee)
		if !ok {
			builder.errorAt(node.Position(), "function call target must be an identifier")
			return valueRef{Name: "<invalid>", Type: types.InvalidType}
		}
		function := builder.functions[calleeName]
		args := make([]string, 0, len(node.Arguments))
		for _, argument := range node.Arguments {
			args = append(args, builder.buildExpression(argument).Name)
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

func (builder *Builder) resolveCallee(expression ast.Expression) (string, bool) {
	switch node := expression.(type) {
	case *ast.IdentifierExpression:
		if _, ok := builder.functions[node.Name]; ok {
			return node.Name, true
		}
		qualified := qualify(builder.moduleName, node.Name)
		if _, ok := builder.functions[qualified]; ok {
			return qualified, true
		}
		return node.Name, true
	case *ast.SelectorExpression:
		module, ok := node.Left.(*ast.IdentifierExpression)
		if !ok {
			return "", false
		}
		return qualify(module.Name, node.Name), true
	default:
		return "", false
	}
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
	if !ok {
		builder.errorAt(position, "unknown type %q", name)
		return types.InvalidType
	}
	return typ
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
	block := builder.currentBlock()
	if len(block.Instructions) > 0 && isTerminator(block.Instructions[len(block.Instructions)-1]) {
		return
	}
	builder.emit(&Jump{Target: target})
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
	case *Return, *Branch, *Jump:
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
