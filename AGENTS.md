# CLAUDE.md

## QwicLang — AI Development Instructions

This file defines the engineering rules, development phases, architecture, coding standards, and delivery requirements for building **QwicLang**.

Claude must treat this document as the primary engineering instruction set for this repository.

---

# 1. Project Overview

**Project name:** QwicLang
**Language command:** `qwic`
**Source extension:** `.qw`
**Compiler implementation language:** Go
**Current objective:** Build a functional **QwicLang v0** compiler capable of reading Qwic source code, compiling it into a native executable, and running it.

QwicLang is intended to become a lightweight, readable, statically typed, natively compiled language designed for:

* high performance
* fast compilation
* small binaries
* safe defaults
* scalable systems
* networking
* concurrency
* game development
* precise numerical workloads

The language syntax is intentionally clean and broadly familiar to developers coming from JavaScript/TypeScript, while the implementation and runtime are designed around native execution.

---

# 2. Current Priority

## The immediate objective is NOT to build the complete QwicLang ecosystem.

The immediate objective is:

> **Build a small, clean, understandable compiler that can successfully compile and execute basic QwicLang programs.**

The first successful milestone is:

```text
hello.qw
    ↓
qwic compiler
    ↓
native executable
    ↓
./hello
    ↓
Hello, Qwic
```

Everything in the initial implementation must support reaching this milestone quickly without creating unnecessary architectural debt.

---

# 3. Core Engineering Principles

Claude MUST follow these principles throughout development.

## 3.1 Simplicity First

Do not implement complexity before it is required.

Prefer:

```text
simple implementation
```

over:

```text
generic framework
```

unless the abstraction is clearly justified.

Do not introduce:

* unnecessary dependency layers
* premature optimization
* complex plugin architectures
* speculative abstractions
* unnecessary interfaces
* complicated compiler passes
* features that are not required by the current phase

---

## 3.2 Clean Architecture

Compiler components must have clear responsibilities.

Avoid putting lexer logic into the parser.

Avoid putting semantic validation into the lexer.

Avoid code-generation logic inside AST nodes.

Avoid runtime logic inside the parser.

Each compiler phase should have a clear purpose.

Preferred architecture:

```text
Source
  ↓
Lexer
  ↓
Parser
  ↓
AST
  ↓
Semantic Analysis
  ↓
Type Checking
  ↓
IR
  ↓
Code Generation
  ↓
Native Executable
```

---

## 3.3 Readable Code

The compiler is itself a learning project.

Code must be readable by another developer who has never seen the repository.

Prefer:

```go
func parseFunction()
```

over:

```go
func pf()
```

Prefer:

```go
token.Kind == TokenIdentifier
```

over:

```go
t.K == IDENT
```

unless the shorter version is genuinely clearer and consistently used.

Names must communicate intent.

---

## 3.4 Comments Should Explain Why

Comments should explain decisions, assumptions, or non-obvious compiler behavior.

Good:

```go
// Newlines are preserved as tokens because the language allows
// semicolons to be omitted. The parser uses them as statement
// boundaries when an expression has been completed.
```

Bad:

```go
// Increment i
i++
```

Do not comment obvious code.

---

## 3.5 Small Functions

Functions should generally have one responsibility.

Avoid huge functions such as:

```go
compileProgram()
```

containing:

* parsing
* type checking
* optimization
* code generation
* file writing

Break responsibilities into separate functions and packages.

---

## 3.6 No Dead Code

Do not leave unused prototypes, abandoned implementations, commented-out code, or speculative files in the repository.

Delete obsolete implementations rather than preserving them “just in case.”

Git already provides history.

---

## 3.7 Tests Are Required

Every compiler feature must have tests.

At minimum:

```text
lexer tests
parser tests
semantic tests
code generation tests
integration tests
```

A compiler feature is incomplete until its behavior is tested.

---

# 4. Repository Structure

The initial repository should follow approximately this structure:

```text
qwiclang/
│
├── cmd/
│   └── qwic/
│       └── main.go
│
├── internal/
│   ├── lexer/
│   ├── parser/
│   ├── ast/
│   ├── token/
│   ├── types/
│   ├── sema/
│   ├── ir/
│   ├── codegen/
│   └── diagnostic/
│
├── runtime/
│
├── examples/
│
├── tests/
│
├── docs/
│
├── go.mod
├── go.sum
├── README.md
└── CLAUDE.md
```

The exact structure may evolve, but responsibilities must remain separated.

---

# 5. Compiler Package Responsibilities

## `token`

Defines lexical token kinds.

Examples:

```text
Identifier
Integer
Float
String
Keyword
Operator
LBrace
RBrace
LParen
RParen
Colon
Comma
Semicolon
Newline
EOF
```

---

## `lexer`

Responsible only for converting source text into tokens.

Input:

```text
public func main() {
    print("Hello")
}
```

Output:

```text
PUBLIC
FUNC
IDENTIFIER(main)
LPAREN
RPAREN
LBRACE
IDENTIFIER(print)
LPAREN
STRING
RPAREN
RBRACE
EOF
```

The lexer must not perform type checking or code generation.

---

## `ast`

Contains the Abstract Syntax Tree definitions.

AST nodes should represent the structure of Qwic programs.

Examples:

```text
Program
FunctionDeclaration
VariableDeclaration
StructDeclaration
BlockStatement
IfStatement
ForStatement
ReturnStatement
CallExpression
BinaryExpression
LiteralExpression
IdentifierExpression
```

---

## `parser`

Transforms tokens into the AST.

The parser must:

* produce useful syntax errors
* preserve source positions
* reject invalid syntax cleanly
* avoid semantic validation where possible

---

## `types`

Defines the internal type representation.

Initial types:

```text
void
bool
int
float
nano
string
bytes
```

The implementation must allow the type system to expand later without rewriting the entire compiler.

---

## `sema`

Semantic analysis and validation.

Responsibilities include:

* symbol resolution
* variable existence
* function existence
* scope handling
* visibility checking
* assignment validation
* return validation
* basic type compatibility

Example error:

```text
error: cannot assign value of type 'string' to variable of type 'int'
```

---

## `ir`

Defines an internal representation between semantic analysis and code generation.

The initial IR should remain deliberately simple.

Do not design an extremely sophisticated optimizer IR in v0.

The purpose of the IR is to prevent the rest of the compiler from depending directly on AST implementation details.

---

## `codegen`

Transforms IR into executable code.

For v0, the backend may use the simplest practical native compilation route that preserves the intended architecture.

Preferred direction:

```text
Qwic → IR → LLVM IR → native executable
```

If LLVM integration significantly blocks the first functioning compiler, a temporary C backend may be used strictly as a bootstrap backend:

```text
Qwic → IR → C → system compiler → executable
```

The C backend must be designed so it can later be replaced with LLVM without rewriting the frontend.

---

## `diagnostic`

All compiler errors should use a shared diagnostic system.

Diagnostics should contain:

* severity
* error code if applicable
* message
* source file
* line
* column
* source snippet where practical

Example:

```text
error[QW1003]: unknown variable 'user'

 --> examples/main.qw:12:10
  |
12 | print(user.name)
  |       ^^^^
```

---

# 6. Language v0 Scope

QwicLang v0 is intentionally small.

Implement only the following initial language features.

## Required

### Program structure

```qwic
public func main() {
    print("Hello, Qwic")
}
```

### Variables

```qwic
const name: string = "Qwic"
let count: int = 0
```

### Assignment

```qwic
let count: int = 0
count = count + 1
```

### Primitive types

```text
bool
int
float
nano
string
void
```

`bytes` may be introduced after the basic compiler is stable.

### Functions

```qwic
public func add(a: int, b: int): int {
    return a + b
}
```

### Function calls

```qwic
const result = add(10, 20)
```

### Return

```qwic
return value
```

### Conditions

```qwic
if value > 10 {
    print("large")
} else {
    print("small")
}
```

### Loops

Initially implement the simplest useful loop form.

Preferred first form:

```qwic
while condition {
}
```

`for ... in` may follow immediately after the basic compiler is functional.

### Comments

```qwic
// comment
```

and:

```qwic
/*
    comment
*/
```

### Optional semicolons

Both must be accepted:

```qwic
const x: int = 10
const y: int = 20
```

and:

```qwic
const x: int = 10;
const y: int = 20;
```

### Optional indentation

Indentation must not define syntax.

Braces define scope.

---

# 7. Features Deliberately Deferred From v0

Do NOT implement these before the basic compiler works:

```text
generics
interfaces
traits
advanced ownership
garbage collector
arenas
object pools
channels
async runtime
networking
HTTP
FFI
reflection
macros
pattern matching
package registry
self-hosting
LSP
debugger integration
GPU programming
advanced SIMD
profile-guided optimization
```

These belong to later phases.

---

# 8. `nano` in v0

`nano` must exist in the type system in v0, but implementation may initially be simple.

Example:

```qwic
const frameTime: nano = 0.016666
```

The long-term intent is:

> `nano` is a precise fixed-point numeric type for small fractional values.

Typical use cases:

```text
game timing
animation
simulation
clock values
interpolation
rate calculations
```

For v0, prioritize:

1. lexical support
2. parsing
3. type checking
4. literal representation
5. arithmetic
6. code generation

Do not prematurely build advanced decimal arithmetic infrastructure.

The exact precision and storage model must be documented clearly in code.

---

# 9. Visibility in v0

Implement:

```text
public
private
```

Visibility rules:

* symbols are private by default
* `public` explicitly exposes a symbol
* `private` prevents external module access
* visibility errors happen at compile time

For single-file v0 programs, visibility still needs to be represented in the AST and semantic model even if module boundaries are not fully implemented yet.

---

# 10. `turbo` in v0

`turbo` is initially a **compiler optimization marker**.

Example:

```qwic
turbo func calculate(value: int): int {
    return value * 2
}
```

The v0 compiler does not need advanced turbo optimization.

It must, however:

1. lex `turbo`
2. parse it
3. store it in the function declaration
4. preserve it through semantic analysis
5. allow code generation to recognize it

The compiler may initially treat:

```qwic
turbo func ...
```

as equivalent to:

```qwic
func ...
```

while establishing the architecture needed for future aggressive optimization.

Do not pretend turbo optimization exists before it is implemented.

---

# 11. `spawn` and `await`

Do not implement the runtime concurrency model in the first compiler milestone.

Reserved keywords may exist, but until the runtime is implemented they should either:

* remain unsupported with a clear diagnostic, or
* remain unreserved until their implementation begins.

Do not create fake concurrency behavior.

---

# 12. Build Strategy

The first compiler should compile real Qwic source into a runnable native program.

The initial end-to-end pipeline must look like:

```text
.qw source
   ↓
lexer
   ↓
parser
   ↓
AST
   ↓
semantic analysis
   ↓
IR
   ↓
code generation
   ↓
native executable
```

A simple command such as:

```bash
qwic run examples/hello.qw
```

should eventually:

1. read the source
2. compile it
3. produce a temporary executable
4. execute it
5. return the program's exit status

Likewise:

```bash
qwic build examples/hello.qw
```

should produce an executable.

---

# 13. Development Phases

Development MUST follow these phases in order unless there is a strong technical reason to change the sequence.

---

# PHASE 0 — Repository Bootstrap

## Objective

Create a clean Go project and development foundation.

## Deliverables

* Go module
* `cmd/qwic`
* basic CLI
* source tree
* README
* test infrastructure
* `.gitignore`
* formatting/linting setup
* sample `.qw` file

Expected:

```bash
go run ./cmd/qwic --help
```

works.

## Completion criteria

The repository builds cleanly:

```bash
go build ./...
```

and tests run:

```bash
go test ./...
```

No compiler functionality is required yet.

---

# PHASE 1 — Lexer

## Objective

Turn Qwic source text into tokens.

## Deliverables

Implement:

* identifiers
* keywords
* integers
* floats
* strings
* operators
* punctuation
* braces
* parentheses
* comments
* optional semicolon handling
* newline handling
* EOF

Keywords should include at minimum:

```text
public
private
const
let
func
return
if
else
while
struct
import
module
turbo
true
false
null
```

## Required tests

Test:

```text
hello world
variables
function declarations
numbers
strings
operators
comments
newlines
semicolons
keywords
```

## Completion criteria

Given:

```qwic
public func main() {
    const x: int = 10
    print(x)
}
```

the lexer produces the expected token stream.

---

# PHASE 2 — AST and Parser

## Objective

Transform valid token streams into a structured AST.

## Deliverables

Implement parsing for:

* programs
* function declarations
* variable declarations
* assignments
* literals
* identifiers
* expressions
* function calls
* return statements
* blocks
* if/else
* while
* visibility modifiers
* turbo modifier

## Completion criteria

This must parse:

```qwic
public func main() {
    const x: int = 10

    if x > 5 {
        print(x)
    }
}
```

and produce a correct AST.

Add AST debug printing for development.

---

# PHASE 3 — Semantic Analysis and Type Checking

## Objective

Make the compiler understand whether a valid syntax tree is a valid program.

## Deliverables

Implement:

* scopes
* symbol table
* variable lookup
* function lookup
* declaration checking
* assignment checking
* return type checking
* function argument checking
* visibility representation
* basic type inference
* basic type compatibility

## Required errors

The compiler must reject:

```qwic
let x: int = "hello"
```

and:

```qwic
const x: int = 10
x = 20
```

and:

```qwic
unknownFunction()
```

## Completion criteria

Compiler errors must be understandable and source-located.

---

# PHASE 4 — Minimal IR

## Objective

Introduce a simple intermediate representation.

The IR should represent enough information to generate executable code without depending directly on the AST.

## Deliverables

Implement at least:

```text
function
constant
variable
load
store
binary operation
call
return
branch
```

Do not build a large optimization framework yet.

## Completion criteria

A basic Qwic function can be represented completely in IR.

---

# PHASE 5 — First Code Generator

## Objective

Compile Qwic programs into executable machine code.

## Preferred implementation

Primary target:

```text
Qwic IR
   ↓
LLVM IR
   ↓
Clang/LLVM toolchain
   ↓
native executable
```

If LLVM integration blocks progress, use a temporary C bootstrap backend:

```text
Qwic IR
   ↓
C source
   ↓
system compiler
   ↓
native executable
```

The frontend must remain independent from the backend choice.

## Required program

```qwic
public func main() {
    print("Hello, Qwic")
}
```

Must compile and execute successfully.

## Completion criteria

This works:

```bash
qwic build examples/hello.qw
```

and:

```bash
./hello
```

prints:

```text
Hello, Qwic
```

This is the first major project milestone.

---

# PHASE 6 — Expressions and Functions

## Objective

Make the compiler useful for actual programming.

## Deliverables

Support:

* arithmetic
* comparisons
* boolean expressions
* variables
* assignment
* function calls
* function arguments
* return values

Example:

```qwic
public func add(a: int, b: int): int {
    return a + b
}

public func main() {
    const result = add(10, 20)
    print(result)
}
```

Expected output:

```text
30
```

---

# PHASE 7 — Control Flow

## Objective

Support meaningful procedural programs.

## Deliverables

* `if`
* `else`
* `while`
* branching
* loop control
* return from nested blocks

Test:

```qwic
public func main() {
    let x: int = 0

    while x < 10 {
        print(x)
        x = x + 1
    }
}
```

---

# PHASE 8 — `nano`

## Objective

Implement the first unique numeric feature of QwicLang.

## Deliverables

* `nano` type
* nano literal parsing
* assignment
* arithmetic
* comparisons
* native code generation

Example:

```qwic
public func main() {
    const delta: nano = 0.016666
    const value: nano = delta * 2

    print(value)
}
```

The internal representation must be documented.

No undocumented floating-point conversions.

---

# PHASE 9 — Modules and Visibility

## Objective

Introduce multi-file compilation and enforce API boundaries.

## Deliverables

Support:

```text
module
import
public
private
```

Example:

```qwic
// users.qw

private func validateUser() {
}

public func createUser() {
}
```

And:

```qwic
// main.qw

import users

public func main() {
    users.createUser()
}
```

External access to private functions must fail at compile time.

---

# PHASE 10 — Runtime Foundation

## Objective

Introduce a minimal Qwic runtime.

## Deliverables

* runtime initialization
* printing
* basic memory helpers
* process exit handling
* basic platform abstraction

The runtime must remain small.

Do not build networking or concurrency yet.

---

# PHASE 11 — CLI and Developer Experience

## Objective

Make the compiler pleasant to use.

Implement:

```bash
qwic build
qwic run
qwic check
qwic fmt
qwic clean
```

Add:

* useful error messages
* stable exit codes
* source file reporting
* output paths
* build diagnostics

---

# PHASE 12 — v0 Release

## Objective

Produce the first usable QwicLang compiler.

## v0 must support

```text
variables
const
let
basic types
nano
functions
function calls
return
if/else
while
strings
arithmetic
comparisons
comments
optional semicolons
optional indentation
basic public/private
basic modules
native executable generation
```

## v0 must demonstrate

### Example 1 — Hello World

```qwic
public func main() {
    print("Hello, Qwic")
}
```

### Example 2 — Functions

```qwic
func add(a: int, b: int): int {
    return a + b
}

public func main() {
    print(add(10, 20))
}
```

### Example 3 — Precision

```qwic
public func main() {
    const delta: nano = 0.016666

    print(delta)
}
```

### Example 4 — Control Flow

```qwic
public func main() {
    let i: int = 0

    while i < 5 {
        print(i)
        i = i + 1
    }
}
```

---

# 14. Definition of Done

A phase is not complete simply because the code compiles.

A phase is complete when:

1. implementation exists
2. tests exist
3. documentation exists where needed
4. error handling works
5. existing tests still pass
6. code is formatted
7. code is readable
8. no known dead code remains
9. examples demonstrate the feature
10. the feature works end-to-end where applicable

---

# 15. Testing Strategy

Testing must occur at multiple levels.

## Unit Tests

Test individual compiler components.

Example:

```text
lexer_test.go
parser_test.go
types_test.go
sema_test.go
codegen_test.go
```

---

## Integration Tests

Compile complete Qwic programs.

Example:

```text
tests/
    hello/
    arithmetic/
    functions/
    control_flow/
    nano/
```

Each test should include:

```text
source.qw
expected output
```

---

## Negative Tests

Compiler errors are just as important as successful programs.

Example:

```text
invalid_assignment.qw
invalid_return.qw
unknown_variable.qw
private_access.qw
```

Each should verify the expected diagnostic.

---

# 16. Regression Rule

Before every major change:

```bash
go test ./...
```

must pass.

After code generation changes:

```bash
go test ./...
```

plus the complete integration test suite must pass.

Never knowingly merge a regression.

---

# 17. Documentation Requirements

Every significant language feature must be documented.

Documentation should explain:

1. what the feature does
2. why it exists
3. syntax
4. examples
5. limitations
6. implementation notes where relevant

Do not write documentation that describes features which are not implemented.

Clearly label experimental or planned features.

---

# 18. Code Review Rules for Claude

Before considering a change complete, Claude must inspect:

* package boundaries
* naming
* duplicated logic
* error propagation
* tests
* comments
* unreachable code
* unused variables
* unnecessary abstractions
* public APIs
* consistency with existing architecture

Claude should prefer a smaller clean implementation over a larger “future-proof” implementation.

---

# 19. Dependency Policy

Dependencies should be minimized.

Before introducing a dependency:

1. determine whether the Go standard library is sufficient
2. determine whether the functionality is simple enough to implement locally
3. determine whether the dependency is mature and maintained
4. determine whether it is necessary for the current phase

Do not introduce dependencies merely for convenience.

The compiler itself should remain lightweight.

---

# 20. Error Handling Rules

Never silently ignore errors.

Bad:

```go
data, _ := os.ReadFile(path)
```

Preferred:

```go
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("read source file: %w", err)
}
```

Compiler errors should be propagated with useful context.

---

# 21. Logging and Debugging

Do not litter production compiler code with uncontrolled debug output.

Avoid:

```go
fmt.Println("HERE")
fmt.Println(node)
```

Use structured debugging helpers or explicit debug flags.

Example:

```bash
qwic build --debug
```

where appropriate.

---

# 22. Performance Policy

Performance is important, but premature optimization is forbidden.

First make the compiler:

```text
correct
readable
testable
```

Then optimize.

The first optimization priorities are:

1. compiler startup
2. repeated parsing
3. memory allocations
4. code-generation efficiency
5. generated binary performance

Do not optimize compiler internals based on assumptions.

Measure first.

---

# 23. Security Policy

Compiler input is untrusted input.

Treat:

* source files
* module files
* dependency metadata
* paths
* generated code
* external compiler output

as potentially malformed.

Never assume source input is valid.

Avoid unsafe shell construction.

Never execute generated code implicitly without an explicit compiler command such as:

```bash
qwic run
```

---

# 24. Generated Code Rules

If the compiler temporarily generates C or LLVM IR:

* generated files must be deterministic
* generated code should be readable enough to debug
* temporary files must be cleaned up
* errors from the external compiler must be captured and reported
* paths must be handled safely
* generated output must never be committed unless explicitly required

---

# 25. Git Practices

Use small, logically focused commits.

Preferred commit style:

```text
lexer: add string literal support
parser: parse function declarations
sema: add assignment type checking
codegen: emit integer arithmetic
cli: add qwic run
```

Avoid commits such as:

```text
stuff
changes
update
fix
more work
```

---

# 26. No Premature Self-Hosting

Do not attempt to rewrite the compiler in QwicLang before v0 is stable.

The intended sequence is:

```text
Go compiler
    ↓
functional QwicLang v0
    ↓
stable compiler architecture
    ↓
richer QwicLang
    ↓
Qwic compiler written in QwicLang
```

Self-hosting is a future milestone.

---

# 27. No Premature Assembly

Assembly is not part of the first compiler implementation.

The priority is:

```text
Qwic
 ↓
compiler
 ↓
native executable
```

Assembly may later be introduced into the runtime or optimized standard-library implementations.

Never add assembly merely because it sounds faster.

Use it when measured performance justifies it.

---

# 28. No Fake Features

Claude MUST NOT claim that a feature is implemented when it is only:

* parsed
* represented in the AST
* stubbed
* ignored
* treated as a no-op
* partially implemented

For example, if `turbo` currently has no special optimization:

Documentation should say so.

Do not claim:

> Qwic automatically generates SIMD machine code.

unless that functionality actually exists and is tested.

---

# 29. Backward Compatibility

During v0 development, the language is experimental and syntax may change.

However, avoid changing syntax casually.

Before changing an existing construct:

1. identify existing usage
2. determine why it needs changing
3. update tests
4. update examples
5. update documentation
6. document the change

---

# 30. Development Workflow

For every task, Claude should follow this workflow:

```text
1. Understand the requirement
2. Inspect the existing architecture
3. Identify affected components
4. Make the smallest appropriate design
5. Implement
6. Add tests
7. Run tests
8. Run formatting/linting
9. Verify integration
10. Update documentation
```

Do not immediately start writing code before understanding the relevant existing components.

---

# 31. When Something Fails

When encountering compiler/runtime problems:

1. reproduce the issue
2. identify the failing phase
3. create a minimal test case
4. fix the root cause
5. add a regression test
6. rerun the full suite

Do not patch symptoms repeatedly.

---

# 32. Compiler Debugging Mode

The compiler should eventually provide debug output for:

```text
tokens
AST
semantic information
IR
generated code
```

Possible commands:

```bash
qwic debug tokens file.qw
qwic debug ast file.qw
qwic debug ir file.qw
```

These are development tools and not required for the first milestone.

---

# 33. Example v0 End-to-End Flow

Source:

```qwic
public func add(a: int, b: int): int {
    return a + b
}

public func main() {
    const result = add(20, 22)
    print(result)
}
```

Compiler:

```text
main.qw
   ↓
Lexer
   ↓
Parser
   ↓
AST
   ↓
Semantic Analysis
   ↓
Type Checking
   ↓
IR
   ↓
Code Generator
   ↓
Native executable
```

Execution:

```bash
qwic run main.qw
```

Output:

```text
42
```

This is a successful v0 milestone.

---

# 34. First Milestone Checklist

Claude should consider the first milestone successful only when all of the following work:

* [ ] `go build ./...`
* [ ] `go test ./...`
* [ ] `qwic --help`
* [ ] Qwic source file can be read
* [ ] lexer works
* [ ] parser works
* [ ] AST works
* [ ] semantic analysis works
* [ ] basic type checking works
* [ ] IR generation works
* [ ] native code generation works
* [ ] executable runs
* [ ] `print()` works
* [ ] variables work
* [ ] functions work
* [ ] arithmetic works
* [ ] conditions work
* [ ] loops work
* [ ] `nano` works at a basic level
* [ ] compiler errors contain source locations
* [ ] integration tests pass

---

# 35. v0 Success Definition

QwicLang v0 is successful when a developer can clone the repository and execute:

```bash
go build ./cmd/qwic
```

then:

```bash
qwic run examples/hello.qw
```

and receive:

```text
Hello, Qwic
```

without needing to manually manipulate compiler internals.

A second program must demonstrate actual computation:

```qwic
public func add(a: int, b: int): int {
    return a + b
}

public func main() {
    print(add(20, 22))
}
```

and produce:

```text
42
```

---

# 36. Long-Term Architecture

The intended long-term architecture is:

```text
                     QwicLang
                         |
                +--------+--------+
                |                 |
             Frontend           Tools
                |                 |
        +-------+-------+     +---+---+
        |       |       |     |       |
      Lexer   Parser   Sema   fmt    test
                |
               AST
                |
               HIR
                |
               IR
                |
          +-----+------+
          |            |
       Normal        Turbo
          |            |
          +-----+------+
                |
             Optimizer
                |
              LLVM
                |
        +-------+-------+
        |       |       |
      x86-64  ARM64   RISC-V
        |
      Runtime
        |
      Native OS
```

---

# 37. Future Runtime Goals

After v0 is stable, development can move toward:

```text
stack-first allocation
escape analysis
lightweight scheduler
work stealing
async I/O
zero-copy networking
arenas
object pools
thread-local storage
per-thread allocators
SIMD
CPU feature detection
profile-guided optimization
```

These are future goals, not v0 requirements.

---

# 38. Future Language Goals

After v0:

```text
generics
interfaces
Result / error model
channels
spawn / await
modules
package manager
FFI
advanced memory model
unsafe blocks
turbo optimization
standard library
language server
formatter
documentation generator
benchmarking tools
```

---

# 39. Final Rule for Claude

When choosing between:

```text
A complicated implementation with more features
```

and:

```text
A smaller implementation that is correct, readable, tested, and easy to extend
```

**choose the second.**

QwicLang v0 is a foundation.

The goal is not to build the entire language immediately.

The goal is to build a **small compiler that works**, with an architecture clean enough that the language can grow without becoming difficult to understand.

The first victory is simple:

```text
Qwic source
    ↓
Qwic compiler
    ↓
native executable
    ↓
working program
```

Everything else comes afterward.
