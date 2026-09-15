# QwicLang v0.1.0-alpha.1

QwicLang `v0.1.0-alpha.1` is the first public developer preview of the Qwic
compiler, runtime, package workflow, and editor tooling.

This release delivers an end-to-end compiler that accepts `.qw` source,
validates it through the frontend and semantic pipeline, lowers it to Qwic IR,
emits portable C, and invokes the platform C toolchain to produce a native
executable.

> [!IMPORTANT]
> This is an alpha release intended for experimentation and contribution.
> Language syntax, runtime behavior, package APIs, and generated-code contracts
> may change before the first stable release.

## Highlights

### Native compilation pipeline

- Lexer with source positions, comments, optional semicolons, and newline-aware
  statement handling.
- Parser and AST for declarations, expressions, control flow, modules, types,
  methods, lambdas, collections, and recoverable errors.
- Semantic analysis with scopes, symbol resolution, type checking, visibility,
  assignment validation, function argument checking, and return validation.
- A backend-independent Qwic IR followed by a readable bootstrap C backend.
- Native executable generation using GCC, Clang, Apple Clang, or MinGW-w64.
- Source-located diagnostics across the compiler pipeline.

### Language features

- Immutable `const` and mutable `let` declarations with type inference.
- Primitive `bool`, `int`, `float`, `nano`, `string`, `void`, and dynamic `any`
  values.
- Typed functions, arguments, return values, calls, and recursion.
- Arithmetic, comparisons, equality, and boolean expressions.
- `if` / `else`, `while`, and `for ... in` control flow.
- Python-style f-string interpolation with Qwic expressions.
- Record types, checked field literals, static type functions, and explicit
  named method receivers.
- Typed lambdas with by-value captures.
- String-based `throw` and structured `try` / `catch` handling.
- The `turbo` function marker, preserved through semantic analysis and IR and
  emitted as native compiler optimization attributes by the C backend.

### Collections and dynamic values

- List literals, indexing, slicing, concatenation, and iteration.
- Dictionary literals with string-key lookup and assignment.
- Set and tuple runtime packages.
- Heterogeneous collection values through the tagged `any` representation.
- Direct printing for lists, dictionaries, and tuples.

### Modules and packages

- Multi-file compilation with `module` and `import`.
- Compile-time `public` and `private` visibility enforcement.
- Shared package installation with `qwic install <package>`.
- Packages are installed once under the platform-specific user directory:
  `~/qwic/packages` on Linux and macOS, or `%USERPROFILE%\qwic\packages` on
  Windows.
- Project-local packages remain available for application-specific code.
- Package lookup supports compiler standard packages, shared packages, project
  packages, and legacy same-directory modules.
- The official package catalog is hosted at
  <https://github.com/QwicLang/packages>.

### Bootstrap standard library

This release includes compiler-known, runtime-backed APIs for:

- `strings`
- `math`
- `time`
- `fs`
- `sync`
- `crypto`
- `net`
- `json`
- `http`
- `lists`
- `sets`
- `dictionaries`
- `tuples`
- `values`

These packages establish the initial API surface. They are deliberately small
and should be treated as experimental during the alpha cycle.

### Signal package compatibility

The compiler and runtime support the language features required by the Qwic
Signal HTTP package, including explicit receivers, qualified package types,
dynamic request values, lambda handlers, routing, JSON values, and a
cross-platform HTTP/1.1 listener.

The integration suite builds a Signal application, starts its native server,
and verifies a routed HTTP response end to end.

### CLI and editor tooling

The `qwic` command includes:

```text
qwic build <source.qw> [-o output]
qwic run <source.qw>
qwic check <source.qw>
qwic fmt <source.qw>
qwic clean [source.qw]
qwic install <package>
qwic lsp
qwic --version
```

A zero-dependency Language Server Protocol implementation provides diagnostics,
document symbols, hover information, definitions, references, completion, and
formatting support for editor integrations. The repository also contains the
initial VS Code/VSCodium language extension.

## Supported Platforms

Release archives are built and tested through GitHub Actions:

| Platform | Architecture | Release archive |
| --- | --- | --- |
| Linux | AMD64 | `qwic-linux-amd64-v0.1.0-alpha.1.zip` |
| Windows | AMD64 | `qwic-windows-amd64-v0.1.0-alpha.1.zip` |
| macOS | Apple Silicon | `qwic-macos-arm64-v0.1.0-alpha.1.zip` |

## Installation

Download the archive for your platform from the GitHub release assets, extract
it, and place `qwic` or `qwic.exe` on your `PATH`.

Linux and macOS users may need to make the binary executable:

```bash
chmod +x qwic
```

Verify the installation:

```bash
qwic --version
```

Qwic currently uses a bootstrap C backend. A C compiler must therefore be
available when building or running Qwic programs:

- Linux: GCC or Clang
- macOS: Apple Clang from the Xcode Command Line Tools
- Windows: MinGW-w64 GCC

Building the compiler from source requires Go 1.24 or newer:

```bash
git clone https://github.com/QwicLang/qwiclang.git
cd qwiclang
go build -o qwic ./cmd/qwic
```

## Quick Start

Create `hello.qw`:

```qwic
public func main() {
    const project: string = "QwicLang"
    print(f"Hello from {project}")
}
```

Run it directly:

```bash
qwic run hello.qw
```

Or build a native executable:

```bash
qwic build hello.qw -o hello
./hello
```

## Known Limitations

- `nano` currently uses an IEEE-754 `double` in the bootstrap backend. It is
  not yet the planned fixed-point representation.
- The C backend is temporary; direct LLVM code generation is not implemented.
- Runtime-managed records, collection values, and closure environments do not
  yet have automatic lifetime reclamation.
- Lambda captures are by-value snapshots, and method references are not
  first-class values.
- Errors are strings. Typed errors, `finally`, stack traces, and cross-thread
  propagation are not available.
- The HTTP listener is synchronous HTTP/1.1 without TLS, streaming, keep-alive,
  concurrent request handling, middleware, or graceful shutdown.
- JSON diagnostics and Unicode escape handling are intentionally limited.
- Package installation uses the catalog's current default-branch snapshot.
  Versions, lockfiles, checksums, updates, and removal are not implemented.
- The formatter is conservative and does not preserve comments.
- F-strings do not yet support Python's formatting mini-language, conversion
  flags, or nested f-strings.
- `break`, `continue`, module aliases, nested module paths, `spawn`, and `await`
  are not implemented.

## Validation

The release pipeline runs the complete Go unit and integration suite before
packaging any binary. CI validates the compiler on Linux AMD64, Windows AMD64,
and macOS ARM64. Integration coverage includes native compilation and execution,
functions, control flow, `nano`, collections, modules, standard packages,
methods, lambdas, exceptions, JSON, and Signal HTTP routing.

## Roadmap

The next alpha releases will focus on:

- Versioned package metadata, dependency resolution, lockfiles, and integrity
  verification.
- A defined ownership and memory-lifetime model.
- A real fixed-point representation for `nano`.
- Expanded standard-library contracts and error handling.
- Concurrency primitives and a thread-safe exception/runtime model.
- Direct LLVM code generation.
- Stronger formatter and language-server behavior.

Once the language and compiler architecture are stable, the long-term plan is
to rewrite the Qwic compiler in Qwic itself. The Go implementation remains the
bootstrap compiler until Qwic is mature enough to support that transition
without compromising correctness.

## Contributing

QwicLang is at the stage where design feedback, platform testing, diagnostics,
compiler tests, runtime work, documentation, and package contributions can have
a direct impact on the language.

- Repository: <https://github.com/QwicLang/qwiclang>
- Package catalog: <https://github.com/QwicLang/packages>
- Issues: <https://github.com/QwicLang/qwiclang/issues>
- Pull requests: <https://github.com/QwicLang/qwiclang/pulls>

Thank you to everyone testing QwicLang and helping build its foundation.
