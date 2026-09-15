package ast

import (
	"fmt"
	"strings"

	"qwiclang/internal/token"
)

type Visibility int

const (
	VisibilityPrivate Visibility = iota
	VisibilityPublic
)

func (visibility Visibility) String() string {
	switch visibility {
	case VisibilityPublic:
		return "public"
	default:
		return "private"
	}
}

type Node interface {
	Position() token.Position
}

type Declaration interface {
	Node
	declarationNode()
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Declarations []Declaration
}

func (program *Program) Position() token.Position {
	if len(program.Declarations) == 0 {
		return token.Position{Line: 1, Column: 1}
	}
	return program.Declarations[0].Position()
}

func (program *Program) DebugString() string {
	var builder strings.Builder
	writeNode(&builder, program, 0)
	return strings.TrimRight(builder.String(), "\n")
}

type Parameter struct {
	Name     string
	TypeName string
	Pos      token.Position
}

type TypeField struct {
	Name     string
	TypeName string
	Pos      token.Position
}

type TypeDeclaration struct {
	Name       string
	Visibility Visibility
	Fields     []TypeField
	Pos        token.Position
}

func (*TypeDeclaration) declarationNode() {}

func (declaration *TypeDeclaration) Position() token.Position {
	return declaration.Pos
}

type ModuleDeclaration struct {
	Name string
	Pos  token.Position
}

func (*ModuleDeclaration) declarationNode() {}

func (declaration *ModuleDeclaration) Position() token.Position {
	return declaration.Pos
}

type ImportDeclaration struct {
	Name string
	Pos  token.Position
}

func (*ImportDeclaration) declarationNode() {}

func (declaration *ImportDeclaration) Position() token.Position {
	return declaration.Pos
}

type FunctionDeclaration struct {
	Name       string
	Owner      string
	Receiver   string
	Visibility Visibility
	Turbo      bool
	Parameters []Parameter
	ReturnType string
	Body       *BlockStatement
	Pos        token.Position
}

func (*FunctionDeclaration) declarationNode() {}

func (declaration *FunctionDeclaration) Position() token.Position {
	return declaration.Pos
}

type BlockStatement struct {
	Statements []Statement
	Pos        token.Position
}

func (*BlockStatement) statementNode() {}

func (statement *BlockStatement) Position() token.Position {
	return statement.Pos
}

type VariableDeclaration struct {
	Mutable  bool
	Name     string
	TypeName string
	Value    Expression
	Pos      token.Position
}

func (*VariableDeclaration) statementNode() {}

func (statement *VariableDeclaration) Position() token.Position {
	return statement.Pos
}

type AssignmentStatement struct {
	Target Expression
	Value  Expression
	Pos    token.Position
}

func (*AssignmentStatement) statementNode() {}

func (statement *AssignmentStatement) Position() token.Position {
	return statement.Pos
}

type ReturnStatement struct {
	Value Expression
	Pos   token.Position
}

func (*ReturnStatement) statementNode() {}

func (statement *ReturnStatement) Position() token.Position {
	return statement.Pos
}

type ExpressionStatement struct {
	Expression Expression
	Pos        token.Position
}

func (*ExpressionStatement) statementNode() {}

func (statement *ExpressionStatement) Position() token.Position {
	return statement.Pos
}

type IfStatement struct {
	Condition  Expression
	ThenBranch *BlockStatement
	ElseBranch *BlockStatement
	Pos        token.Position
}

func (*IfStatement) statementNode() {}

func (statement *IfStatement) Position() token.Position {
	return statement.Pos
}

type WhileStatement struct {
	Condition Expression
	Body      *BlockStatement
	Pos       token.Position
}

type TryStatement struct {
	TryBlock      *BlockStatement
	CatchVariable string
	CatchBlock    *BlockStatement
	Pos           token.Position
}

func (*TryStatement) statementNode() {}

func (statement *TryStatement) Position() token.Position {
	return statement.Pos
}

type ThrowStatement struct {
	Value Expression
	Pos   token.Position
}

func (*ThrowStatement) statementNode() {}

func (statement *ThrowStatement) Position() token.Position {
	return statement.Pos
}

func (*WhileStatement) statementNode() {}

func (statement *WhileStatement) Position() token.Position {
	return statement.Pos
}

// ForStatement represents: for variable in iterable { ... }
type ForStatement struct {
	Variable string
	Iterable Expression
	Body     *BlockStatement
	Pos      token.Position
}

func (*ForStatement) statementNode() {}

func (statement *ForStatement) Position() token.Position {
	return statement.Pos
}

type IdentifierExpression struct {
	Name string
	Pos  token.Position
}

func (*IdentifierExpression) expressionNode() {}

func (expression *IdentifierExpression) Position() token.Position {
	return expression.Pos
}

type SelectorExpression struct {
	Left Expression
	Name string
	Pos  token.Position
}

func (*SelectorExpression) expressionNode() {}

func (expression *SelectorExpression) Position() token.Position {
	return expression.Pos
}

type LiteralKind int

const (
	LiteralInteger LiteralKind = iota
	LiteralFloat
	LiteralString
	LiteralBool
	LiteralNull
)

func (kind LiteralKind) String() string {
	switch kind {
	case LiteralInteger:
		return "integer"
	case LiteralFloat:
		return "float"
	case LiteralString:
		return "string"
	case LiteralBool:
		return "bool"
	case LiteralNull:
		return "null"
	default:
		return "unknown"
	}
}

type LiteralExpression struct {
	Kind  LiteralKind
	Value string
	Pos   token.Position
}

func (*LiteralExpression) expressionNode() {}

func (expression *LiteralExpression) Position() token.Position {
	return expression.Pos
}

type InterpolatedStringPart struct {
	Text       string
	Expression Expression
}

type InterpolatedStringExpression struct {
	Parts []InterpolatedStringPart
	Pos   token.Position
}

func (*InterpolatedStringExpression) expressionNode() {}

func (expression *InterpolatedStringExpression) Position() token.Position {
	return expression.Pos
}

type UnaryExpression struct {
	Operator token.Kind
	Right    Expression
	Pos      token.Position
}

func (*UnaryExpression) expressionNode() {}

func (expression *UnaryExpression) Position() token.Position {
	return expression.Pos
}

type BinaryExpression struct {
	Left     Expression
	Operator token.Kind
	Right    Expression
	Pos      token.Position
}

type RangeExpression struct {
	Start Expression
	End   Expression
	Pos   token.Position
}

func (*RangeExpression) expressionNode() {}

func (expression *RangeExpression) Position() token.Position {
	return expression.Pos
}

func (*BinaryExpression) expressionNode() {}

func (expression *BinaryExpression) Position() token.Position {
	return expression.Pos
}

type CallExpression struct {
	Callee    Expression
	Arguments []Expression
	Pos       token.Position
}

type LambdaExpression struct {
	Parameters []Parameter
	ReturnType string
	Body       *BlockStatement
	Pos        token.Position
}

func (*LambdaExpression) expressionNode() {}

func (expression *LambdaExpression) Position() token.Position {
	return expression.Pos
}

type TypeFieldValue struct {
	Name  string
	Value Expression
	Pos   token.Position
}

type TypeLiteralExpression struct {
	Type   Expression
	Fields []TypeFieldValue
	Pos    token.Position
}

func (*TypeLiteralExpression) expressionNode() {}

func (expression *TypeLiteralExpression) Position() token.Position {
	return expression.Pos
}

func (*CallExpression) expressionNode() {}

func (expression *CallExpression) Position() token.Position {
	return expression.Pos
}

// ArrayLiteralExpression represents: [1, 2, "a", "b"]
type ArrayLiteralExpression struct {
	Elements []Expression
	Pos      token.Position
}

func (*ArrayLiteralExpression) expressionNode() {}

func (expression *ArrayLiteralExpression) Position() token.Position {
	return expression.Pos
}

// DictionaryPair represents a key-value pair in a dictionary literal
type DictionaryPair struct {
	Key   Expression
	Value Expression
}

// DictionaryLiteralExpression represents: {"a": 1, "b": 2}
type DictionaryLiteralExpression struct {
	Pairs []DictionaryPair
	Pos   token.Position
}

func (*DictionaryLiteralExpression) expressionNode() {}

func (expression *DictionaryLiteralExpression) Position() token.Position {
	return expression.Pos
}

// TupleLiteralExpression represents: (1, 2, 3, 4)
type TupleLiteralExpression struct {
	Elements []Expression
	Pos      token.Position
}

func (*TupleLiteralExpression) expressionNode() {}

func (expression *TupleLiteralExpression) Position() token.Position {
	return expression.Pos
}

// SliceExpression represents: collection[start:end]
type SliceExpression struct {
	Left  Expression
	Start Expression
	End   Expression
	Pos   token.Position
}

func (*SliceExpression) expressionNode() {}

func (expression *SliceExpression) Position() token.Position {
	return expression.Pos
}

// IndexExpression represents: collection[index]
type IndexExpression struct {
	Left  Expression
	Index Expression
	Pos   token.Position
}

func (*IndexExpression) expressionNode() {}

func (expression *IndexExpression) Position() token.Position {
	return expression.Pos
}

func writeNode(builder *strings.Builder, node Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch n := node.(type) {
	case *Program:
		builder.WriteString("Program\n")
		for _, declaration := range n.Declarations {
			writeNode(builder, declaration, indent+1)
		}
	case *FunctionDeclaration:
		name := n.Name
		if n.Owner != "" {
			name = n.Owner + "." + n.Name
		}
		if n.Receiver != "" {
			fmt.Fprintf(builder, "%sFunctionDeclaration name=%s receiver=%s visibility=%s turbo=%t return=%s\n", prefix, name, n.Receiver, n.Visibility, n.Turbo, n.ReturnType)
		} else {
			fmt.Fprintf(builder, "%sFunctionDeclaration name=%s visibility=%s turbo=%t return=%s\n", prefix, name, n.Visibility, n.Turbo, n.ReturnType)
		}
		for _, parameter := range n.Parameters {
			fmt.Fprintf(builder, "%s  Parameter name=%s type=%s\n", prefix, parameter.Name, parameter.TypeName)
		}
		writeNode(builder, n.Body, indent+1)
	case *ModuleDeclaration:
		fmt.Fprintf(builder, "%sModuleDeclaration name=%s\n", prefix, n.Name)
	case *ImportDeclaration:
		fmt.Fprintf(builder, "%sImportDeclaration name=%s\n", prefix, n.Name)
	case *TypeDeclaration:
		fmt.Fprintf(builder, "%sTypeDeclaration name=%s visibility=%s\n", prefix, n.Name, n.Visibility)
		for _, field := range n.Fields {
			fmt.Fprintf(builder, "%s  Field name=%s type=%s\n", prefix, field.Name, field.TypeName)
		}
	case *BlockStatement:
		builder.WriteString(prefix + "BlockStatement\n")
		for _, statement := range n.Statements {
			writeNode(builder, statement, indent+1)
		}
	case *VariableDeclaration:
		mutability := "const"
		if n.Mutable {
			mutability = "let"
		}
		fmt.Fprintf(builder, "%sVariableDeclaration kind=%s name=%s type=%s\n", prefix, mutability, n.Name, n.TypeName)
		writeNode(builder, n.Value, indent+1)
	case *AssignmentStatement:
		builder.WriteString(prefix + "AssignmentStatement\n")
		writeNode(builder, n.Target, indent+1)
		writeNode(builder, n.Value, indent+1)
	case *ReturnStatement:
		builder.WriteString(prefix + "ReturnStatement\n")
		if n.Value != nil {
			writeNode(builder, n.Value, indent+1)
		}
	case *ExpressionStatement:
		builder.WriteString(prefix + "ExpressionStatement\n")
		writeNode(builder, n.Expression, indent+1)
	case *IfStatement:
		builder.WriteString(prefix + "IfStatement\n")
		writeNode(builder, n.Condition, indent+1)
		writeNode(builder, n.ThenBranch, indent+1)
		if n.ElseBranch != nil {
			writeNode(builder, n.ElseBranch, indent+1)
		}
	case *WhileStatement:
		builder.WriteString(prefix + "WhileStatement\n")
		writeNode(builder, n.Condition, indent+1)
		writeNode(builder, n.Body, indent+1)
	case *TryStatement:
		fmt.Fprintf(builder, "%sTryStatement catch=%s\n", prefix, n.CatchVariable)
		writeNode(builder, n.TryBlock, indent+1)
		writeNode(builder, n.CatchBlock, indent+1)
	case *ThrowStatement:
		builder.WriteString(prefix + "ThrowStatement\n")
		writeNode(builder, n.Value, indent+1)
	case *ForStatement:
		fmt.Fprintf(builder, "%sForStatement variable=%s\n", prefix, n.Variable)
		writeNode(builder, n.Iterable, indent+1)
		writeNode(builder, n.Body, indent+1)
	case *IdentifierExpression:
		fmt.Fprintf(builder, "%sIdentifierExpression name=%s\n", prefix, n.Name)
	case *SelectorExpression:
		fmt.Fprintf(builder, "%sSelectorExpression name=%s\n", prefix, n.Name)
		writeNode(builder, n.Left, indent+1)
	case *LiteralExpression:
		fmt.Fprintf(builder, "%sLiteralExpression kind=%s value=%s\n", prefix, n.Kind, n.Value)
	case *InterpolatedStringExpression:
		builder.WriteString(prefix + "InterpolatedStringExpression\n")
		for _, part := range n.Parts {
			if part.Expression == nil {
				fmt.Fprintf(builder, "%s  StringPart text=%q\n", prefix, part.Text)
				continue
			}
			builder.WriteString(prefix + "  Interpolation\n")
			writeNode(builder, part.Expression, indent+2)
		}
	case *UnaryExpression:
		fmt.Fprintf(builder, "%sUnaryExpression operator=%s\n", prefix, n.Operator)
		writeNode(builder, n.Right, indent+1)
	case *BinaryExpression:
		fmt.Fprintf(builder, "%sBinaryExpression operator=%s\n", prefix, n.Operator)
		writeNode(builder, n.Left, indent+1)
		writeNode(builder, n.Right, indent+1)
	case *RangeExpression:
		builder.WriteString(prefix + "RangeExpression\n")
		writeNode(builder, n.Start, indent+1)
		writeNode(builder, n.End, indent+1)
	case *CallExpression:
		builder.WriteString(prefix + "CallExpression\n")
		writeNode(builder, n.Callee, indent+1)
		for _, argument := range n.Arguments {
			writeNode(builder, argument, indent+1)
		}
	case *LambdaExpression:
		fmt.Fprintf(builder, "%sLambdaExpression return=%s\n", prefix, n.ReturnType)
		for _, parameter := range n.Parameters {
			fmt.Fprintf(builder, "%s  Parameter name=%s type=%s\n", prefix, parameter.Name, parameter.TypeName)
		}
		writeNode(builder, n.Body, indent+1)
	case *TypeLiteralExpression:
		builder.WriteString(prefix + "TypeLiteralExpression\n")
		writeNode(builder, n.Type, indent+1)
		for _, field := range n.Fields {
			fmt.Fprintf(builder, "%s  FieldValue name=%s\n", prefix, field.Name)
			writeNode(builder, field.Value, indent+2)
		}
	case *ArrayLiteralExpression:
		fmt.Fprintf(builder, "%sArrayLiteralExpression (length=%d)\n", prefix, len(n.Elements))
		for _, elem := range n.Elements {
			writeNode(builder, elem, indent+1)
		}
	case *DictionaryLiteralExpression:
		fmt.Fprintf(builder, "%sDictionaryLiteralExpression (length=%d)\n", prefix, len(n.Pairs))
		for _, pair := range n.Pairs {
			builder.WriteString(prefix + "  Pair\n")
			writeNode(builder, pair.Key, indent+2)
			writeNode(builder, pair.Value, indent+2)
		}
	case *TupleLiteralExpression:
		fmt.Fprintf(builder, "%sTupleLiteralExpression (length=%d)\n", prefix, len(n.Elements))
		for _, elem := range n.Elements {
			writeNode(builder, elem, indent+1)
		}
	case *IndexExpression:
		builder.WriteString(prefix + "IndexExpression\n")
		writeNode(builder, n.Left, indent+1)
		writeNode(builder, n.Index, indent+1)
	default:
		fmt.Fprintf(builder, "%s<unknown %T>\n", prefix, n)
	}
}
