package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"qwiclang/internal/diagnostic"
	"qwiclang/internal/ir"
	"qwiclang/internal/stdlib"
	"qwiclang/internal/token"
	"qwiclang/internal/types"
)

type Diagnostic = diagnostic.Diagnostic

type Options struct {
	OutputPath  string
	WorkDir     string
	KeepC       bool
	RuntimePath string
}

func BuildExecutable(module ir.Module, options Options) []Diagnostic {
	if options.OutputPath == "" {
		return []Diagnostic{diagnostic.Error(token.Position{Line: 1, Column: 1}, "missing output path")}
	}

	workDir := options.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "qwic-build-*")
		if err != nil {
			return []Diagnostic{diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("create temporary build directory: %v", err))}
		}
		if !options.KeepC {
			defer os.RemoveAll(workDir)
		}
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return []Diagnostic{diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("create build directory: %v", err))}
	}

	cSource, diagnostics := GenerateC(module)
	if len(diagnostics) > 0 {
		return diagnostics
	}

	cPath := filepath.Join(workDir, "main.c")
	if err := os.WriteFile(cPath, []byte(cSource), 0o644); err != nil {
		return []Diagnostic{diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("write generated C source: %v", err))}
	}

	runtimePath := options.RuntimePath
	if runtimePath == "" {
		runtimePath = "runtime"
	}
	runtimeCPath := filepath.Join(runtimePath, "qwic_runtime.c")
	command := exec.Command("cc", cPath, runtimeCPath, "-I", runtimePath, "-o", options.OutputPath)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return []Diagnostic{diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("C compiler failed: %s", message))}
	}

	return nil
}

func GenerateC(module ir.Module) (string, []Diagnostic) {
	generator := cGenerator{
		valueTypes: map[string]types.Type{},
	}
	generator.writePreamble()
	for _, function := range module.Functions {
		generator.writePrototype(function)
	}
	if len(module.Functions) > 0 {
		generator.builder.WriteString("\n")
	}
	for _, function := range module.Functions {
		generator.writeFunction(function)
	}
	return generator.builder.String(), generator.diagnostics
}

type cGenerator struct {
	builder     strings.Builder
	diagnostics []Diagnostic
	valueTypes  map[string]types.Type
}

func (generator *cGenerator) writePreamble() {
	generator.builder.WriteString("#include <stdbool.h>\n")
	generator.builder.WriteString("#include <stdint.h>\n")
	generator.builder.WriteString("#include <stdio.h>\n")
	generator.builder.WriteString("#include \"qwic_runtime.h\"\n\n")
}

func (generator *cGenerator) writePrototype(function ir.Function) {
	fmt.Fprintf(&generator.builder, "%s %s(", cFunctionReturnType(function), cFunctionName(function.Name))
	generator.writeParameters(function.Parameters)
	generator.builder.WriteString(");\n")
}

func (generator *cGenerator) writeFunction(function ir.Function) {
	generator.valueTypes = map[string]types.Type{}

	fmt.Fprintf(&generator.builder, "%s %s(", cFunctionReturnType(function), cFunctionName(function.Name))
	generator.writeParameters(function.Parameters)
	generator.builder.WriteString(") {\n")
	if function.Name == "main" {
		generator.builder.WriteString("  qwic_runtime_init();\n")
	}
	for _, parameter := range function.Parameters {
		generator.valueTypes[parameter.Name] = parameter.Type
	}

	for _, block := range function.Blocks {
		fmt.Fprintf(&generator.builder, "%s: ;\n", cLabel(block.Name))
		for _, instruction := range block.Instructions {
			generator.writeInstruction(function, instruction)
		}
	}

	if function.Name == "main" && function.ReturnType.Kind == types.Void && !functionEndsWithTerminator(function) {
		generator.builder.WriteString("  return qwic_runtime_exit_code();\n")
	}
	generator.builder.WriteString("}\n\n")
}

func (generator *cGenerator) writeParameters(parameters []ir.Parameter) {
	if len(parameters) == 0 {
		generator.builder.WriteString("void")
		return
	}
	for i, parameter := range parameters {
		if i > 0 {
			generator.builder.WriteString(", ")
		}
		fmt.Fprintf(&generator.builder, "%s %s", cType(parameter.Type), cIdentifier(parameter.Name))
	}
}

func (generator *cGenerator) writeInstruction(function ir.Function, instruction ir.Instruction) {
	switch node := instruction.(type) {
	case *ir.Constant:
		generator.valueTypes[node.Target] = node.Type
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(node.Type), cValue(node.Target), cLiteral(node.Type, node.Value))
	case *ir.Variable:
		generator.valueTypes[node.Name] = node.Type
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(node.Type), cIdentifier(node.Name), zeroValue(node.Type))
	case *ir.Load:
		generator.valueTypes[node.Target] = node.Type
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(node.Type), cValue(node.Target), cIdentifier(node.Source))
	case *ir.Store:
		fmt.Fprintf(&generator.builder, "  %s = %s;\n", cIdentifier(node.Target), cValue(node.Value))
	case *ir.BinaryOperation:
		generator.valueTypes[node.Target] = node.Type
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(node.Type), cValue(node.Target), cBinaryExpression(node))
	case *ir.Call:
		generator.writeCall(node)
	case *ir.FormatString:
		generator.writeFormatString(node)
	case *ir.Return:
		generator.writeReturn(function, node)
	case *ir.Branch:
		fmt.Fprintf(&generator.builder, "  if (%s) goto %s; else goto %s;\n", cValue(node.Condition), cLabel(node.ThenBlock), cLabel(node.ElseBlock))
	case *ir.Jump:
		fmt.Fprintf(&generator.builder, "  goto %s;\n", cLabel(node.Target))
	default:
		generator.diagnostics = append(generator.diagnostics, diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("unsupported IR instruction %T", instruction)))
	}
}

func (generator *cGenerator) writeFormatString(instruction *ir.FormatString) {
	generator.valueTypes[instruction.Target] = instruction.Type
	format, args := cFormatString(instruction.Parts)
	lengthName := cValue(instruction.Target) + "_len"
	fmt.Fprintf(&generator.builder, "  size_t %s = (size_t)snprintf(NULL, 0, %s%s) + 1;\n", lengthName, cStringLiteral(format), args)
	fmt.Fprintf(&generator.builder, "  char *%s = qwic_alloc(%s);\n", cValue(instruction.Target), lengthName)
	fmt.Fprintf(&generator.builder, "  snprintf(%s, %s, %s%s);\n", cValue(instruction.Target), lengthName, cStringLiteral(format), args)
}

func (generator *cGenerator) writeCall(call *ir.Call) {
	if call.Function == "print" {
		generator.writePrint(call.Args)
		return
	}

	if call.Target != "" {
		generator.valueTypes[call.Target] = call.Type
		fmt.Fprintf(&generator.builder, "  %s %s = ", cType(call.Type), cValue(call.Target))
	} else {
		generator.builder.WriteString("  ")
	}
	fmt.Fprintf(&generator.builder, "%s(", cFunctionName(call.Function))
	for i, arg := range call.Args {
		if i > 0 {
			generator.builder.WriteString(", ")
		}
		generator.builder.WriteString(cValue(arg))
	}
	generator.builder.WriteString(");\n")
}

func cFormatString(parts []ir.FormatPart) (string, string) {
	var format strings.Builder
	var args strings.Builder
	for _, part := range parts {
		if part.Value == "" {
			format.WriteString(strings.ReplaceAll(part.Text, "%", "%%"))
			continue
		}
		specifier, expression := cFormatArgument(part)
		format.WriteString(specifier)
		args.WriteString(", ")
		args.WriteString(expression)
	}
	return format.String(), args.String()
}

func cFormatArgument(part ir.FormatPart) (string, string) {
	switch part.Type.Kind {
	case types.Int:
		return "%lld", fmt.Sprintf("(long long)%s", cValue(part.Value))
	case types.Float, types.Nano:
		return "%.6f", cValue(part.Value)
	case types.String:
		return "%s", cValue(part.Value)
	case types.Bool:
		return "%s", fmt.Sprintf("(%s ? \"true\" : \"false\")", cValue(part.Value))
	default:
		return "%s", "\"<invalid>\""
	}
}

func (generator *cGenerator) writePrint(args []string) {
	if len(args) != 1 {
		generator.diagnostics = append(generator.diagnostics, diagnostic.Error(token.Position{Line: 1, Column: 1}, "print expects one argument"))
		return
	}

	arg := args[0]
	switch generator.valueTypes[arg].Kind {
	case types.Int:
		fmt.Fprintf(&generator.builder, "  qwic_print_int(%s);\n", cValue(arg))
	case types.Float, types.Nano:
		fmt.Fprintf(&generator.builder, "  qwic_print_float(%s);\n", cValue(arg))
	case types.String:
		fmt.Fprintf(&generator.builder, "  qwic_print_string(%s);\n", cValue(arg))
	case types.Bool:
		fmt.Fprintf(&generator.builder, "  qwic_print_bool(%s);\n", cValue(arg))
	default:
		generator.diagnostics = append(generator.diagnostics, diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("cannot print value of type %q", generator.valueTypes[arg])))
	}
}

func (generator *cGenerator) writeReturn(function ir.Function, instruction *ir.Return) {
	if function.Name == "main" && function.ReturnType.Kind == types.Void {
		generator.builder.WriteString("  return qwic_runtime_exit_code();\n")
		return
	}
	if instruction.Value == "" {
		generator.builder.WriteString("  return;\n")
		return
	}
	fmt.Fprintf(&generator.builder, "  return %s;\n", cValue(instruction.Value))
}

func cBinaryExpression(instruction *ir.BinaryOperation) string {
	if instruction.Operator == token.Not {
		return fmt.Sprintf("!%s", cValue(instruction.Left))
	}

	operator := map[token.Kind]string{
		token.Plus:         "+",
		token.Minus:        "-",
		token.Star:         "*",
		token.Slash:        "/",
		token.Percent:      "%",
		token.Equal:        "==",
		token.NotEqual:     "!=",
		token.Less:         "<",
		token.LessEqual:    "<=",
		token.Greater:      ">",
		token.GreaterEqual: ">=",
		token.And:          "&&",
		token.Or:           "||",
	}[instruction.Operator]
	return fmt.Sprintf("%s %s %s", cValue(instruction.Left), operator, cValue(instruction.Right))
}

func cType(typ types.Type) string {
	switch typ.Kind {
	case types.Void:
		return "void"
	case types.Bool:
		return "bool"
	case types.Int:
		return "int64_t"
	case types.Float, types.Nano:
		return "double"
	case types.String:
		return "const char *"
	case types.List, types.Set, types.Dictionary, types.Tuple:
		return "void *"
	default:
		return "void *"
	}
}

func cFunctionReturnType(function ir.Function) string {
	if function.Name == "main" && function.ReturnType.Kind == types.Void {
		return "int"
	}
	return cType(function.ReturnType)
}

func cLiteral(typ types.Type, value string) string {
	switch typ.Kind {
	case types.Bool:
		if value == "true" {
			return "true"
		}
		return "false"
	case types.Null:
		return "NULL"
	default:
		return value
	}
}

func cStringLiteral(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\':
			builder.WriteString(`\\`)
		case '"':
			builder.WriteString(`\"`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			builder.WriteRune(r)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

func zeroValue(typ types.Type) string {
	switch typ.Kind {
	case types.Bool:
		return "false"
	case types.Int:
		return "0"
	case types.Float, types.Nano:
		return "0.0"
	case types.String:
		return "\"\""
	case types.List, types.Set, types.Dictionary, types.Tuple:
		return "NULL"
	default:
		return "0"
	}
}

func cValue(name string) string {
	if strings.HasPrefix(name, "%") {
		return "qw_tmp_" + strings.TrimPrefix(name, "%")
	}
	return cIdentifier(name)
}

func cIdentifier(name string) string {
	return sanitizeCName(name)
}

func cFunctionName(name string) string {
	if name == "main" {
		return "main"
	}
	if function, ok := stdlib.LookupFunction(name); ok {
		return function.RuntimeName
	}
	return sanitizeCName("qw_func_" + name)
}

func cLabel(name string) string {
	return sanitizeCName("label_" + name)
}

var invalidCName = regexp.MustCompile(`[^A-Za-z0-9_]`)

func sanitizeCName(name string) string {
	name = invalidCName.ReplaceAllString(name, "_")
	if name == "" {
		return "qw_empty"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "qw_" + name
	}
	return name
}

func functionEndsWithTerminator(function ir.Function) bool {
	if len(function.Blocks) == 0 {
		return false
	}
	lastBlock := function.Blocks[len(function.Blocks)-1]
	if len(lastBlock.Instructions) == 0 {
		return false
	}
	switch lastBlock.Instructions[len(lastBlock.Instructions)-1].(type) {
	case *ir.Return, *ir.Branch, *ir.Jump:
		return true
	default:
		return false
	}
}

func RunExecutable(path string) (int, string, error) {
	command := exec.Command(path)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()

	// Normalize Windows \r\n line endings to \n so callers receive
	// consistent output regardless of platform.
	normalized := strings.ReplaceAll(output.String(), "\r\n", "\n")

	if err == nil {
		return 0, normalized, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), normalized, nil
	}
	return 1, normalized, err
}
