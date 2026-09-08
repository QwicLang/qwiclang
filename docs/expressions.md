# Expressions and Functions

Phase 6 makes Qwic useful for basic computation.

## Supported Features

- Integer, float, string, bool, and null literals.
- Variable declarations with `const` and `let`.
- Assignment to `let` variables.
- Arithmetic operators: `+`, `-`, `*`, `/`, `%`.
- Comparison operators: `<`, `<=`, `>`, `>=`.
- Equality operators: `==`, `!=`.
- Boolean operators: `&&`, `||`, `!`.
- Function declarations with typed parameters.
- Function calls.
- Return values.

Example:

```qwic
func add(a: int, b: int): int {
    return a + b
}

public func main() {
    print(add(20, 22))
}
```

Expected output:

```text
42
```

## Current Limits

- Mixed numeric arithmetic is intentionally conservative.
- `%` is intended for integers in v0.
- Strings can be printed, but string operations are otherwise deferred.
