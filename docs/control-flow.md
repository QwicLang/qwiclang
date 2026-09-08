# Control Flow

Phase 7 supports procedural control flow with braces defining scope.

## If and Else

```qwic
if value > 10 {
    print("large")
} else {
    print("small")
}
```

Conditions must be `bool`.

## While

```qwic
let i: int = 0

while i < 5 {
    print(i)
    i = i + 1
}
```

`while` evaluates its condition before every iteration. Braces define the loop
body. Indentation is optional and does not affect syntax.

## Current Limits

- `break` and `continue` are not implemented yet.
- `else if` is not special syntax yet; use a nested `if` inside `else` when
  needed.
