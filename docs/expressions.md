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
- Python-style f-string interpolation: `f"Hello, {name}"`.
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

F-strings can interpolate ordinary Qwic expressions:

```qwic
public func main() {
    const name: string = "Qwic"
    const count: int = 2

    print(f"Hello, {name}: {count + 1}")
}
```

Expected output:

```text
Hello, Qwic: 3
```

## Current Limits

- Mixed numeric arithmetic is intentionally conservative.
- `%` is intended for integers in v0.
- Strings can be printed and formatted with f-strings, but general string
  operations are otherwise deferred.
- F-strings do not yet implement Python's formatting mini-language, conversion
  flags, or nested f-strings inside interpolation expressions.
