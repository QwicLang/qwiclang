package format

import (
	"strings"

	"qwiclang/internal/diagnostic"
	"qwiclang/internal/lexer"
	"qwiclang/internal/token"
)

type Diagnostic = diagnostic.Diagnostic

func Source(filename, source string) (string, []Diagnostic) {
	tokens, diagnostics := lexer.Lex(filename, source)
	if len(diagnostics) > 0 {
		return "", diagnostics
	}

	formatter := formatter{}
	return formatter.format(tokens), nil
}

type formatter struct {
	builder     strings.Builder
	indent      int
	atLineStart bool
	lastKind    token.Kind
}

func (formatter *formatter) format(tokens []token.Token) string {
	formatter.atLineStart = true
	for _, tok := range tokens {
		if tok.Kind == token.EOF {
			break
		}
		formatter.writeToken(tok)
	}
	output := strings.TrimRight(formatter.builder.String(), " \t\n")
	if output != "" {
		output += "\n"
	}
	return output
}

func (formatter *formatter) writeToken(tok token.Token) {
	switch tok.Kind {
	case token.Newline, token.Semicolon:
		formatter.newline()
	case token.LBrace:
		formatter.spaceBeforeIfNeeded()
		formatter.builder.WriteString("{")
		formatter.indent++
		formatter.newline()
	case token.RBrace:
		formatter.indent--
		if !formatter.atLineStart {
			formatter.newline()
		}
		formatter.writeIndent()
		formatter.builder.WriteString("}")
		formatter.lastKind = tok.Kind
	case token.LParen:
		formatter.builder.WriteString("(")
		formatter.lastKind = tok.Kind
	case token.RParen:
		formatter.builder.WriteString(")")
		formatter.lastKind = tok.Kind
	case token.Colon:
		formatter.builder.WriteString(": ")
		formatter.lastKind = tok.Kind
	case token.Comma:
		formatter.builder.WriteString(", ")
		formatter.lastKind = tok.Kind
	case token.Assign, token.Equal, token.NotEqual, token.Less, token.LessEqual, token.Greater, token.GreaterEqual,
		token.Plus, token.Minus, token.Star, token.Slash, token.Percent, token.And, token.Or:
		formatter.spaceBeforeIfNeeded()
		formatter.builder.WriteString(tok.Lexeme)
		formatter.builder.WriteString(" ")
		formatter.lastKind = tok.Kind
	default:
		if formatter.atLineStart {
			formatter.writeIndent()
		} else if formatter.needsSpaceBefore(tok.Kind) {
			formatter.builder.WriteString(" ")
		}
		formatter.builder.WriteString(tok.Lexeme)
		formatter.atLineStart = false
		formatter.lastKind = tok.Kind
	}
}

func (formatter *formatter) needsSpaceBefore(kind token.Kind) bool {
	if strings.HasSuffix(formatter.builder.String(), " ") {
		return false
	}
	if formatter.lastKind == 0 || formatter.lastKind == token.LParen || formatter.lastKind == token.Dot {
		return false
	}
	switch kind {
	case token.RParen, token.Comma, token.Colon, token.Dot:
		return false
	default:
		return true
	}
}

func (formatter *formatter) spaceBeforeIfNeeded() {
	if formatter.atLineStart {
		formatter.writeIndent()
		return
	}
	output := formatter.builder.String()
	if output == "" || strings.HasSuffix(output, " ") || strings.HasSuffix(output, "\n") {
		return
	}
	formatter.builder.WriteString(" ")
}

func (formatter *formatter) newline() {
	output := strings.TrimRight(formatter.builder.String(), " \t")
	formatter.builder.Reset()
	formatter.builder.WriteString(output)
	if !strings.HasSuffix(output, "\n") {
		formatter.builder.WriteString("\n")
	}
	formatter.atLineStart = true
	formatter.lastKind = token.Newline
}

func (formatter *formatter) writeIndent() {
	if formatter.indent < 0 {
		formatter.indent = 0
	}
	formatter.builder.WriteString(strings.Repeat("    ", formatter.indent))
	formatter.atLineStart = false
}
