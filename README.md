![Qwic](https://github.com/QwicLang/qwiclang/blob/main/image/readme-banner.png)
# QwicLang

QwicLang is an experimental systems programming language focused on readable
syntax, fast compilation, native executables, and precise numeric work.

The project is in active v0 development. The compiler is currently written in
Go and already supports the full bootstrap pipeline:

```text
Qwic source -> lexer -> parser -> semantic analysis -> IR -> C -> runtime -> native executable
```

The long-term goal is for the Qwic compiler to become self-hosted: once the
language is stable enough, the compiler will be rewritten in Qwic.

## Why QwicLang

QwicLang is designed around a simple idea: high-performance native software
should not require noisy syntax or heavyweight tooling.

The language aims to feel familiar to developers coming from JavaScript or
TypeScript, while compiling to small native programs suitable for systems,
networking, concurrency, games, simulations, and precise numerical workloads.

## Current Compiler Status

Implemented today:

- `.qw` source files
- lexer
- parser and AST
- semantic analysis and type checking
- minimal IR
- bootstrap C backend
- native executable generation through the system C compiler
- small C runtime
- CLI commands for build, run, check, fmt, and clean
- variables with `const` and `let`
- primitive types: `void`, `bool`, `int`, `float`, `nano`, `string`, `list`,
  `set`, `dictionary`, `tuple`
- functions, parameters, calls, and returns
- Python-style f-string interpolation with `{expression}` placeholders
- arithmetic and comparisons
- boolean expressions
- `if` / `else`
- `while`
- basic `nano` support
- same-directory modules and imports
- `public` / `private` function visibility checks
- standard `strings` package
- bootstrap data-structure packages: `lists`, `sets`, `dictionaries`, `tuples`

Still intentionally deferred:

- final LLVM backend
- most standard-library packages beyond the initial string and data-structure
  packages
- package registry
- advanced module paths
- generics
- interfaces or traits
- async runtime
- networking standard library
- garbage collection or ownership model
- self-hosted compiler

See [PHASE.md](PHASE.md) for the full v0 implementation plan.

## Install From Source

Prerequisites:

- Go
- a system C compiler available as `cc`

Build the compiler:

```bash
go build -o qwic ./cmd/qwic
```

Install it somewhere on your `PATH`, for example:

```bash
mkdir -p ~/go/bin
mv qwic ~/go/bin/qwic
```

Then verify it:

```bash
qwic --help
```

The `qwic` binary includes the small bootstrap runtime sources it needs during
compilation, so it does not need to be run from the repository root. You still
need a system C compiler available as `cc`, because v0 currently lowers Qwic IR
to C before producing a native executable.

## Quick Start

Create `hello.qw`:

```qwic
public func main() {
    print("Hello, Qwic")
}
```

Run it:

```bash
qwic run hello.qw
```

Expected output:

```text
Hello, Qwic
```

Build a native executable:

```bash
qwic build hello.qw -o hello
./hello
```

## Examples

Functions:

```qwic
func add(a: int, b: int): int {
    return a + b
}

public func main() {
    print(add(20, 22))
}
```

Output:

```text
42
```

Control flow:

```qwic
public func main() {
    let i: int = 0

    while i < 5 {
        print(i)
        i = i + 1
    }
}
```

Basic `nano`:

```qwic
public func main() {
    const delta: nano = 0.016666
    const value: nano = delta * 2

    print(value)
}
```

Modules:

```qwic
// users.qw
module users

private func validateUser() {
}

public func createUser() {
    print("created")
}
```

```qwic
// main.qw
import users

public func main() {
    users.createUser()
}
```

## CLI

```bash
qwic build <source.qw> [-o output]
qwic run <source.qw>
qwic check <source.qw>
qwic fmt <source.qw>
qwic clean [source.qw]
qwic --help
```

Command behavior:

- `build` compiles a source file and same-directory imports into a native
  executable.
- `run` builds a temporary executable, runs it, and returns the program exit
  status.
- `check` runs the frontend and IR generation without producing an executable.
- `fmt` rewrites one `.qw` file with conservative formatting.
- `clean` removes v0 build artifacts.

## Architecture

The v0 compiler is deliberately small:

```text
cmd/qwic
  -> lexer
  -> parser / ast
  -> sema / types
  -> ir
  -> codegen
  -> runtime
```

This keeps the frontend independent from the bootstrap C backend. LLVM is the
intended future backend, but the current priority is a correct, readable,
tested compiler that can run real programs.

## Development

Run all tests:

```bash
go test ./...
```

Build all packages:

```bash
go build ./...
```

Run sample programs:

```bash
go run ./cmd/qwic run examples/hello.qw
go run ./cmd/qwic run examples/functions.qw
go run ./cmd/qwic run examples/control_flow.qw
go run ./cmd/qwic run examples/nano.qw
go run ./cmd/qwic run examples/main.qw
```

## Project Principles

- Keep the compiler understandable.
- Prefer small, complete features over broad stubs.
- Keep compiler phases separate.
- Test every language feature.
- Do not claim planned features are implemented.
- Optimize only after correctness and clarity.

## License

QwicLang is licensed under the [MIT License](LICENCE).
