# Lexer

The Phase 1 lexer converts Qwic source text into a flat token stream. It does
not parse expressions, validate declarations, resolve names, or perform type
checking.

## Responsibilities

- Preserve source locations for diagnostics.
- Emit identifiers, literals, keywords, operators, punctuation, newlines,
  explicit semicolons, and EOF.
- Skip line comments and block comments while preserving statement-separating
  newlines.
- Report malformed lexical input with diagnostics instead of silently ignoring
  it.

## Required Token Coverage

The Phase 1 tests cover:

- hello world source
- `const` and `let` variable declarations
- function declarations with visibility and `turbo`
- integer and floating-point literals
- string literals and common escapes
- arithmetic, comparison, assignment, and boolean operators
- line and block comments
- newlines and semicolons
- all Phase 1 keywords listed in `PHASE.md`

## Expected Public API

Until the implementation lands, the tests assume this lightweight API:

```go
l := lexer.New(filename, source)
tokens, diagnostics := l.Lex()
```

Expected token shape:

```go
type Token struct {
    Kind   token.Kind
    Lexeme string
    Start  token.Position
    End    token.Position
}
```

Expected diagnostic shape:

```go
type Error struct {
    Message  string
    Position token.Position
}
```

`token.Kind` should format to stable names such as `Identifier`, `Integer`,
`String`, `Public`, `Func`, `LParen`, `Newline`, and `EOF`. The parser can use
the typed constants directly, while tests can compare the formatted names.

String token lexemes are expected to preserve the source spelling, including
the surrounding quotes and escape sequences. Later compiler stages can decide
where to unescape the value.

## Notes

Primitive type names such as `int`, `float`, `nano`, `string`, `bool`, and
`void` are currently expected to lex as identifiers. The parser and later type
checker are responsible for interpreting them as type names.
