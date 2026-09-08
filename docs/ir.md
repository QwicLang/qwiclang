# Intermediate Representation

Phase 4 introduces a deliberately small IR between semantic analysis and code
generation. The IR is not an optimizer framework; its job is to give the
backend a stable representation that does not depend directly on AST nodes.

## Model

An IR module contains functions. Each function contains basic blocks. Blocks
contain instructions.

The initial instruction set supports:

- `constant`
- `variable`
- `load`
- `store`
- `binary`
- `call`
- `return`
- `branch`
- `jump`

`jump` is included as the unconditional companion to conditional branches so
`if` and `while` can be represented as basic blocks without inventing a codegen
contract later.

## Public API

```go
module, diagnostics := ir.Build(semaResult)
module, diagnostics := ir.BuildSource(filename, source)
```

`BuildSource` runs parsing and semantic analysis first. If either frontend
phase reports diagnostics, IR generation does not continue.

`Module.DebugString()` returns a stable text form for tests and development
debugging. It is not a serialized IR format.

## Current Limits

- The IR has no optimization pass.
- Temporary names are local to each function and deterministic.
- Blocks are generated in source order with simple labels such as `if.then.1`
  and `while.cond.1`.
- Expression result types are inferred again from checked source structure where
  needed. Later phases can preserve typed expression data from semantic
  analysis if that becomes useful.
