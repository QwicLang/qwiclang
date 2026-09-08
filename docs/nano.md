# Nano

Phase 8 gives `nano` a complete basic path through lexing, parsing, semantic
analysis, IR, code generation, and native execution.

## Syntax

```qwic
const delta: nano = 0.016666
const value: nano = delta * 2

print(value)
```

## v0 Representation

`nano` is represented as IEEE-754 `double` in the temporary C backend. Decimal
source literals are parsed as float literals, and semantic analysis permits a
float literal initializer for a declared `nano` value.

This is intentionally documented as a v0 bootstrap representation. It is not
the final fixed-point design. The fixed-point storage model will replace this
only after the compiler pipeline is stable enough to make that change cleanly.

## Current Operations

- Declaration with `nano` type annotations.
- Assignment from decimal literals.
- Arithmetic with other `nano` values.
- Multiplication and division by integer scalar values.
- Numeric comparisons.
- Printing with six fractional digits.
