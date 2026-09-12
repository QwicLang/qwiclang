package lsp

import (
	"fmt"
	"strings"
	"unicode"

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

const (
	CompletionItemKindFunction = 3
	CompletionItemKindVariable = 6
	CompletionItemKindModule   = 9
	CompletionItemKindKeyword  = 14
	CompletionItemKindType     = 25
)

func (s *Server) complete(uri string, pos Position) []CompletionItem {
	val, ok := s.documents.Load(uri)
	if !ok {
		return []CompletionItem{}
	}
	content := val.(string)

	prefix, isDotTrigger, packageName := getWordBeforePosition(content, pos.Line, pos.Character)

	var items []CompletionItem

	// 1. If triggered by '.', offer stdlib package functions (e.g. math. or strings.)
	if isDotTrigger && packageName != "" {
		for _, fn := range stdlib.Functions() {
			if fn.Package == packageName {
				doc := formatFunctionSignature(fn.Package, fn.Name, fn.Parameters, fn.ReturnType.String())
				items = append(items, CompletionItem{
					Label:         fn.Name,
					Kind:          CompletionItemKindFunction,
					Detail:        &doc,
					Documentation: doc,
				})
			}
		}
		return items
	}

	// 2. Keywords
	for _, kw := range keywords {
		if prefix == "" || strings.HasPrefix(kw, prefix) {
			items = append(items, CompletionItem{
				Label: kw,
				Kind:  CompletionItemKindKeyword,
			})
		}
	}

	// 3. Primitive types
	for _, typ := range primitiveTypes {
		if prefix == "" || strings.HasPrefix(typ, prefix) {
			items = append(items, CompletionItem{
				Label: typ,
				Kind:  CompletionItemKindType,
			})
		}
	}

	// 4. Builtins (e.g., print)
	printDetail := "print(value: any): void"
	items = append(items, CompletionItem{
		Label:  "print",
		Kind:   CompletionItemKindFunction,
		Detail: &printDetail,
	})

	// 5. User-declared functions in current file
	program, _ := parser.Parse(uri, content)
	if program != nil {
		for _, decl := range program.Declarations {
			if fn, ok := decl.(*ast.FunctionDeclaration); ok {
				if prefix == "" || strings.HasPrefix(fn.Name, prefix) {
					detail := fmt.Sprintf("func %s(...): %s", fn.Name, fn.ReturnType)
					items = append(items, CompletionItem{
						Label:  fn.Name,
						Kind:   CompletionItemKindFunction,
						Detail: &detail,
					})
				}
			}
		}
	}

	// 6. Stdlib packages (e.g., math, strings, time, http, json)
	seenPackages := map[string]bool{}
	for _, fn := range stdlib.Functions() {
		if !seenPackages[fn.Package] {
			seenPackages[fn.Package] = true
			if prefix == "" || strings.HasPrefix(fn.Package, prefix) {
				detail := fmt.Sprintf("package %s", fn.Package)
				items = append(items, CompletionItem{
					Label:  fn.Package,
					Kind:   CompletionItemKindModule,
					Detail: &detail,
				})
			}
		}
	}

	return items
}

func (s *Server) hover(uri string, pos Position) *Hover {
	val, ok := s.documents.Load(uri)
	if !ok {
		return nil
	}
	content := val.(string)

	word, fullQualified := getIdentifierAtPosition(content, pos.Line, pos.Character)
	if word == "" && fullQualified == "" {
		return nil
	}

	// Check if cursor is on a stdlib function (e.g. math.abs or abs when qualified)
	target := fullQualified
	if target == "" {
		target = word
	}

	if fn, ok := stdlib.LookupFunction(target); ok {
		sig := formatFunctionSignature(fn.Package, fn.Name, fn.Parameters, fn.ReturnType.String())
		markdown := fmt.Sprintf("```qwic\n%s\n```\n*Qwic standard library*", sig)
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: markdown,
			},
		}
	}

	// Check keywords
	for _, kw := range keywords {
		if kw == word {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("**keyword** `%s`", kw),
				},
			}
		}
	}

	// Check primitive types
	for _, typ := range primitiveTypes {
		if typ == word {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("**type** `%s`", typ),
				},
			}
		}
	}

	// Check builtins
	if word == "print" {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: "```qwic\nprint(value: any): void\n```\nPrints a value to standard output with a trailing newline.",
			},
		}
	}

	// Check AST declarations in current file
	checkResult, _ := sema.Check(uri, content)
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
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("```qwic\n%s\n```", sig),
				},
			}
		}
	}

	return nil
}

func (s *Server) definition(uri string, pos Position) *Location {
	val, ok := s.documents.Load(uri)
	if !ok {
		return nil
	}
	content := val.(string)

	word, _ := getIdentifierAtPosition(content, pos.Line, pos.Character)
	if word == "" {
		return nil
	}

	program, _ := parser.Parse(uri, content)
	if program == nil {
		return nil
	}

	// 1. Check if word matches a top-level function declaration
	for _, decl := range program.Declarations {
		if fn, ok := decl.(*ast.FunctionDeclaration); ok {
			if fn.Name == word {
				loc := makeLocation(uri, fn.Position())
				return &loc
			}
			// Check function parameters
			for _, param := range fn.Parameters {
				if param.Name == word {
					loc := makeLocation(uri, param.Pos)
					return &loc
				}
			}
		}
	}

	return nil
}

func makeLocation(uri string, pos token.Position) Location {
	line := uint32(0)
	if pos.Line > 0 {
		line = uint32(pos.Line - 1)
	}
	col := uint32(0)
	if pos.Column > 0 {
		col = uint32(pos.Column - 1)
	}

	return Location{
		URI: uri,
		Range: Range{
			Start: Position{Line: line, Character: col},
			End:   Position{Line: line, Character: col},
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

	start := col
	for start > 0 && isIdentChar(rune(l[start-1])) {
		start--
	}
	end := col
	for end < len(l) && isIdentChar(rune(l[end])) {
		end++
	}
	word = l[start:end]

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
