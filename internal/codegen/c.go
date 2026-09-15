package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	arguments := []string{cPath, runtimeCPath, "-I", runtimePath, "-lm", "-o", options.OutputPath}
	if runtime.GOOS == "windows" {
		arguments = append(arguments, "-lws2_32")
	}
	command := exec.Command("cc", arguments...)
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
		valueTypes:      map[string]types.Type{},
		functions:       map[string]ir.Function{},
		typeDefinitions: map[string]ir.TypeDefinition{},
	}
	for _, definition := range module.Types {
		generator.typeDefinitions[definition.Name] = definition
	}
	for _, function := range module.Functions {
		generator.functions[function.Name] = function
	}
	generator.writePreamble()
	generator.writeTypeDeclarations(module)
	for _, function := range module.Functions {
		generator.writePrototype(function)
	}
	for _, function := range module.Functions {
		if !function.Lambda && function.Name != "main" {
			generator.writeAdapterPrototype(function)
		}
	}
	if len(module.Functions) > 0 {
		generator.builder.WriteString("\n")
	}
	for _, function := range module.Functions {
		generator.writeFunction(function)
	}
	for _, function := range module.Functions {
		if !function.Lambda && function.Name != "main" {
			generator.writeAdapter(function)
		}
	}
	for _, definition := range module.Types {
		generator.writeStructConverters(definition)
	}
	return generator.builder.String(), generator.diagnostics
}

type cGenerator struct {
	builder         strings.Builder
	diagnostics     []Diagnostic
	valueTypes      map[string]types.Type
	functions       map[string]ir.Function
	typeDefinitions map[string]ir.TypeDefinition
	syntheticIndex  int
}

func (generator *cGenerator) writePreamble() {
	generator.builder.WriteString("#include <stdbool.h>\n")
	generator.builder.WriteString("#include <stdint.h>\n")
	generator.builder.WriteString("#include <stdio.h>\n")
	generator.builder.WriteString("#include <string.h>\n")
	generator.builder.WriteString("#include \"qwic_runtime.h\"\n\n")
}

func (generator *cGenerator) writeTypeDeclarations(module ir.Module) {
	for _, definition := range module.Types {
		name := cStructName(definition.Name)
		fmt.Fprintf(&generator.builder, "typedef struct %s %s;\n", name, name)
	}
	if len(module.Types) > 0 {
		generator.builder.WriteString("\n")
	}

	for _, definition := range module.Types {
		name := cStructName(definition.Name)
		fmt.Fprintf(&generator.builder, "struct %s {\n", name)
		if len(definition.Fields) == 0 {
			generator.builder.WriteString("  unsigned char qw_empty;\n")
		}
		for _, field := range definition.Fields {
			fmt.Fprintf(&generator.builder, "  %s %s;\n", cType(field.Type), cIdentifier(field.Name))
		}
		generator.builder.WriteString("};\n\n")
	}

	for _, definition := range module.Types {
		fmt.Fprintf(&generator.builder, "static qwic_value *%s(%s *value);\n", cToAnyName(definition.Name), cStructName(definition.Name))
		fmt.Fprintf(&generator.builder, "static %s *%s(qwic_value *value);\n", cStructName(definition.Name), cFromAnyName(definition.Name))
	}
	if len(module.Types) > 0 {
		generator.builder.WriteString("\n")
	}

	for _, function := range module.Functions {
		if !function.Lambda || len(function.Captures) == 0 {
			continue
		}
		fmt.Fprintf(&generator.builder, "typedef struct %s {\n", cCaptureTypeName(function.Name))
		for _, capture := range function.Captures {
			fmt.Fprintf(&generator.builder, "  %s %s;\n", cType(capture.Type), cIdentifier(capture.Name))
		}
		fmt.Fprintf(&generator.builder, "} %s;\n\n", cCaptureTypeName(function.Name))
	}
}

func (generator *cGenerator) writePrototype(function ir.Function) {
	if function.Lambda {
		fmt.Fprintf(&generator.builder, "static qwic_value *%s(void *qw_context, qwic_value **qw_args, size_t qw_arg_count);\n", cFunctionName(function.Name))
		return
	}
	prefix := ""
	if function.Turbo && function.Name != "main" {
		prefix = "static inline __attribute__((always_inline, hot, optimize(\"O3\"))) "
	}
	fmt.Fprintf(&generator.builder, "%s%s %s(", prefix, cFunctionReturnType(function), cFunctionName(function.Name))
	generator.writeParameters(function.Parameters)
	generator.builder.WriteString(");\n")
}

func (generator *cGenerator) writeAdapterPrototype(function ir.Function) {
	fmt.Fprintf(&generator.builder, "static qwic_value *%s(void *qw_context, qwic_value **qw_args, size_t qw_arg_count);\n", cAdapterName(function.Name))
}

func (generator *cGenerator) writeFunction(function ir.Function) {
	generator.valueTypes = map[string]types.Type{}

	if function.Lambda {
		fmt.Fprintf(&generator.builder, "static qwic_value *%s(void *qw_context, qwic_value **qw_args, size_t qw_arg_count) {\n", cFunctionName(function.Name))
		generator.builder.WriteString("  (void)qw_arg_count;\n")
		if len(function.Captures) > 0 {
			fmt.Fprintf(&generator.builder, "  %s *qw_captures = (%s *)qw_context;\n", cCaptureTypeName(function.Name), cCaptureTypeName(function.Name))
			for _, capture := range function.Captures {
				generator.valueTypes[capture.Name] = capture.Type
				fmt.Fprintf(&generator.builder, "  %s %s = qw_captures->%s;\n", cType(capture.Type), cIdentifier(capture.Name), cIdentifier(capture.Name))
			}
		} else {
			generator.builder.WriteString("  (void)qw_context;\n")
		}
		for index, parameter := range function.Parameters {
			generator.valueTypes[parameter.Name] = parameter.Type
			fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(parameter.Type), cIdentifier(parameter.Name), generator.adaptExpression(fmt.Sprintf("qw_args[%d]", index), types.AnyType, parameter.Type))
		}
	} else {
		prefix := ""
		if function.Turbo && function.Name != "main" {
			prefix = "static inline __attribute__((always_inline, hot, optimize(\"O3\"))) "
		}
		fmt.Fprintf(&generator.builder, "%s%s %s(", prefix, cFunctionReturnType(function), cFunctionName(function.Name))
		generator.writeParameters(function.Parameters)
		generator.builder.WriteString(") {\n")
		if function.Name == "main" {
			generator.builder.WriteString("  qwic_runtime_init();\n")
		}
		for _, parameter := range function.Parameters {
			generator.valueTypes[parameter.Name] = parameter.Type
		}
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

func (generator *cGenerator) writeAdapter(function ir.Function) {
	fmt.Fprintf(&generator.builder, "static qwic_value *%s(void *qw_context, qwic_value **qw_args, size_t qw_arg_count) {\n", cAdapterName(function.Name))
	generator.builder.WriteString("  (void)qw_context;\n")
	generator.builder.WriteString("  (void)qw_arg_count;\n")
	var call strings.Builder
	fmt.Fprintf(&call, "%s(", cFunctionName(function.Name))
	for index, parameter := range function.Parameters {
		if index > 0 {
			call.WriteString(", ")
		}
		call.WriteString(generator.adaptExpression(fmt.Sprintf("qw_args[%d]", index), types.AnyType, parameter.Type))
	}
	call.WriteString(")")
	if function.ReturnType.Kind == types.Void {
		fmt.Fprintf(&generator.builder, "  %s;\n", call.String())
		generator.builder.WriteString("  return qwic_any_null();\n")
	} else {
		fmt.Fprintf(&generator.builder, "  return %s;\n", generator.adaptExpression(call.String(), function.ReturnType, types.AnyType))
	}
	generator.builder.WriteString("}\n\n")
}

func (generator *cGenerator) writeStructConverters(definition ir.TypeDefinition) {
	fmt.Fprintf(&generator.builder, "static qwic_value *%s(%s *value) {\n", cToAnyName(definition.Name), cStructName(definition.Name))
	generator.builder.WriteString("  if (value == NULL) return qwic_any_null();\n")
	generator.builder.WriteString("  void *dictionary = qwic_dictionaries_new();\n")
	for _, field := range definition.Fields {
		boxed := generator.adaptExpression("value->"+cIdentifier(field.Name), field.Type, types.AnyType)
		fmt.Fprintf(&generator.builder, "  qwic_dictionaries_set(dictionary, %s, %s);\n", cStringLiteral(field.Name), boxed)
	}
	generator.builder.WriteString("  return qwic_any_dictionary(dictionary);\n")
	generator.builder.WriteString("}\n\n")

	fmt.Fprintf(&generator.builder, "static %s *%s(qwic_value *value) {\n", cStructName(definition.Name), cFromAnyName(definition.Name))
	fmt.Fprintf(&generator.builder, "  %s *result = (%s *)qwic_alloc(sizeof(%s));\n", cStructName(definition.Name), cStructName(definition.Name), cStructName(definition.Name))
	for _, field := range definition.Fields {
		dynamic := fmt.Sprintf("qwic_any_field(value, %s)", cStringLiteral(field.Name))
		fmt.Fprintf(&generator.builder, "  result->%s = %s;\n", cIdentifier(field.Name), generator.adaptExpression(dynamic, types.AnyType, field.Type))
	}
	generator.builder.WriteString("  return result;\n")
	generator.builder.WriteString("}\n\n")
}

func (generator *cGenerator) writeFunctionReference(reference *ir.FunctionReference) {
	generator.valueTypes[reference.Target] = reference.Type
	context := "NULL"
	if len(reference.Captures) > 0 {
		contextName := generator.nextSynthetic("captures")
		captureType := cCaptureTypeName(reference.Function)
		fmt.Fprintf(&generator.builder, "  %s *%s = (%s *)qwic_alloc(sizeof(%s));\n", captureType, contextName, captureType, captureType)
		function := generator.functions[reference.Function]
		captureTypes := map[string]types.Type{}
		for _, capture := range function.Captures {
			captureTypes[capture.Name] = capture.Type
		}
		for _, capture := range reference.Captures {
			fmt.Fprintf(&generator.builder, "  %s->%s = %s;\n", contextName, cIdentifier(capture), generator.adaptExpression(cIdentifier(capture), generator.valueTypes[capture], captureTypes[capture]))
		}
		context = contextName
	}
	callback := cAdapterName(reference.Function)
	if function, ok := generator.functions[reference.Function]; ok && function.Lambda {
		callback = cFunctionName(reference.Function)
	}
	fmt.Fprintf(&generator.builder, "  qwic_closure *%s = qwic_closure_new(%s, %s);\n", cValue(reference.Target), callback, context)
}

func (generator *cGenerator) callSignature(name string) (types.Type, []types.Type, bool) {
	if function, ok := generator.functions[name]; ok {
		parameters := make([]types.Type, 0, len(function.Parameters))
		for _, parameter := range function.Parameters {
			parameters = append(parameters, parameter.Type)
		}
		return function.ReturnType, parameters, true
	}
	if function, ok := stdlib.LookupFunction(name); ok {
		parameters := make([]types.Type, 0, len(function.Parameters))
		for _, parameter := range function.Parameters {
			parameters = append(parameters, parameter.Type)
		}
		return function.ReturnType, parameters, true
	}
	return types.InvalidType, nil, false
}

func (generator *cGenerator) fieldType(object types.Type, field string) types.Type {
	definition, ok := generator.typeDefinitions[object.Name]
	if !ok {
		return types.AnyType
	}
	for _, candidate := range definition.Fields {
		if candidate.Name == field {
			return candidate.Type
		}
	}
	return types.AnyType
}

func (generator *cGenerator) nextSynthetic(kind string) string {
	generator.syntheticIndex++
	return fmt.Sprintf("qw_%s_%d", kind, generator.syntheticIndex)
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
		fmt.Fprintf(&generator.builder, "  %s = %s;\n", cIdentifier(node.Target), generator.adaptExpression(cValue(node.Value), generator.valueTypes[node.Value], generator.valueTypes[node.Target]))
	case *ir.BinaryOperation:
		generator.valueTypes[node.Target] = node.Type
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(node.Type), cValue(node.Target), generator.cBinaryExpression(node))
	case *ir.Call:
		generator.writeCall(node)
	case *ir.FunctionReference:
		generator.writeFunctionReference(node)
	case *ir.IndirectCall:
		generator.writeIndirectCall(node)
	case *ir.NewStruct:
		generator.valueTypes[node.Target] = node.Type
		fmt.Fprintf(&generator.builder, "  %s %s = (%s)qwic_alloc(sizeof(%s));\n", cType(node.Type), cValue(node.Target), cType(node.Type), cStructName(node.Type.Name))
	case *ir.LoadField:
		generator.valueTypes[node.Target] = node.Type
		objectType := generator.valueTypes[node.Object]
		if objectType.Kind == types.Any {
			value := fmt.Sprintf("qwic_any_field(%s, %s)", cValue(node.Object), cStringLiteral(node.Field))
			fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(node.Type), cValue(node.Target), generator.adaptExpression(value, types.AnyType, node.Type))
		} else {
			fmt.Fprintf(&generator.builder, "  %s %s = %s->%s;\n", cType(node.Type), cValue(node.Target), cValue(node.Object), cIdentifier(node.Field))
		}
	case *ir.StoreField:
		objectType := generator.valueTypes[node.Object]
		if objectType.Kind == types.Any {
			value := generator.adaptExpression(cValue(node.Value), generator.valueTypes[node.Value], types.AnyType)
			fmt.Fprintf(&generator.builder, "  qwic_any_set_field(%s, %s, %s);\n", cValue(node.Object), cStringLiteral(node.Field), value)
		} else {
			fieldType := generator.fieldType(objectType, node.Field)
			fmt.Fprintf(&generator.builder, "  %s->%s = %s;\n", cValue(node.Object), cIdentifier(node.Field), generator.adaptExpression(cValue(node.Value), generator.valueTypes[node.Value], fieldType))
		}
	case *ir.TryBegin:
		fmt.Fprintf(&generator.builder, "  qwic_try_frame *%s = qwic_try_new();\n", cIdentifier(node.Frame))
		fmt.Fprintf(&generator.builder, "  qwic_try_push(%s);\n", cIdentifier(node.Frame))
		fmt.Fprintf(&generator.builder, "  if (setjmp(%s->environment) == 0) goto %s; else goto %s;\n", cIdentifier(node.Frame), cLabel(node.TryBlock), cLabel(node.CatchBlock))
	case *ir.TryEnd:
		fmt.Fprintf(&generator.builder, "  qwic_try_end(%s);\n", cIdentifier(node.Frame))
	case *ir.Catch:
		generator.valueTypes[node.Target] = types.StringType
		fmt.Fprintf(&generator.builder, "  const char *%s = qwic_try_message(%s);\n", cIdentifier(node.Target), cIdentifier(node.Frame))
		fmt.Fprintf(&generator.builder, "  qwic_try_caught(%s);\n", cIdentifier(node.Frame))
	case *ir.Throw:
		fmt.Fprintf(&generator.builder, "  qwic_throw(%s);\n", generator.adaptExpression(cValue(node.Value), generator.valueTypes[node.Value], types.StringType))
	case *ir.FormatString:
		generator.writeFormatString(node)
	case *ir.Return:
		generator.writeReturn(function, node)
	case *ir.Branch:
		condition := generator.adaptExpression(cValue(node.Condition), generator.valueTypes[node.Condition], types.BoolType)
		fmt.Fprintf(&generator.builder, "  if (%s) goto %s; else goto %s;\n", condition, cLabel(node.ThenBlock), cLabel(node.ElseBlock))
	case *ir.Jump:
		fmt.Fprintf(&generator.builder, "  goto %s;\n", cLabel(node.Target))
	default:
		generator.diagnostics = append(generator.diagnostics, diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("unsupported IR instruction %T", instruction)))
	}
}

func (generator *cGenerator) writeIndirectCall(call *ir.IndirectCall) {
	argumentsName := generator.nextSynthetic("args")
	fmt.Fprintf(&generator.builder, "  qwic_value *%s[%d];\n", argumentsName, max(1, len(call.Args)))
	for index, argument := range call.Args {
		fmt.Fprintf(&generator.builder, "  %s[%d] = %s;\n", argumentsName, index, generator.adaptExpression(cValue(argument), generator.valueTypes[argument], types.AnyType))
	}
	callee := generator.adaptExpression(cValue(call.Callee), generator.valueTypes[call.Callee], types.FunctionType(nil, types.AnyType))
	resultName := generator.nextSynthetic("result")
	fmt.Fprintf(&generator.builder, "  qwic_value *%s = qwic_closure_call(%s, %s, %d);\n", resultName, callee, argumentsName, len(call.Args))
	if call.Target != "" {
		generator.valueTypes[call.Target] = call.Type
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(call.Type), cValue(call.Target), generator.adaptExpression(resultName, types.AnyType, call.Type))
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

	returnType, parameters, ok := generator.callSignature(call.Function)
	if !ok {
		generator.diagnostics = append(generator.diagnostics, diagnostic.Error(token.Position{Line: 1, Column: 1}, fmt.Sprintf("unknown function %q during C generation", call.Function)))
		return
	}

	var expression strings.Builder
	fmt.Fprintf(&expression, "%s(", cFunctionName(call.Function))
	for i, arg := range call.Args {
		if i > 0 {
			expression.WriteString(", ")
		}
		targetType := generator.valueTypes[arg]
		if i < len(parameters) {
			targetType = parameters[i]
		}
		expression.WriteString(generator.adaptExpression(cValue(arg), generator.valueTypes[arg], targetType))
	}
	expression.WriteString(")")
	if call.Target != "" {
		generator.valueTypes[call.Target] = call.Type
		value := generator.adaptExpression(expression.String(), returnType, call.Type)
		fmt.Fprintf(&generator.builder, "  %s %s = %s;\n", cType(call.Type), cValue(call.Target), value)
	} else {
		fmt.Fprintf(&generator.builder, "  %s;\n", expression.String())
	}
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
	case types.Any:
		return "%s", fmt.Sprintf("qwic_any_to_string(%s)", cValue(part.Value))
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
	case types.List:
		fmt.Fprintf(&generator.builder, "  qwic_print_list(%s);\n", cValue(arg))
	case types.Set:
		fmt.Fprintf(&generator.builder, "  qwic_print_set(%s);\n", cValue(arg))
	case types.Dictionary:
		fmt.Fprintf(&generator.builder, "  qwic_print_dictionary(%s);\n", cValue(arg))
	case types.Tuple:
		fmt.Fprintf(&generator.builder, "  qwic_print_tuple(%s);\n", cValue(arg))
	case types.Any:
		fmt.Fprintf(&generator.builder, "  qwic_print_any(%s);\n", cValue(arg))
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
		if function.Lambda {
			generator.builder.WriteString("  return qwic_any_null();\n")
		} else {
			generator.builder.WriteString("  return;\n")
		}
		return
	}
	returnType := function.ReturnType
	if function.Lambda {
		returnType = types.AnyType
	}
	fmt.Fprintf(&generator.builder, "  return %s;\n", generator.adaptExpression(cValue(instruction.Value), generator.valueTypes[instruction.Value], returnType))
}

func (generator *cGenerator) cBinaryExpression(instruction *ir.BinaryOperation) string {
	leftType := generator.valueTypes[instruction.Left]
	rightType := generator.valueTypes[instruction.Right]
	if instruction.Operator == token.Plus && leftType.Kind == types.String && rightType.Kind == types.String {
		return fmt.Sprintf("qwic_strings_concat(%s, %s)", cValue(instruction.Left), cValue(instruction.Right))
	}
	if leftType.Kind == types.String && rightType.Kind == types.String {
		comparison := fmt.Sprintf("strcmp(%s, %s)", cValue(instruction.Left), cValue(instruction.Right))
		switch instruction.Operator {
		case token.Equal:
			return comparison + " == 0"
		case token.NotEqual:
			return comparison + " != 0"
		case token.Less:
			return comparison + " < 0"
		case token.LessEqual:
			return comparison + " <= 0"
		case token.Greater:
			return comparison + " > 0"
		case token.GreaterEqual:
			return comparison + " >= 0"
		}
	}
	if instruction.Operator == token.Plus && leftType.Kind == types.List && rightType.Kind == types.List {
		return fmt.Sprintf("qwic_lists_concat(%s, %s)", cValue(instruction.Left), cValue(instruction.Right))
	}
	if leftType.Kind == types.Any || rightType.Kind == types.Any {
		left := generator.adaptExpression(cValue(instruction.Left), leftType, types.AnyType)
		right := generator.adaptExpression(cValue(instruction.Right), rightType, types.AnyType)
		switch instruction.Operator {
		case token.Plus:
			value := fmt.Sprintf("qwic_any_add(%s, %s)", left, right)
			return generator.adaptExpression(value, types.AnyType, instruction.Type)
		case token.Equal:
			return fmt.Sprintf("qwic_any_equal(%s, %s)", left, right)
		case token.NotEqual:
			return fmt.Sprintf("!qwic_any_equal(%s, %s)", left, right)
		case token.Less, token.LessEqual, token.Greater, token.GreaterEqual:
			operator := map[token.Kind]string{token.Less: "<", token.LessEqual: "<=", token.Greater: ">", token.GreaterEqual: ">="}[instruction.Operator]
			return fmt.Sprintf("qwic_any_compare(%s, %s) %s 0", left, right, operator)
		}
	}
	if instruction.Operator == token.Not {
		return fmt.Sprintf("!%s", generator.adaptExpression(cValue(instruction.Left), leftType, types.BoolType))
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

func (generator *cGenerator) adaptExpression(expression string, source, target types.Type) string {
	if sameType(source, target) || source.Kind == types.Invalid || target.Kind == types.Invalid {
		return expression
	}
	if target.Kind == types.Any {
		switch source.Kind {
		case types.Null:
			return "qwic_any_null()"
		case types.Bool:
			return fmt.Sprintf("qwic_any_bool(%s)", expression)
		case types.Int:
			return fmt.Sprintf("qwic_any_int(%s)", expression)
		case types.Float, types.Nano:
			return fmt.Sprintf("qwic_any_float(%s)", expression)
		case types.String:
			return fmt.Sprintf("qwic_any_string(%s)", expression)
		case types.List:
			return fmt.Sprintf("qwic_any_list(%s)", expression)
		case types.Set:
			return fmt.Sprintf("qwic_any_set(%s)", expression)
		case types.Dictionary:
			return fmt.Sprintf("qwic_any_dictionary(%s)", expression)
		case types.Tuple:
			return fmt.Sprintf("qwic_any_tuple(%s)", expression)
		case types.Function:
			return fmt.Sprintf("qwic_any_function(%s)", expression)
		case types.Struct:
			return fmt.Sprintf("%s(%s)", cToAnyName(source.Name), expression)
		default:
			return fmt.Sprintf("qwic_any_pointer(%s)", expression)
		}
	}
	if source.Kind == types.Any || source.Kind == types.Null {
		switch target.Kind {
		case types.Bool:
			return fmt.Sprintf("qwic_any_as_bool(%s)", expression)
		case types.Int:
			return fmt.Sprintf("qwic_any_as_int(%s)", expression)
		case types.Float, types.Nano:
			return fmt.Sprintf("qwic_any_as_float(%s)", expression)
		case types.String:
			return fmt.Sprintf("qwic_any_as_string(%s)", expression)
		case types.List:
			return fmt.Sprintf("qwic_any_as_list(%s)", expression)
		case types.Set:
			return fmt.Sprintf("qwic_any_as_set(%s)", expression)
		case types.Dictionary:
			return fmt.Sprintf("qwic_any_as_dictionary(%s)", expression)
		case types.Tuple:
			return fmt.Sprintf("qwic_any_as_tuple(%s)", expression)
		case types.Function:
			return fmt.Sprintf("qwic_any_as_function(%s)", expression)
		case types.Struct:
			return fmt.Sprintf("%s(%s)", cFromAnyName(target.Name), expression)
		default:
			return fmt.Sprintf("qwic_any_as_pointer(%s)", expression)
		}
	}
	return expression
}

func sameType(left, right types.Type) bool {
	if left.Kind != right.Kind || left.Name != right.Name || len(left.Parameters) != len(right.Parameters) {
		return false
	}
	for index := range left.Parameters {
		if !sameType(left.Parameters[index], right.Parameters[index]) {
			return false
		}
	}
	if left.ReturnType == nil || right.ReturnType == nil {
		return left.ReturnType == nil && right.ReturnType == nil
	}
	return sameType(*left.ReturnType, *right.ReturnType)
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
	case types.Struct:
		return cStructName(typ.Name) + " *"
	case types.Function:
		return "qwic_closure *"
	case types.Any, types.Null:
		return "qwic_value *"
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
		return "qwic_any_null()"
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
	case types.List, types.Set, types.Dictionary, types.Tuple, types.Struct, types.Function:
		return "NULL"
	case types.Any, types.Null:
		return "qwic_any_null()"
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

func cStructName(name string) string {
	return sanitizeCName("qw_type_" + name)
}

func cAdapterName(name string) string {
	return sanitizeCName("qw_adapter_" + name)
}

func cCaptureTypeName(name string) string {
	return sanitizeCName("qw_captures_" + name)
}

func cToAnyName(name string) string {
	return sanitizeCName("qw_to_any_" + name)
}

func cFromAnyName(name string) string {
	return sanitizeCName("qw_from_any_" + name)
}

func cLabel(name string) string {
	return sanitizeCName("label_" + name)
}

var invalidCName = regexp.MustCompile(`[^A-Za-z0-9_]`)

var cKeywords = map[string]bool{
	"auto": true, "break": true, "case": true, "char": true, "const": true,
	"continue": true, "default": true, "do": true, "double": true, "else": true,
	"enum": true, "extern": true, "float": true, "for": true, "goto": true,
	"if": true, "inline": true, "int": true, "long": true, "register": true,
	"restrict": true, "return": true, "short": true, "signed": true, "sizeof": true,
	"static": true, "struct": true, "switch": true, "typedef": true, "union": true,
	"unsigned": true, "void": true, "volatile": true, "while": true,
	"_Alignas": true, "_Alignof": true, "_Atomic": true, "_Bool": true,
	"_Complex": true, "_Generic": true, "_Imaginary": true, "_Noreturn": true,
	"_Static_assert": true, "_Thread_local": true,
}

func sanitizeCName(name string) string {
	name = invalidCName.ReplaceAllString(name, "_")
	if name == "" {
		return "qw_empty"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "qw_" + name
	}
	if cKeywords[name] {
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
	case *ir.Return, *ir.Branch, *ir.Jump, *ir.TryBegin, *ir.Throw:
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
