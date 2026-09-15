package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"qwiclang/internal/diagnostic"
	"qwiclang/internal/token"
)

type Error = diagnostic.Diagnostic
type Token = token.Token
type Diagnostic = Error

// Lexer converts Qwic source text into tokens without applying parser or
// semantic rules.
type Lexer struct {
	filename string
	source   string
	offset   int
	line     int
	column   int
	pending  []token.Token
}

// New creates a lexer for source. Filename is optional and is copied into token
// positions for diagnostics.
func New(filename, source string) *Lexer {
	return &Lexer{
		filename: filename,
		source:   source,
		line:     1,
		column:   1,
	}
}

// Lex tokenizes the complete input. It returns any tokens produced before a
// malformed token along with the lexing errors that stopped the scan.
func Lex(filename, source string) ([]token.Token, []Error) {
	return New(filename, source).Lex()
}

// Lex tokenizes the complete input.
func (lexer *Lexer) Lex() ([]token.Token, []Error) {
	var tokens []token.Token
	var errors []Error

	for {
		next, err := lexer.NextToken()
		if err != nil {
			errors = append(errors, *err)
			break
		}

		tokens = append(tokens, next)
		if next.Kind == token.EOF {
			break
		}
	}

	return tokens, errors
}

// NextToken returns the next token in the source.
func (lexer *Lexer) NextToken() (token.Token, *Error) {
	if len(lexer.pending) > 0 {
		next := lexer.pending[0]
		lexer.pending = lexer.pending[1:]
		return next, nil
	}

	lexer.skipHorizontalWhitespace()

	start := lexer.position()
	r, ok := lexer.peek()
	if !ok {
		return lexer.makeToken(token.EOF, "", start), nil
	}

	if r == '\n' || r == '\r' {
		lexeme := lexer.consumeNewline()
		return lexer.makeToken(token.Newline, lexeme, start), nil
	}

	if unicode.IsDigit(r) {
		return lexer.lexNumber(start), nil
	}

	if r == 'f' && lexer.peekAheadRune(1, '"') {
		lexer.advance()
		return lexer.lexStringWithKind(start, token.FString)
	}

	if isIdentifierStart(r) {
		return lexer.lexIdentifier(start), nil
	}

	if r == '"' {
		return lexer.lexStringWithKind(start, token.String)
	}

	return lexer.lexSymbol(start)
}

func (lexer *Lexer) consumeNewline() string {
	start := lexer.offset
	if lexer.peekRune('\r') {
		lexer.advance()
		if lexer.peekRune('\n') {
			lexer.advance()
		} else {
			lexer.line++
			lexer.column = 1
		}
		return lexer.source[start:lexer.offset]
	}
	lexer.advance()
	return lexer.source[start:lexer.offset]
}

func (lexer *Lexer) lexIdentifier(start token.Position) token.Token {
	for {
		r, ok := lexer.peek()
		if !ok || !isIdentifierPart(r) {
			break
		}
		lexer.advance()
	}

	lexeme := lexer.source[start.Offset:lexer.offset]
	return lexer.makeToken(token.LookupIdentifier(lexeme), lexeme, start)
}

func (lexer *Lexer) lexNumber(start token.Position) token.Token {
	for {
		r, ok := lexer.peek()
		if !ok || !unicode.IsDigit(r) {
			break
		}
		lexer.advance()
	}

	kind := token.Integer
	if lexer.peekRune('.') {
		if next, ok := lexer.peekAhead(1); ok && unicode.IsDigit(next) {
			kind = token.Float
			lexer.advance()
			for {
				r, ok := lexer.peek()
				if !ok || !unicode.IsDigit(r) {
					break
				}
				lexer.advance()
			}
		}
	}

	return lexer.makeToken(kind, lexer.source[start.Offset:lexer.offset], start)
}

func (lexer *Lexer) lexStringWithKind(start token.Position, kind token.Kind) (token.Token, *Error) {
	lexer.advance()

	for {
		r, ok := lexer.peek()
		if !ok {
			err := diagnostic.Error(start, "unterminated string literal")
			return token.Token{}, &err
		}
		if r == '\n' || r == '\r' {
			err := diagnostic.Error(start, "unterminated string literal")
			return token.Token{}, &err
		}
		if r == '"' {
			lexer.advance()
			return lexer.makeToken(kind, lexer.source[start.Offset:lexer.offset], start), nil
		}
		if r == '\\' {
			lexer.advance()
			if _, ok := lexer.peek(); !ok {
				err := diagnostic.Error(start, "unterminated string literal")
				return token.Token{}, &err
			}
		}
		lexer.advance()
	}
}

func (lexer *Lexer) lexSymbol(start token.Position) (token.Token, *Error) {
	r, _ := lexer.peek()

	switch r {
	case '+':
		lexer.advance()
		return lexer.makeToken(token.Plus, "+", start), nil
	case '-':
		if lexer.peekAheadRune(1, '>') {
			lexer.advance()
			lexer.advance()
			return lexer.makeToken(token.ReturnArrow, "->", start), nil
		}
		lexer.advance()
		return lexer.makeToken(token.Minus, "-", start), nil
	case '*':
		lexer.advance()
		return lexer.makeToken(token.Star, "*", start), nil
	case '%':
		lexer.advance()
		return lexer.makeToken(token.Percent, "%", start), nil
	case '=':
		if lexer.peekAheadRune(1, '>') {
			lexer.advance()
			lexer.advance()
			return lexer.makeToken(token.FatArrow, "=>", start), nil
		}
		return lexer.lexOneOrTwoChar(start, '=', token.Assign, token.Equal), nil
	case '!':
		return lexer.lexOneOrTwoChar(start, '=', token.Not, token.NotEqual), nil
	case '<':
		return lexer.lexOneOrTwoChar(start, '=', token.Less, token.LessEqual), nil
	case '>':
		return lexer.lexOneOrTwoChar(start, '=', token.Greater, token.GreaterEqual), nil
	case '&':
		if lexer.peekAheadRune(1, '&') {
			lexer.advance()
			lexer.advance()
			return lexer.makeToken(token.And, "&&", start), nil
		}
	case '|':
		if lexer.peekAheadRune(1, '|') {
			lexer.advance()
			lexer.advance()
			return lexer.makeToken(token.Or, "||", start), nil
		}
	case '/':
		return lexer.lexSlash(start)
	case '(':
		lexer.advance()
		return lexer.makeToken(token.LParen, "(", start), nil
	case ')':
		lexer.advance()
		return lexer.makeToken(token.RParen, ")", start), nil
	case '{':
		lexer.advance()
		return lexer.makeToken(token.LBrace, "{", start), nil
	case '}':
		lexer.advance()
		return lexer.makeToken(token.RBrace, "}", start), nil
	case '[':
		lexer.advance()
		return lexer.makeToken(token.LBracket, "[", start), nil
	case ']':
		lexer.advance()
		return lexer.makeToken(token.RBracket, "]", start), nil
	case ':':
		lexer.advance()
		return lexer.makeToken(token.Colon, ":", start), nil
	case ',':
		lexer.advance()
		return lexer.makeToken(token.Comma, ",", start), nil
	case '.':
		if lexer.peekAheadRune(1, '.') {
			lexer.advance()
			lexer.advance()
			return lexer.makeToken(token.Range, "..", start), nil
		}
		lexer.advance()
		return lexer.makeToken(token.Dot, ".", start), nil
	case ';':
		lexer.advance()
		return lexer.makeToken(token.Semicolon, ";", start), nil
	}

	lexer.advance()
	err := diagnostic.Error(start, fmt.Sprintf("unexpected character %q", r))
	return token.Token{}, &err
}

func (lexer *Lexer) lexSlash(start token.Position) (token.Token, *Error) {
	if lexer.peekAheadRune(1, '/') {
		lexer.advance()
		lexer.advance()
		for {
			r, ok := lexer.peek()
			if !ok || r == '\n' || r == '\r' {
				break
			}
			lexer.advance()
		}
		return lexer.NextToken()
	}

	if lexer.peekAheadRune(1, '*') {
		lexer.advance()
		lexer.advance()
		for {
			r, ok := lexer.peek()
			if !ok {
				err := diagnostic.Error(start, "unterminated block comment")
				return token.Token{}, &err
			}
			if r == '*' && lexer.peekAheadRune(1, '/') {
				lexer.advance()
				lexer.advance()
				return lexer.NextToken()
			}
			if r == '\n' || r == '\r' {
				newlineStart := lexer.position()
				lexeme := lexer.consumeNewline()
				lexer.pending = append(lexer.pending, lexer.makeToken(token.Newline, lexeme, newlineStart))
				continue
			}
			lexer.advance()
		}
	}

	lexer.advance()
	return lexer.makeToken(token.Slash, "/", start), nil
}

func (lexer *Lexer) lexOneOrTwoChar(start token.Position, second rune, one token.Kind, two token.Kind) token.Token {
	lexer.advance()
	if lexer.peekRune(second) {
		lexer.advance()
		return lexer.makeToken(two, lexer.source[start.Offset:lexer.offset], start)
	}
	return lexer.makeToken(one, lexer.source[start.Offset:lexer.offset], start)
}

func (lexer *Lexer) skipHorizontalWhitespace() {
	for {
		r, ok := lexer.peek()
		if !ok || (r != ' ' && r != '\t') {
			return
		}
		lexer.advance()
	}
}

func (lexer *Lexer) makeToken(kind token.Kind, lexeme string, start token.Position) token.Token {
	return token.Token{
		Kind:   kind,
		Lexeme: lexeme,
		Start:  start,
		End:    lexer.position(),
	}
}

func (lexer *Lexer) position() token.Position {
	return token.Position{
		Filename: lexer.filename,
		Offset:   lexer.offset,
		Line:     lexer.line,
		Column:   lexer.column,
	}
}

func (lexer *Lexer) peek() (rune, bool) {
	if lexer.offset >= len(lexer.source) {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(lexer.source[lexer.offset:])
	return r, true
}

func (lexer *Lexer) peekAhead(distance int) (rune, bool) {
	offset := lexer.offset
	for i := 0; i < distance; i++ {
		if offset >= len(lexer.source) {
			return 0, false
		}
		_, size := utf8.DecodeRuneInString(lexer.source[offset:])
		offset += size
	}
	if offset >= len(lexer.source) {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(lexer.source[offset:])
	return r, true
}

func (lexer *Lexer) peekRune(expected rune) bool {
	r, ok := lexer.peek()
	return ok && r == expected
}

func (lexer *Lexer) peekAheadRune(distance int, expected rune) bool {
	r, ok := lexer.peekAhead(distance)
	return ok && r == expected
}

func (lexer *Lexer) advance() (rune, bool) {
	if lexer.offset >= len(lexer.source) {
		return 0, false
	}

	r, size := utf8.DecodeRuneInString(lexer.source[lexer.offset:])
	lexer.offset += size
	if r == '\n' {
		lexer.line++
		lexer.column = 1
	} else {
		lexer.column++
	}
	return r, true
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r)
}
