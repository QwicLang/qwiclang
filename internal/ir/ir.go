package ir

import (
	"fmt"
	"strings"

	"qwiclang/internal/ast"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Module struct {
	Types     []TypeDefinition
	Functions []Function
}

type TypeDefinition struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type types.Type
}

func (module Module) DebugString() string {
	var builder strings.Builder
	builder.WriteString("Module\n")
	for _, function := range module.Functions {
		fmt.Fprintf(&builder, "  Function name=%s visibility=%s turbo=%t return=%s\n", function.Name, function.Visibility, function.Turbo, function.ReturnType)
		for _, parameter := range function.Parameters {
			fmt.Fprintf(&builder, "    Parameter name=%s type=%s\n", parameter.Name, parameter.Type)
		}
		for _, block := range function.Blocks {
			fmt.Fprintf(&builder, "    Block name=%s\n", block.Name)
			for _, instruction := range block.Instructions {
				fmt.Fprintf(&builder, "      %s\n", instruction.String())
			}
		}
	}
	return strings.TrimRight(builder.String(), "\n")
}

type Function struct {
	Name       string
	Visibility ast.Visibility
	Turbo      bool
	Parameters []Parameter
	Captures   []Parameter
	ReturnType types.Type
	Blocks     []Block
	Lambda     bool
}

type Parameter struct {
	Name string
	Type types.Type
}

type Block struct {
	Name         string
	Instructions []Instruction
}

type Instruction interface {
	instructionNode()
	String() string
}

type Constant struct {
	Target string
	Type   types.Type
	Value  string
}

func (*Constant) instructionNode() {}

func (instruction *Constant) String() string {
	return fmt.Sprintf("%s = constant %s %s", instruction.Target, instruction.Type, instruction.Value)
}

type Variable struct {
	Name    string
	Type    types.Type
	Mutable bool
}

func (*Variable) instructionNode() {}

func (instruction *Variable) String() string {
	kind := "const"
	if instruction.Mutable {
		kind = "let"
	}
	return fmt.Sprintf("%s %s: %s", kind, instruction.Name, instruction.Type)
}

type Load struct {
	Target string
	Source string
	Type   types.Type
}

func (*Load) instructionNode() {}

func (instruction *Load) String() string {
	return fmt.Sprintf("%s = load %s: %s", instruction.Target, instruction.Source, instruction.Type)
}

type Store struct {
	Target string
	Value  string
}

func (*Store) instructionNode() {}

func (instruction *Store) String() string {
	return fmt.Sprintf("store %s, %s", instruction.Target, instruction.Value)
}

type BinaryOperation struct {
	Target   string
	Operator token.Kind
	Left     string
	Right    string
	Type     types.Type
}

func (*BinaryOperation) instructionNode() {}

func (instruction *BinaryOperation) String() string {
	return fmt.Sprintf("%s = binary %s %s, %s: %s", instruction.Target, instruction.Operator, instruction.Left, instruction.Right, instruction.Type)
}

type Call struct {
	Target   string
	Function string
	Args     []string
	Type     types.Type
}

func (*Call) instructionNode() {}

func (instruction *Call) String() string {
	prefix := ""
	if instruction.Target != "" {
		prefix = instruction.Target + " = "
	}
	return fmt.Sprintf("%scall %s(%s): %s", prefix, instruction.Function, strings.Join(instruction.Args, ", "), instruction.Type)
}

type FunctionReference struct {
	Target   string
	Function string
	Type     types.Type
	Captures []string
}

func (*FunctionReference) instructionNode() {}

func (instruction *FunctionReference) String() string {
	return fmt.Sprintf("%s = function %s captures=(%s): %s", instruction.Target, instruction.Function, strings.Join(instruction.Captures, ", "), instruction.Type)
}

type IndirectCall struct {
	Target string
	Callee string
	Args   []string
	Type   types.Type
}

func (*IndirectCall) instructionNode() {}

func (instruction *IndirectCall) String() string {
	prefix := ""
	if instruction.Target != "" {
		prefix = instruction.Target + " = "
	}
	return fmt.Sprintf("%scall-indirect %s(%s): %s", prefix, instruction.Callee, strings.Join(instruction.Args, ", "), instruction.Type)
}

type NewStruct struct {
	Target string
	Type   types.Type
}

func (*NewStruct) instructionNode() {}

func (instruction *NewStruct) String() string {
	return fmt.Sprintf("%s = new %s", instruction.Target, instruction.Type)
}

type LoadField struct {
	Target string
	Object string
	Field  string
	Type   types.Type
}

func (*LoadField) instructionNode() {}

func (instruction *LoadField) String() string {
	return fmt.Sprintf("%s = load-field %s.%s: %s", instruction.Target, instruction.Object, instruction.Field, instruction.Type)
}

type StoreField struct {
	Object string
	Field  string
	Value  string
}

func (*StoreField) instructionNode() {}

func (instruction *StoreField) String() string {
	return fmt.Sprintf("store-field %s.%s, %s", instruction.Object, instruction.Field, instruction.Value)
}

type TryBegin struct {
	Frame      string
	TryBlock   string
	CatchBlock string
}

func (*TryBegin) instructionNode() {}

func (instruction *TryBegin) String() string {
	return fmt.Sprintf("try %s, %s, %s", instruction.Frame, instruction.TryBlock, instruction.CatchBlock)
}

type TryEnd struct {
	Frame string
}

func (*TryEnd) instructionNode() {}

func (instruction *TryEnd) String() string {
	return "try-end " + instruction.Frame
}

type Catch struct {
	Target string
	Frame  string
}

func (*Catch) instructionNode() {}

func (instruction *Catch) String() string {
	return fmt.Sprintf("%s = catch %s", instruction.Target, instruction.Frame)
}

type Throw struct {
	Value string
}

func (*Throw) instructionNode() {}

func (instruction *Throw) String() string {
	return "throw " + instruction.Value
}

type FormatPart struct {
	Text  string
	Value string
	Type  types.Type
}

type FormatString struct {
	Target string
	Parts  []FormatPart
	Type   types.Type
}

func (*FormatString) instructionNode() {}

func (instruction *FormatString) String() string {
	return fmt.Sprintf("%s = format-string %d part(s): %s", instruction.Target, len(instruction.Parts), instruction.Type)
}

type Return struct {
	Value string
}

func (*Return) instructionNode() {}

func (instruction *Return) String() string {
	if instruction.Value == "" {
		return "return"
	}
	return fmt.Sprintf("return %s", instruction.Value)
}

type Branch struct {
	Condition string
	ThenBlock string
	ElseBlock string
}

func (*Branch) instructionNode() {}

func (instruction *Branch) String() string {
	return fmt.Sprintf("branch %s, %s, %s", instruction.Condition, instruction.ThenBlock, instruction.ElseBlock)
}

type Jump struct {
	Target string
}

func (*Jump) instructionNode() {}

func (instruction *Jump) String() string {
	return fmt.Sprintf("jump %s", instruction.Target)
}
