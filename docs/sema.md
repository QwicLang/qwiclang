# Semantic Analysis

Phase 3 validates that a parsed Qwic program is meaningful before later phases
lower it into IR.

## Responsibilities

- Collect top-level function declarations.
- Preserve function visibility and `turbo` metadata in semantic symbols.
- Resolve local variables and function calls.
- Track lexical scopes for function bodies, `if`/`else`, and `while` blocks.
- Check `const` assignment rules.
- Check declared variable types against initializer expressions.
- Check function argument counts and argument types.
- Check return statements against the containing function return type.
- Require boolean conditions for `if` and `while`.
- Report source-located diagnostics.

## Builtins

`print(value)` is treated as a builtin function during Phase 3 so examples can
be checked before the runtime and code generator exist. It accepts one argument
of any type and returns `void`.

Imported standard-library package functions are also registered as builtin
semantic symbols. They still require explicit imports, so `strings.trim(value)`
is valid only after `import strings`.

## Type Rules

The initial builtin types are:

- `void`
- `bool`
- `int`
- `float`
- `nano`
- `string`
- `list`
- `set`
- `dictionary`
- `tuple`

Integer literals infer as `int`, decimal literals infer as `float`, string
literals infer as `string`, `true` and `false` infer as `bool`, and `null`
infers as `null`.

Most assignments require exact type compatibility. One temporary v0 exception
allows assigning a decimal literal expression to a declared `nano` variable.
That keeps `nano` present in the type system before the later fixed-point
literal implementation is designed.

`nano` arithmetic also accepts integer scalar operands, so `delta * 2` remains
typed as `nano`.

The collection types are opaque runtime-backed handles in v0. They are created
and used through explicit standard packages such as `lists`, `sets`,
`dictionaries`, and `tuples`.

## Current Limits

- Cross-file module boundary enforcement currently covers function visibility.
- Visibility is represented but not enforced for single-file programs.
- `print` and the implemented standard-library package functions are builtin
  symbols.
- Type inference is local to initializer and expression trees.
- Return path analysis is intentionally simple.
