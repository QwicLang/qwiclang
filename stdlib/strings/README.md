# strings

The `strings` package provides basic string operations through explicit imports:

```qwic
import strings

public func main() {
    const name = "  QwicLang  "
    const clean = strings.trim(name)
    const upper = strings.upper(clean)

    print(upper)
}
```

Output:

```text
QWICLANG
```

## Implemented API

- `strings.length(value: string): int`
- `strings.empty(value: string): bool`
- `strings.trim(value: string): string`
- `strings.trimLeft(value: string): string`
- `strings.trimRight(value: string): string`
- `strings.upper(value: string): string`
- `strings.lower(value: string): string`
- `strings.contains(value: string, needle: string): bool`
- `strings.startsWith(value: string, prefix: string): bool`
- `strings.endsWith(value: string, suffix: string): bool`
- `strings.indexOf(value: string, needle: string): int`

## Current Limits

- `upper` and `lower` use C runtime character mapping. ASCII behavior is the
  supported v0 contract; full Unicode case mapping is deferred.
- Functions that return strings allocate through the Qwic runtime. Lifetime
  management is still runtime-owned in v0.
- `substring`, `slice`, `replace`, `split`, `join`, conversions, and bytes APIs
  are planned but not implemented yet.
