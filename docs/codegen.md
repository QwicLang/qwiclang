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
- record allocation and field access
- direct static and instance method calls
- portable closures for lambdas and named function references
- tagged `any` values and dynamic field/index operations
- struct conversion at typed/dynamic boundaries
- try/catch frames and throws

`print(value)` is emitted as a call into the Qwic runtime. It supports `int`,
`float`, `nano`, `string`, and `bool` values.

F-strings are lowered to a simple IR formatting instruction. The C backend emits
a deterministic `snprintf` size pass, allocates the result with `qwic_alloc`,
then writes the final string with a second `snprintf` call.

Standard-library data-structure calls are lowered to runtime C functions. The
opaque collection types are emitted as `void *` handles in the bootstrap C
backend.

Record declarations become forward-declared C structs and pointer-backed
values. Instance methods receive their explicitly named record receiver as the
first native argument. Lambdas use one callback ABI on every target and carry a
runtime context containing captured values.

Dynamic values use a tagged runtime representation. Calls box typed arguments
when entering `any` code and unbox them when returning to typed code. User
records convert to dictionary-backed values at this boundary, which lets HTTP
callbacks enter typed package APIs without weakening those APIs internally.

Try/catch uses runtime-managed frames around standard C `setjmp`/`longjmp`.
Returns from protected blocks explicitly unwind active frames before leaving the
function.

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
- F-string formatting supports `int`, `float`, `nano`, `string`, `bool`, and `any`
  values. Width, precision, and conversion specifiers are not implemented yet.
- Collection storage is heterogeneous through tagged values and uses opaque
  runtime pointers in generated C.
- Record values are not automatically reclaimed yet.
- Lambda captures are snapshots and are not automatically reclaimed yet.
- Thrown values and catch bindings are strings; typed errors are deferred.
