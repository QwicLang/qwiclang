# Code Generation

Phase 5 uses a temporary C backend to reach the first native executable
milestone without blocking on LLVM integration.

The pipeline is:

```text
Qwic source -> lexer -> parser -> semantic analysis -> IR -> C -> cc -> executable
```

The frontend remains independent from this backend. Later LLVM work should
replace the backend without requiring lexer, parser, semantic, or IR rewrites.

## Supported Output

The C backend currently supports the Phase 4 IR instruction set:

- constants
- variables
- loads
- stores
- binary operations
- calls
- returns
- conditional branches
- jumps
- formatted strings

`print(value)` is emitted as a call into the Qwic runtime. It supports `int`,
`float`, `nano`, `string`, and `bool` values.

F-strings are lowered to a simple IR formatting instruction. The C backend emits
a deterministic `snprintf` size pass, allocates the result with `qwic_alloc`,
then writes the final string with a second `snprintf` call.

## CLI

```bash
go run ./cmd/qwic --help
go run ./cmd/qwic build examples/hello.qw
./hello
go run ./cmd/qwic run examples/hello.qw
```

Generated C is written to a temporary directory and removed after a successful
or failed build unless the codegen API is called with `KeepC`.

## Current Limits

- This is a bootstrap backend, not the final LLVM backend.
- `nano` is emitted as `double` until the fixed-point representation is
  implemented in a later phase.
- String literals preserve lexer spelling and are emitted directly as C string
  literals.
- F-string formatting supports `int`, `float`, `nano`, `string`, and `bool`
  values. Width, precision, and conversion specifiers are not implemented yet.
