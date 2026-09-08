# QwicLang Development Phases

This document tracks the planned implementation phases for QwicLang v0.
The goal is a small, readable Go compiler that can compile `.qw` source into
a native executable and run basic programs.

## Phase 0 - Repository Bootstrap

Objective: create the basic Go project and development foundation.

Deliverables:

- Go module.
- `cmd/qwic` command.
- Basic CLI with `--help`.
- Initial source tree.
- README.
- Test infrastructure.
- `.gitignore`.
- Formatting and linting setup.
- Sample `.qw` file.

Completion criteria:

- `go run ./cmd/qwic --help` works.
- `go build ./...` passes.
- `go test ./...` passes.

## Phase 1 - Lexer

Objective: convert Qwic source text into tokens.

Deliverables:

- Token definitions for identifiers, literals, keywords, operators,
  punctuation, braces, parentheses, newlines, semicolons, and EOF.
- Lexer package that handles:
  - identifiers
  - keywords
  - integers
  - floats
  - strings
  - operators
  - punctuation
  - braces
  - parentheses
  - line comments
  - block comments
  - explicit semicolons
  - newlines
  - EOF
- Keywords:
  - `public`
  - `private`
  - `const`
  - `let`
  - `func`
  - `return`
  - `if`
  - `else`
  - `while`
  - `struct`
  - `import`
  - `module`
  - `turbo`
  - `true`
  - `false`
  - `null`
- Tests for:
  - hello world
  - variables
  - function declarations
  - numbers
  - strings
  - operators
  - comments
  - newlines
  - semicolons
  - keywords

Completion criteria:

- The lexer produces the expected token stream for:

```qwic
public func main() {
    const x: int = 10
    print(x)
}
```

## Phase 2 - AST and Parser

Objective: transform valid token streams into a structured AST.

Deliverables:

- AST nodes for programs, declarations, statements, expressions, and types.
- Parser support for:
  - programs
  - function declarations
  - variable declarations
  - assignments
  - literals
  - identifiers
  - expressions
  - function calls
  - return statements
  - blocks
  - if/else
  - while
  - visibility modifiers
  - turbo modifier
- Useful syntax errors with source positions.
- AST debug printing for development.

Completion criteria:

- The parser produces a correct AST for:

```qwic
public func main() {
    const x: int = 10

    if x > 5 {
        print(x)
    }
}
```

## Phase 3 - Semantic Analysis and Type Checking

Objective: validate whether syntactically valid programs are semantically valid.

Deliverables:

- Scope handling.
- Symbol table.
- Variable lookup.
- Function lookup.
- Declaration checking.
- Assignment checking.
- Return type checking.
- Function argument checking.
- Visibility representation.
- Basic type inference.
- Basic type compatibility.
- Source-located diagnostics.

Required rejected programs:

- Assigning a `string` value to an `int` variable.
- Assigning to a `const`.
- Calling an unknown function.

Completion criteria:

- Compiler errors are understandable and source-located.

## Phase 4 - Minimal IR

Objective: introduce a simple intermediate representation between the frontend
and backend.

Deliverables:

- IR support for:
  - function
  - constant
  - variable
  - load
  - store
  - binary operation
  - call
  - return
  - branch

Completion criteria:

- A basic Qwic function can be represented completely in IR.

## Phase 5 - First Code Generator

Objective: compile Qwic programs into executable native programs.

Preferred route:

```text
Qwic IR -> LLVM IR -> Clang/LLVM toolchain -> native executable
```

Allowed bootstrap route:

```text
Qwic IR -> C source -> system compiler -> native executable
```

Deliverables:

- Backend independent from frontend AST details.
- Deterministic generated code.
- External compiler errors captured and reported.
- Build command capable of producing an executable.
- `print()` support sufficient for hello world.

Completion criteria:

- `qwic build examples/hello.qw` produces an executable.
- Running the executable prints `Hello, Qwic`.

## Phase 6 - Expressions and Functions

Objective: make the compiler useful for basic computation.

Deliverables:

- Arithmetic.
- Comparisons.
- Boolean expressions.
- Variables.
- Assignment.
- Function calls.
- Function arguments.
- Return values.

Completion criteria:

- A program containing `add(20, 22)` can compile and print `42`.

## Phase 7 - Control Flow

Objective: support meaningful procedural programs.

Deliverables:

- `if`.
- `else`.
- `while`.
- Branching.
- Loop control.
- Return from nested blocks.

Completion criteria:

- A loop from `0` to `9` can compile and print each value.

## Phase 8 - Nano

Objective: implement the first Qwic-specific numeric feature.

Deliverables:

- `nano` type.
- Nano literal parsing.
- Assignment.
- Arithmetic.
- Comparisons.
- Native code generation.
- Documented internal representation.

Completion criteria:

- A program can assign, multiply, and print a `nano` value.
- No undocumented floating-point conversions exist.

## Phase 9 - Modules and Visibility

Objective: introduce multi-file compilation and API boundary enforcement.

Deliverables:

- `module`.
- `import`.
- `public`.
- `private`.
- Compile-time rejection of external private access.

Completion criteria:

- Public symbols can be imported and used across files.
- Private symbols cannot be accessed from other modules.

## Phase 10 - Runtime Foundation

Objective: introduce a minimal Qwic runtime.

Deliverables:

- Runtime initialization.
- Printing.
- Basic memory helpers.
- Process exit handling.
- Basic platform abstraction.

Completion criteria:

- Runtime remains small and supports v0 generated programs.

## Phase 11 - CLI and Developer Experience

Objective: make the compiler pleasant and predictable to use.

Deliverables:

- `qwic build`.
- `qwic run`.
- `qwic check`.
- `qwic fmt`.
- `qwic clean`.
- Useful error messages.
- Stable exit codes.
- Source file reporting.
- Output paths.
- Build diagnostics.

Completion criteria:

- Common compiler workflows can be run through stable CLI commands.

## Phase 12 - v0 Release

Objective: produce the first usable QwicLang compiler.

v0 support:

- Variables.
- `const`.
- `let`.
- Basic types.
- `nano`.
- Functions.
- Function calls.
- Return.
- `if`/`else`.
- `while`.
- Strings.
- Arithmetic.
- Comparisons.
- Comments.
- Optional semicolons.
- Optional indentation.
- Basic `public`/`private`.
- Basic modules.
- Native executable generation.

Required demonstrations:

- Hello world prints `Hello, Qwic`.
- Function call program prints `42`.
- Precision example prints a `nano` value.
- Control-flow example loops and prints values.

Completion criteria:

- A developer can run `go build ./cmd/qwic`.
- `qwic run examples/hello.qw` prints `Hello, Qwic`.
- A computation program using `add(20, 22)` prints `42`.
