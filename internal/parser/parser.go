package parser

import (
	"strings"

	"qwiclang/internal/ast"
	"qwiclang/internal/diagnostic"
	"qwiclang/internal/lexer"
	"qwiclang/internal/token"
)

type Diagnostic = diagnostic.Diagnostic

type Parser struct {
	tokens      []token.Token
	current     int
	diagnostics []Diagnostic
}

func Parse(filename, source string) (*ast.Program, []Diagnostic) {
	tokens, lexerErrors := lexer.Lex(filename, source)
	diagnostics := make([]Diagnostic, 0, len(lexerErrors))
	for _, err := range lexerErrors {
		diagnostics = append(diagnostics, err)
	}
	if len(diagnostics) > 0 {
		return &ast.Program{}, diagnostics
	}
	return New(tokens).ParseProgram()
}

func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (parser *Parser) ParseProgram() (*ast.Program, []Diagnostic) {
	program := &ast.Program{}
	parser.skipStatementBoundaries()

	for !parser.isAtEnd() {
		declaration := parser.parseDeclaration()
		if declaration != nil {
			program.Declarations = append(program.Declarations, declaration)
		}
		parser.skipStatementBoundaries()
	}

	return program, parser.diagnostics
}

func (parser *Parser) parseDeclaration() ast.Declaration {
	if parser.match(token.Module) {
		return parser.parseModuleDeclaration(parser.previous().Start)
	}
	if parser.match(token.Import) {
		return parser.parseImportDeclaration(parser.previous().Start)
	}

	visibility := ast.VisibilityPrivate
	turbo := false
	start := parser.peek().Start

	for {
		switch {
		case parser.match(token.Public):
			visibility = ast.VisibilityPublic
			start = parser.previous().Start
		case parser.match(token.Private):
			visibility = ast.VisibilityPrivate
			start = parser.previous().Start
		case parser.match(token.Turbo):
			turbo = true
			start = parser.previous().Start
		default:
			if parser.match(token.Func) {
				return parser.parseFunctionDeclaration(start, visibility, turbo)
			}
			parser.errorAtCurrent("expected function declaration")
			parser.synchronizeTopLevel()
			return nil
		}
	}
}

func (parser *Parser) parseModuleDeclaration(start token.Position) ast.Declaration {
	name, ok := parser.consumeIdentifier("expected module name")
	if !ok {
		parser.synchronizeTopLevel()
		return nil
	}
	parser.consumeOptionalStatementEnd()
	return &ast.ModuleDeclaration{Name: name.Lexeme, Pos: start}
}

func (parser *Parser) parseImportDeclaration(start token.Position) ast.Declaration {
	name, ok := parser.consumeIdentifier("expected import name")
	if !ok {
		parser.synchronizeTopLevel()
		return nil
	}
	parser.consumeOptionalStatementEnd()
	return &ast.ImportDeclaration{Name: name.Lexeme, Pos: start}
}

func (parser *Parser) parseFunctionDeclaration(start token.Position, visibility ast.Visibility, turbo bool) ast.Declaration {
	name, ok := parser.consumeIdentifier("expected function name")
	if !ok {
		parser.synchronizeTopLevel()
		return nil
	}

	if !parser.consume(token.LParen, "expected '(' after function name") {
		parser.synchronizeTopLevel()
		return nil
	}

	parameters := parser.parseParameters()
	if !parser.consume(token.RParen, "expected ')' after function parameters") {
		parser.synchronizeTopLevel()
		return nil
	}

	returnType := "void"
	if parser.match(token.Colon) {
		typeName, ok := parser.consumeIdentifier("expected return type")
		if !ok {
			parser.synchronizeTopLevel()
			return nil
		}
		returnType = typeName.Lexeme
	}

	body := parser.parseBlock()
	if body == nil {
		parser.synchronizeTopLevel()
		return nil
	}

	return &ast.FunctionDeclaration{
		Name:       name.Lexeme,
		Visibility: visibility,
		Turbo:      turbo,
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body,
		Pos:        start,
	}
}

func (parser *Parser) parseParameters() []ast.Parameter {
	var parameters []ast.Parameter
	if parser.check(token.RParen) {
		return parameters
	}

	for {
		name, ok := parser.consumeIdentifier("expected parameter name")
		if !ok {
			return parameters
		}
		if !parser.consume(token.Colon, "expected ':' after parameter name") {
			return parameters
		}
		typeName, ok := parser.consumeIdentifier("expected parameter type")
		if !ok {
			return parameters
		}
		parameters = append(parameters, ast.Parameter{
			Name:     name.Lexeme,
			TypeName: typeName.Lexeme,
			Pos:      name.Start,
		})

		if !parser.match(token.Comma) {
			break
		}
	}

	return parameters
}

func (parser *Parser) parseBlock() *ast.BlockStatement {
	if !parser.consume(token.LBrace, "expected '{' before block") {
		return nil
	}
	start := parser.previous().Start
	block := &ast.BlockStatement{Pos: start}
	parser.skipStatementBoundaries()

	for !parser.check(token.RBrace) && !parser.isAtEnd() {
		statement := parser.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
		parser.skipStatementBoundaries()
	}

	parser.consume(token.RBrace, "expected '}' after block")
	return block
}

func (parser *Parser) parseStatement() ast.Statement {
	switch {
	case parser.match(token.Const):
		return parser.parseVariableDeclaration(false, parser.previous().Start)
	case parser.match(token.Let):
		return parser.parseVariableDeclaration(true, parser.previous().Start)
	case parser.match(token.Return):
		return parser.parseReturnStatement(parser.previous().Start)
	case parser.match(token.If):
		return parser.parseIfStatement(parser.previous().Start)
	case parser.match(token.While):
		return parser.parseWhileStatement(parser.previous().Start)
	case parser.match(token.For):
		return parser.parseForStatement(parser.previous().Start)
	default:
		return parser.parseAssignmentOrExpressionStatement()
	}
}

func (parser *Parser) parseVariableDeclaration(mutable bool, start token.Position) ast.Statement {
	name, ok := parser.consumeIdentifier("expected variable name")
	if !ok {
		parser.synchronizeStatement()
		return nil
	}

	typeName := ""
	if parser.match(token.Colon) {
		typeToken, ok := parser.consumeIdentifier("expected variable type")
		if !ok {
			parser.synchronizeStatement()
			return nil
		}
		typeName = typeToken.Lexeme
	}

	if !parser.consume(token.Assign, "expected '=' in variable declaration") {
		parser.synchronizeStatement()
		return nil
	}

	value := parser.parseExpression()
	parser.consumeOptionalStatementEnd()
	return &ast.VariableDeclaration{
		Mutable:  mutable,
		Name:     name.Lexeme,
		TypeName: typeName,
		Value:    value,
		Pos:      start,
	}
}

func (parser *Parser) parseReturnStatement(start token.Position) ast.Statement {
	if parser.isStatementEnd() {
		parser.consumeOptionalStatementEnd()
		return &ast.ReturnStatement{Pos: start}
	}

	value := parser.parseExpression()
	parser.consumeOptionalStatementEnd()
	return &ast.ReturnStatement{Value: value, Pos: start}
}

func (parser *Parser) parseIfStatement(start token.Position) ast.Statement {
	condition := parser.parseExpression()
	thenBranch := parser.parseBlock()
	if thenBranch == nil {
		parser.synchronizeStatement()
		return nil
	}

	var elseBranch *ast.BlockStatement
	if parser.match(token.Else) {
		elseBranch = parser.parseBlock()
		if elseBranch == nil {
			parser.synchronizeStatement()
			return nil
		}
	}

	return &ast.IfStatement{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
		Pos:        start,
	}
}

func (parser *Parser) parseWhileStatement(start token.Position) ast.Statement {
	condition := parser.parseExpression()
	body := parser.parseBlock()
	if body == nil {
		parser.synchronizeStatement()
		return nil
	}
	return &ast.WhileStatement{
		Condition: condition,
		Body:      body,
		Pos:       start,
	}
}

// parseForStatement parses: for variable in iterable { ... }
func (parser *Parser) parseForStatement(start token.Position) ast.Statement {
	variable, ok := parser.consumeIdentifier("expected variable name in for loop")
	if !ok {
		parser.synchronizeStatement()
		return nil
	}

	if !parser.match(token.In) {
		parser.errorAtCurrent("expected 'in' in for loop")
		parser.synchronizeStatement()
		return nil
	}

	iterable := parser.parseExpression()
	body := parser.parseBlock()
	if body == nil {
		parser.synchronizeStatement()
		return nil
	}

	return &ast.ForStatement{
		Variable: variable.Lexeme,
		Iterable: iterable,
		Body:     body,
		Pos:      start,
	}
}

func (parser *Parser) parseAssignmentOrExpressionStatement() ast.Statement {
	start := parser.peek().Start
	if parser.check(token.Identifier) && parser.checkNext(token.Assign) {
		name := parser.advance()
		parser.advance()
		value := parser.parseExpression()
		parser.consumeOptionalStatementEnd()
		return &ast.AssignmentStatement{
			Name:  name.Lexeme,
			Value: value,
			Pos:   start,
		}
	}

	expression := parser.parseExpression()
	parser.consumeOptionalStatementEnd()
	return &ast.ExpressionStatement{
		Expression: expression,
		Pos:        start,
	}
}

func (parser *Parser) parseExpression() ast.Expression {
	return parser.parseBinaryExpression(1)
}

func (parser *Parser) parseBinaryExpression(minPrecedence int) ast.Expression {
	left := parser.parseUnaryExpression()

	for {
		operator := parser.peek()
		precedence := binaryPrecedence(operator.Kind)
		if precedence < minPrecedence {
			break
		}

		parser.advance()
		right := parser.parseBinaryExpression(precedence + 1)
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: operator.Kind,
			Right:    right,
			Pos:      operator.Start,
		}
	}

	return left
}

func (parser *Parser) parseUnaryExpression() ast.Expression {
	if parser.match(token.Not, token.Minus) {
		operator := parser.previous()
		return &ast.UnaryExpression{
			Operator: operator.Kind,
			Right:    parser.parseUnaryExpression(),
			Pos:      operator.Start,
		}
	}

	return parser.parseCallExpression()
}

func (parser *Parser) parseCallExpression() ast.Expression {
	expression := parser.parsePrimaryExpression()

	for {
		// Index access: collection[index]
		if parser.match(token.LBracket) {
			start := parser.previous().Start
			index := parser.parseExpression()
			parser.consume(token.RBracket, "expected ']' after index")
			expression = &ast.IndexExpression{
				Left:  expression,
				Index: index,
				Pos:   start,
			}
			continue
		}

		// Dot selector
		if parser.match(token.Dot) {
			name, ok := parser.consumeIdentifier("expected selector name after '.'")
			if !ok {
				return expression
			}
			expression = &ast.SelectorExpression{
				Left: expression,
				Name: name.Lexeme,
				Pos:  name.Start,
			}
			continue
		}

		// Function call
		if !parser.match(token.LParen) {
			break
		}
		start := parser.previous().Start
		call := &ast.CallExpression{
			Callee: expression,
			Pos:    start,
		}

		if !parser.check(token.RParen) {
			for {
				call.Arguments = append(call.Arguments, parser.parseExpression())
				if !parser.match(token.Comma) {
					break
				}
			}
		}

		parser.consume(token.RParen, "expected ')' after arguments")
		expression = call
	}

	return expression
}

func (parser *Parser) parsePrimaryExpression() ast.Expression {
	current := parser.peek()
	switch {
	case parser.match(token.Identifier):
		return &ast.IdentifierExpression{Name: current.Lexeme, Pos: current.Start}
	case parser.match(token.Integer):
		return &ast.LiteralExpression{Kind: ast.LiteralInteger, Value: current.Lexeme, Pos: current.Start}
	case parser.match(token.Float):
		return &ast.LiteralExpression{Kind: ast.LiteralFloat, Value: current.Lexeme, Pos: current.Start}
	case parser.match(token.String):
		return &ast.LiteralExpression{Kind: ast.LiteralString, Value: current.Lexeme, Pos: current.Start}
	case parser.match(token.FString):
		return parser.parseInterpolatedString(current)
	case parser.match(token.True, token.False):
		return &ast.LiteralExpression{Kind: ast.LiteralBool, Value: current.Lexeme, Pos: current.Start}
	case parser.match(token.Null):
		return &ast.LiteralExpression{Kind: ast.LiteralNull, Value: current.Lexeme, Pos: current.Start}
	case parser.match(token.LBracket):
		return parser.parseArrayLiteral(current.Start)
	case parser.match(token.LBrace):
		return parser.parseDictionaryLiteral(current.Start)
	case parser.match(token.LParen):
		return parser.parseTupleLiteral(current.Start)
	default:
		parser.errorAtCurrent("expected expression")
		if !parser.isAtEnd() {
			parser.advance()
		}
		return &ast.IdentifierExpression{Name: "<error>", Pos: current.Start}
	}
}

// parseArrayLiteral parses: [1, 2, "a", "b"]
func (parser *Parser) parseArrayLiteral(start token.Position) ast.Expression {
	var elements []ast.Expression

	if !parser.check(token.RBracket) {
		for {
			elements = append(elements, parser.parseExpression())
			if !parser.match(token.Comma) {
				break
			}
		}
	}

	parser.consume(token.RBracket, "expected ']' after array elements")
	return &ast.ArrayLiteralExpression{Elements: elements, Pos: start}
}

// parseDictionaryLiteral parses: {"a": 1, "b": 2}
func (parser *Parser) parseDictionaryLiteral(start token.Position) ast.Expression {
	var pairs []ast.DictionaryPair

	if !parser.check(token.RBrace) {
		for {
			key := parser.parseExpression()
			parser.consume(token.Colon, "expected ':' after dictionary key")
			value := parser.parseExpression()
			pairs = append(pairs, ast.DictionaryPair{Key: key, Value: value})

			if !parser.match(token.Comma) {
				break
			}
		}
	}

	parser.consume(token.RBrace, "expected '}' after dictionary pairs")
	return &ast.DictionaryLiteralExpression{Pairs: pairs, Pos: start}
}

// parseTupleLiteral parses: (1, 2, 3, 4)
// Distinguishes between (expr) which is a grouped expression and (expr, ...) which is a tuple
func (parser *Parser) parseTupleLiteral(start token.Position) ast.Expression {
	var elements []ast.Expression

	if !parser.check(token.RParen) {
		elements = append(elements, parser.parseExpression())

		// If only one element and no comma, it's a grouped expression
		if parser.check(token.RParen) {
			parser.consume(token.RParen, "expected ')' after expression")
			return elements[0]
		}

		// If comma follows, it's a tuple
		if !parser.match(token.Comma) {
			parser.errorAtCurrent("expected ',' or ')' in tuple")
			parser.consume(token.RParen, "expected ')' after tuple element")
			return &ast.TupleLiteralExpression{Elements: elements, Pos: start}
		}

		// Parse remaining elements
		if !parser.check(token.RParen) {
			elements = append(elements, parser.parseExpression())
		}

		for parser.match(token.Comma) {
			if parser.check(token.RParen) {
				break
			}
			elements = append(elements, parser.parseExpression())
		}
	}

	parser.consume(token.RParen, "expected ')' after tuple elements")
	return &ast.TupleLiteralExpression{Elements: elements, Pos: start}
}

func (parser *Parser) parseInterpolatedString(tok token.Token) ast.Expression {
	body := strings.TrimSuffix(strings.TrimPrefix(tok.Lexeme, `f"`), `"`)
	parts, diagnostics := parseInterpolatedStringParts(tok.Start, body)
	parser.diagnostics = append(parser.diagnostics, diagnostics...)
	return &ast.InterpolatedStringExpression{Parts: parts, Pos: tok.Start}
}

func parseInterpolatedStringParts(start token.Position, body string) ([]ast.InterpolatedStringPart, []Diagnostic) {
	var parts []ast.InterpolatedStringPart
	var diagnostics []Diagnostic
	var text strings.Builder

	flushText := func() {
		if text.Len() == 0 {
			return
		}
		parts = append(parts, ast.InterpolatedStringPart{Text: text.String()})
		text.Reset()
	}

	for offset := 0; offset < len(body); offset++ {
		switch body[offset] {
		case '\\':
			if offset+1 < len(body) {
				offset++
				text.WriteByte(decodedEscape(body[offset]))
			} else {
				text.WriteByte('\\')
			}
		case '{':
			if offset+1 < len(body) && body[offset+1] == '{' {
				text.WriteByte('{')
				offset++
				continue
			}
			flushText()
			end := findInterpolationEnd(body, offset+1)
			if end < 0 {
				diagnostics = append(diagnostics, diagnostic.Error(start, "unterminated f-string interpolation"))
				return parts, diagnostics
			}
			expressionSource := strings.TrimSpace(body[offset+1 : end])
			if expressionSource == "" {
				diagnostics = append(diagnostics, diagnostic.Error(start, "empty f-string interpolation"))
				offset = end
				continue
			}
			expression, expressionDiagnostics := parseExpressionFragment(start.Filename, expressionSource, start)
			diagnostics = append(diagnostics, expressionDiagnostics...)
			parts = append(parts, ast.InterpolatedStringPart{Expression: expression})
			offset = end
		case '}':
			if offset+1 < len(body) && body[offset+1] == '}' {
				text.WriteByte('}')
				offset++
				continue
			}
			diagnostics = append(diagnostics, diagnostic.Error(start, "single '}' is not allowed in f-string literal text"))
		default:
			text.WriteByte(body[offset])
		}
	}
	flushText()
	return parts, diagnostics
}

func decodedEscape(char byte) byte {
	switch char {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	default:
		return char
	}
}

func findInterpolationEnd(body string, offset int) int {
	for offset < len(body) {
		if body[offset] == '\\' {
			offset += 2
			continue
		}
		if body[offset] == '}' {
			return offset
		}
		offset++
	}
	return -1
}

func parseExpressionFragment(filename, source string, position token.Position) (ast.Expression, []Diagnostic) {
	tokens, lexerDiagnostics := lexer.Lex(filename, source)
	diagnostics := make([]Diagnostic, 0, len(lexerDiagnostics)+1)
	for _, err := range lexerDiagnostics {
		diagnostics = append(diagnostics, err)
	}
	if len(diagnostics) > 0 {
		return &ast.IdentifierExpression{Name: "<error>", Pos: position}, diagnostics
	}
	parser := New(tokens)
	expression := parser.parseExpression()
	if !parser.check(token.EOF) {
		parser.errorAtCurrent("expected end of f-string interpolation")
	}
	diagnostics = append(diagnostics, parser.diagnostics...)
	if len(diagnostics) > 0 {
		return &ast.IdentifierExpression{Name: "<error>", Pos: position}, diagnostics
	}
	return expression, nil
}

func binaryPrecedence(kind token.Kind) int {
	switch kind {
	case token.Or:
		return 1
	case token.And:
		return 2
	case token.Equal, token.NotEqual:
		return 3
	case token.Less, token.LessEqual, token.Greater, token.GreaterEqual:
		return 4
	case token.Plus, token.Minus:
		return 5
	case token.Star, token.Slash, token.Percent:
		return 6
	default:
		return 0
	}
}

func (parser *Parser) consumeIdentifier(message string) (token.Token, bool) {
	if parser.check(token.Identifier) {
		return parser.advance(), true
	}
	parser.errorAtCurrent(message)
	return token.Token{}, false
}

func (parser *Parser) consume(kind token.Kind, message string) bool {
	if parser.check(kind) {
		parser.advance()
		return true
	}
	parser.errorAtCurrent(message)
	return false
}

func (parser *Parser) consumeOptionalStatementEnd() {
	if parser.match(token.Semicolon) {
		parser.match(token.Newline)
		return
	}
	if parser.match(token.Newline) {
		return
	}
	if parser.check(token.RBrace) || parser.check(token.EOF) {
		return
	}
	parser.errorAtCurrent("expected statement boundary")
	parser.synchronizeStatement()
}

func (parser *Parser) skipStatementBoundaries() {
	for parser.match(token.Newline, token.Semicolon) {
	}
}

func (parser *Parser) isStatementEnd() bool {
	return parser.check(token.Newline) || parser.check(token.Semicolon) || parser.check(token.RBrace) || parser.check(token.EOF)
}

func (parser *Parser) synchronizeStatement() {
	for !parser.isAtEnd() {
		if parser.match(token.Newline, token.Semicolon) {
			return
		}
		if parser.check(token.RBrace) {
			return
		}
		parser.advance()
	}
}

func (parser *Parser) synchronizeTopLevel() {
	for !parser.isAtEnd() {
		if parser.match(token.Newline, token.Semicolon) {
			return
		}
		if parser.check(token.Func, token.Public, token.Private, token.Turbo) {
			return
		}
		parser.advance()
	}
}

func (parser *Parser) match(kinds ...token.Kind) bool {
	for _, kind := range kinds {
		if parser.check(kind) {
			parser.advance()
			return true
		}
	}
	return false
}

func (parser *Parser) check(kinds ...token.Kind) bool {
	for _, kind := range kinds {
		if parser.peek().Kind == kind {
			return true
		}
	}
	return false
}

func (parser *Parser) checkNext(kind token.Kind) bool {
	if parser.current+1 >= len(parser.tokens) {
		return false
	}
	return parser.tokens[parser.current+1].Kind == kind
}

func (parser *Parser) advance() token.Token {
	if !parser.isAtEnd() {
		parser.current++
	}
	return parser.previous()
}

func (parser *Parser) isAtEnd() bool {
	return parser.peek().Kind == token.EOF
}

func (parser *Parser) peek() token.Token {
	if len(parser.tokens) == 0 {
		return token.Token{Kind: token.EOF, Start: token.Position{Line: 1, Column: 1}}
	}
	if parser.current >= len(parser.tokens) {
		return parser.tokens[len(parser.tokens)-1]
	}
	return parser.tokens[parser.current]
}

func (parser *Parser) previous() token.Token {
	if parser.current == 0 {
		return parser.peek()
	}
	return parser.tokens[parser.current-1]
}

func (parser *Parser) errorAtCurrent(message string) {
	parser.diagnostics = append(parser.diagnostics, diagnostic.Error(parser.peek().Start, message))
}
