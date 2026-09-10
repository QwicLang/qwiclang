# tuples

`tuples` is a v0 runtime-backed fixed pair package.

## Implemented API

- `tuples.new2(first: string, second: string): tuple`
- `tuples.first(tuple: tuple): string`
- `tuples.second(tuple: tuple): string`
- `tuples.length(tuple: tuple): int`

## Current Limits

- Tuples currently support two string values only.
- Literal tuple syntax is deferred until the parser and type system can model
  tuple arity and element types.
