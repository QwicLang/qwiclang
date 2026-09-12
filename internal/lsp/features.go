package lsp

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"qwiclang/internal/ast"
	"qwiclang/internal/parser"
	"qwiclang/internal/sema"
	"qwiclang/internal/stdlib"
	"qwiclang/internal/token"
)

var keywords = []string{
	"public", "private", "const", "let", "func", "turbo",
	"return", "if", "else", "while", "for", "in", "import", "module",
	"true", "false", "null", "struct",
}

var primitiveTypes = []string{
	"int", "float", "nano", "string", "bool", "void", "list", "set", "dictionary", "tuple",
}

func (s *Server) textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	val, ok := s.documents.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	content := val.(string)

	prefix, isDotTrigger, packageName := getWordBeforePosition(content, params.Position.Line, params.Position.Character)

	var items []protocol.CompletionItem

	// 1. If triggered by '.', offer stdlib package functions (e.g. math. or strings.)
	if isDotTrigger && packageName != "" {
		for _, fn := range stdlib.Functions() {
			if fn.Package == packageName {
				doc := formatFunctionSignature(fn.Package, fn.Name, fn.Parameters, fn.ReturnType.String())
				kind := protocol.CompletionItemKindFunction
				items = append(items, protocol.CompletionItem{
					Label:         fn.Name,
					Kind:          &kind,
					Detail:        &doc,
					Documentation: doc,
				})
			}
		}
		return items, nil
	}

	// 2. Keywords
	kwKind := protocol.CompletionItemKindKeyword
	for _, kw := range keywords {
		if prefix == "" || strings.HasPrefix(kw, prefix) {
			items = append(items, protocol.CompletionItem{
				Label: kw,
				Kind:  &kwKind,
			})
		}
	}

	// 3. Primitive types
	typeKind := protocol.CompletionItemKindTypeParameter
	for _, typ := range primitiveTypes {
		if prefix == "" || strings.HasPrefix(typ, prefix) {
			items = append(items, protocol.CompletionItem{
				Label: typ,
				Kind:  &typeKind,
			})
		}
	}

	// 4. Builtins (e.g., print)
	fnKind := protocol.CompletionItemKindFunction
	printDetail := "print(value: any): void"
	items = append(items, protocol.CompletionItem{
		Label:  "print",
		Kind:   &fnKind,
		Detail: &printDetail,
	})

	// 5. User-declared functions & variables in the current file
	program, _ := parser.Parse(params.TextDocument.URI, content)
	if program != nil {
		for _, decl := range program.Declarations {
			if fn, ok := decl.(*ast.FunctionDeclaration); ok {
				if prefix == "" || strings.HasPrefix(fn.Name, prefix) {
					detail := fmt.Sprintf("func %s(...): %s", fn.Name, fn.ReturnType)
					items = append(items, protocol.CompletionItem{
						Label:  fn.Name,
						Kind:   &fnKind,
						Detail: &detail,
					})
				}
			}
		}
	}

	// 6. Stdlib packages (e.g., math, strings, time, http, json)
	modKind := protocol.CompletionItemKindModule
	seenPackages := map[string]bool{}
	for _, fn := range stdlib.Functions() {
		if !seenPackages[fn.Package] {
			seenPackages[fn.Package] = true
			if prefix == "" || strings.HasPrefix(fn.Package, prefix) {
				detail := fmt.Sprintf("package %s", fn.Package)
				items = append(items, protocol.CompletionItem{
					Label:  fn.Package,
					Kind:   &modKind,
					Detail: &detail,
				})
			}
		}
	}

	return items, nil
}

func (s *Server) textDocumentHover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	val, ok := s.documents.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	content := val.(string)

	word, fullQualified := getIdentifierAtPosition(content, params.Position.Line, params.Position.Character)
	if word == "" && fullQualified == "" {
		return nil, nil
	}

	// Check if cursor is on a stdlib function (e.g. math.abs or abs when qualified)
	target := fullQualified
	if target == "" {
		target = word
	}

	if fn, ok := stdlib.LookupFunction(target); ok {
		sig := formatFunctionSignature(fn.Package, fn.Name, fn.Parameters, fn.ReturnType.String())
		markdown := fmt.Sprintf("```qwic\n%s\n```\n*Qwic standard library*", sig)
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: markdown,
			},
		}, nil
	}

	// Check keywords
	for _, kw := range keywords {
		if kw == word {
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: fmt.Sprintf("**keyword** `%s`", kw),
				},
			}, nil
		}
	}

	// Check primitive types
	for _, typ := range primitiveTypes {
		if typ == word {
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: fmt.Sprintf("**type** `%s`", typ),
				},
			}, nil
		}
	}

	// Check builtins
	if word == "print" {
		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: "```qwic\nprint(value: any): void\n```\nPrints a value to standard output with a trailing newline.",
			},
		}, nil
	}

	// Check AST declarations in current file
	checkResult, _ := sema.Check(params.TextDocument.URI, content)
	if checkResult != nil {
		if fnSymbol, ok := checkResult.Functions[word]; ok {
			paramsStr := make([]string, 0, len(fnSymbol.Parameters))
			for _, p := range fnSymbol.Parameters {
				paramsStr = append(paramsStr, fmt.Sprintf("%s: %s", p.Name, p.Type))
			}
			turboPrefix := ""
			if fnSymbol.Turbo {
				turboPrefix = "turbo "
			}
			visPrefix := ""
			if fnSymbol.Visibility == ast.VisibilityPublic {
				visPrefix = "public "
			}
			sig := fmt.Sprintf("%s%sfunc %s(%s): %s", visPrefix, turboPrefix, fnSymbol.Name, strings.Join(paramsStr, ", "), fnSymbol.ReturnType)
			return &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: fmt.Sprintf("```qwic\n%s\n```", sig),
				},
			}, nil
		}
	}

	return nil, nil
}

func (s *Server) textDocumentDefinition(context *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	val, ok := s.documents.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	content := val.(string)

	word, _ := getIdentifierAtPosition(content, params.Position.Line, params.Position.Character)
	if word == "" {
		return nil, nil
	}

	program, _ := parser.Parse(params.TextDocument.URI, content)
	if program == nil {
		return nil, nil
	}

	// 1. Check if word matches a top-level function declaration
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*ast.FunctionDeclaration); ok {
			if fn.Name == word {
				return makeLocation(params.TextDocument.URI, fn.Position()), nil
			}
			// Check function parameters
			for _, param := range fn.Parameters {
				if param.Name == word {
					return makeLocation(params.TextDocument.URI, param.Pos), nil
				}
			}
		}
	}

	return nil, nil
}

func makeLocation(uri protocol.DocumentUri, pos token.Position) protocol.Location {
	line := uint32(0)
	if pos.Line > 0 {
		line = uint32(pos.Line - 1)
	}
	col := uint32(0)
	if pos.Column > 0 {
		col = uint32(pos.Column - 1)
	}

	return protocol.Location{
		URI: uri,
		Range: protocol.Range{
			Start: protocol.Position{Line: line, Character: col},
			End:   protocol.Position{Line: line, Character: col},
		},
	}
}

func formatFunctionSignature(pkg, name string, params []stdlib.Parameter, returnType string) string {
	parts := make([]string, 0, len(params))
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("%s: %s", p.Name, p.Type))
	}
	if pkg != "" {
		return fmt.Sprintf("%s.%s(%s): %s", pkg, name, strings.Join(parts, ", "), returnType)
	}
	return fmt.Sprintf("%s(%s): %s", name, strings.Join(parts, ", "), returnType)
}

func getWordBeforePosition(content string, line uint32, char uint32) (prefix string, isDotTrigger bool, packageName string) {
	lines := strings.Split(content, "\n")
	if int(line) >= len(lines) {
		return "", false, ""
	}
	currentLine := lines[line]
	if int(char) > len(currentLine) {
		char = uint32(len(currentLine))
	}

	before := currentLine[:char]
	if strings.HasSuffix(before, ".") {
		// package.
		trim := strings.TrimSuffix(before, ".")
		pkg := getLastWord(trim)
		return "", true, pkg
	}

	word := getLastWord(before)
	return word, false, ""
}

func getLastWord(s string) string {
	end := len(s)
	for end > 0 && isIdentChar(rune(s[end-1])) {
		end--
	}
	return s[end:]
}

func getIdentifierAtPosition(content string, line uint32, char uint32) (word string, qualified string) {
	lines := strings.Split(content, "\n")
	if int(line) >= len(lines) {
		return "", ""
	}
	l := lines[line]
	if len(l) == 0 || int(char) > len(l) {
		return "", ""
	}

	col := int(char)
	if col == len(l) && col > 0 {
		col--
	}

	if !isIdentChar(rune(l[col])) && l[col] != '.' {
		if col > 0 && isIdentChar(rune(l[col-1])) {
			col--
		} else {
			return "", ""
		}
	}

	// Find start and end of current word
	start := col
	for start > 0 && isIdentChar(rune(l[start-1])) {
		start--
	}
	end := col
	for end < len(l) && isIdentChar(rune(l[end])) {
		end++
	}
	word = l[start:end]

	// Check if part of package.function (e.g. math.abs)
	qStart := start
	if qStart >= 2 && l[qStart-1] == '.' {
		pkgStart := qStart - 2
		for pkgStart > 0 && isIdentChar(rune(l[pkgStart-1])) {
			pkgStart--
		}
		qualified = l[pkgStart:end]
	} else if end < len(l) && l[end] == '.' {
		fnEnd := end + 1
		for fnEnd < len(l) && isIdentChar(rune(l[fnEnd])) {
			fnEnd++
		}
		qualified = l[start:fnEnd]
	}

	return word, qualified
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
