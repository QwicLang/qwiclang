# QwicLang

QwicLang is an experimental statically typed language implemented in Go.

The current compiler is a v0 bootstrap compiler. It reads `.qw` source, lexes,
parses, performs semantic checks, lowers to a small IR, emits C, links a tiny
runtime, and produces a native executable.

## Try It

```bash
go run ./cmd/qwic --help
go run ./cmd/qwic run examples/hello.qw
go run ./cmd/qwic run examples/functions.qw
go run ./cmd/qwic run examples/control_flow.qw
go run ./cmd/qwic run examples/nano.qw
go run ./cmd/qwic run examples/main.qw
```

## Commands

```bash
qwic build <source.qw> [-o output]
qwic run <source.qw>
qwic check <source.qw>
qwic fmt <source.qw>
qwic clean [source.qw]
```

## Current Status

Implemented:

- lexer
- parser and AST
- semantic analysis and type checking
- minimal IR
- bootstrap C backend
- small runtime
- basic CLI
- variables
- functions
- calls
- returns
- arithmetic
- comparisons
- boolean expressions
- `if`/`else`
- `while`
- basic `nano`
- same-directory modules/imports
- `public`/`private` function visibility checks

Not implemented yet:

- final LLVM backend
- package registry
- advanced runtime
- concurrency
- generics
- self-hosting

See [PHASE.md](PHASE.md) for the full implementation plan.
