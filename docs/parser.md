# Parser

The Phase 2 parser turns the Phase 1 token stream into an abstract syntax tree.
It does not resolve names, validate types, or check whether functions exist;
those responsibilities belong to semantic analysis.

## Supported Syntax

- Top-level function declarations.
- `public`, `private`, and `turbo` function modifiers.
- Function parameters and return type annotations.
- Blocks delimited by braces.
- `const` and `let` variable declarations.
- Assignment statements.
- Return statements.
- Expression statements, including function calls.
- `if`/`else`.
- `while`.
- Integer, float, string, bool, and null literals.
- Identifier expressions.
- Unary `!` and `-`.
- Binary arithmetic, comparison, equality, and boolean operators.
- Optional semicolon or newline statement boundaries.

## Public API

```go
program, diagnostics := parser.Parse(filename, source)
```

The returned `*ast.Program` contains declarations from the source file. Syntax
errors are returned as parser diagnostics with source positions. Lexer errors
are converted into parser diagnostics so callers can use a single result shape
while the compiler frontend is still small.

For tests and development, `Program.DebugString()` returns a stable tree-shaped
view of the AST. It is intended for debugging compiler work, not as a serialized
language format.

## Current Limits

- Only function declarations are accepted at top level.
- Type names are parsed as identifiers and are not validated yet.
- `else if` is not special-cased; use nested `if` inside an `else` block for now.
- The parser recovers from simple statement and top-level errors, but recovery
  is intentionally conservative until later phases need richer diagnostics.
