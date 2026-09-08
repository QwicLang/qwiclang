package ir

import (
	"fmt"
	"strings"

	"qwiclang/internal/ast"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Module struct {
	Functions []Function
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
	ReturnType types.Type
	Blocks     []Block
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
