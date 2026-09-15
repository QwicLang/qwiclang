# QwicLang

## v0.1.0-alpha.1 Language Documentation

**Fast. Native. Precise.**

---

# 1. Welcome to QwicLang

QwicLang is a lightweight, compiled programming language designed around a simple idea:

> **Write clean code. Compile quickly. Run natively. Scale without unnecessary complexity.**

QwicLang is designed to combine the readability and approachability of modern languages with the performance and predictability expected from systems programming.

The language takes inspiration from several established languages, including:

* **Python** — readability and developer friendliness
* **JavaScript / TypeScript** — familiar modern syntax
* **Go** — simplicity, fast compilation, and scalable systems
* **Rust** — safety and performance-oriented design

QwicLang does **not** attempt to become a copy of any of these languages.

Instead, it asks:

> **What are the ideas worth keeping, and how can they work together in a smaller, cleaner language?**

---

# 2. The Philosophy of QwicLang

QwicLang is built around several core principles.

## 2.1 Simple Code

Programming languages can become difficult not because programming itself is difficult, but because the language contains too many rules.

QwicLang tries to keep the number of rules small.

For example:

```qwic
public func main() {
    const name: string = "QwicLang"

    print(name)
}
```

The code should be understandable without requiring the developer to learn a large amount of syntax first.

---

# 2.2 Readability Matters

QwicLang code should be easy to read months after it was written.

This means:

* clear keywords
* explicit declarations
* predictable syntax
* minimal punctuation
* meaningful compiler errors
* no unnecessary ceremony

For example:

```qwic
const age: int = 25
```

is deliberately straightforward.

---

# 2.3 Native by Design

QwicLang is a compiled language.

A QwicLang program is transformed into native machine code rather than requiring a virtual machine or interpreter for normal execution.

Conceptually:

```text
Qwic Source
     ↓
Compiler
     ↓
Intermediate Representation
     ↓
Native Code
     ↓
Executable
```

The goal is to produce programs that execute directly on the target platform.

---

# 2.4 Fast Compilation

Compilation speed is a first-class design goal.

A language should not require developers to wait unnecessarily before testing a change.

QwicLang therefore favors:

* simple syntax
* predictable parsing
* efficient compiler architecture
* small compiler passes
* incremental compilation in future versions
* minimal runtime dependencies

Fast compilation is not merely a convenience.

It improves the entire development experience.

---

# 2.5 Small Builds

QwicLang aims to produce small, efficient native programs.

The language should avoid requiring a large runtime simply to execute a basic program.

A simple program should remain simple after compilation.

---

# 2.6 Safety Without Excessive Complexity

QwicLang is intended to provide safe defaults while avoiding a language design that requires developers to fight the compiler.

Safety should come from:

* strong typing
* clear visibility
* predictable memory behavior
* explicit APIs
* compiler validation
* controlled unsafe functionality in future versions

Safety is a design property, not merely a collection of restrictions.

---

# 2.7 Performance Should Be Earned

QwicLang does not assume that adding complicated features automatically makes programs faster.

Performance should come from:

* good language design
* efficient generated code
* optimized runtime primitives
* predictable data representation
* compiler optimization
* measured improvements

QwicLang follows a simple rule:

> **Measure first. Optimize second.**

---

# 2.8 Developer Experience

Performance is important, but developers should not have to sacrifice usability to obtain it.

QwicLang aims to provide:

```text
Readable source
       +
Fast compiler
       +
Native execution
       +
Useful tooling
```

---

# 3. What Is QwicLang?

QwicLang is a:

* statically typed language
* compiled language
* native language
* general-purpose language
* systems-oriented language
* scalable application language

Potential applications include:

* backend services
* networking
* APIs
* command-line applications
* infrastructure
* game development
* real-time applications
* developer tools
* system utilities

---

# 4. Your First QwicLang Program

A basic QwicLang program looks like this:

```qwic
public func main() {
    print("Hello, QwicLang")
}
```

Let's break it down.

### `public`

Declares that the function is publicly accessible.

### `func`

Declares a function.

### `main`

The entry point of the application.

### `()`

The function accepts no parameters.

### `{ }`

The braces define the function body.

### `print(...)`

Prints a value to standard output.

---

# 5. Running a QwicLang Program

A QwicLang source file uses the:

```text
.qw
```

extension.

For example:

```text
hello.qw
```

The compiler is called:

```bash
qwic
```

Compile:

```bash
qwic build hello.qw
```

Run:

```bash
qwic run hello.qw
```

The expected result is:

```text
Hello, QwicLang
```

---

# 6. Syntax Philosophy

QwicLang syntax is intentionally influenced by modern languages such as JavaScript and TypeScript rather than closely following Go syntax.

For example:

```qwic
const name: string = "Tanaka"
let count: int = 0
```

The syntax should feel familiar to developers who have used TypeScript, JavaScript, Python, Go, or similar languages.

---

# 7. Statements and Semicolons

Semicolons are optional.

Both are valid.

Without semicolons:

```qwic
const name: string = "Qwic"
const age: int = 1
print(name)
```

With semicolons:

```qwic
const name: string = "Qwic";
const age: int = 1;
print(name);
```

QwicLang does not require developers to use semicolons.

The recommended style is:

```qwic
const name: string = "Qwic"
const age: int = 1
```

---

# 8. Indentation

Indentation is optional.

Braces determine blocks.

This is valid:

```qwic
public func main() {
    const value: int = 10

    if value > 5 {
        print(value)
    }
}
```

So is:

```qwic
public func main() {
const value: int = 10
if value > 5 {
print(value)
}
}
```

However, QwicLang strongly recommends indentation for readability.

The compiler cares about structure.

Humans care about indentation.

QwicLang allows both.

---

# 9. Comments

Single-line comments use `//`.

```qwic
// This is a comment
const age: int = 25
```

Multi-line comments use:

```qwic
/*
    This is a
    multi-line comment.
*/
```

Comments are ignored by the compiler.

---

# 10. Variables

QwicLang provides two primary variable declarations:

```text
const
let
```

---

# 11. `const`

`const` creates a value that cannot be reassigned.

```qwic
const name: string = "QwicLang"
const age: int = 1
```

Attempting to reassign it is invalid:

```qwic
const age: int = 1

age = 2
```

The compiler should report an error.

Use `const` when a value should remain unchanged.

---

# 12. `let`

`let` creates a mutable variable.

```qwic
let score: int = 0

score = 10
score = 20
```

Use `let` when the value needs to change.

---

# 13. Type Annotations

A type can be explicitly declared using `:`.

```qwic
const age: int = 25
const name: string = "Qwic"
const active: bool = true
```

The general form is:

```text
name: type
```

---

# 14. Type Inference

QwicLang can infer the type of a value when it is obvious.

For example:

```qwic
const age = 25
const name = "QwicLang"
const active = true
```

The compiler can determine:

```text
age    → int
name   → string
active → bool
```

Explicit types can still be used when clarity is preferred:

```qwic
const age: int = 25
```

---

# 15. Primitive Types

QwicLang v0 defines several basic types.

| Type     | Purpose                         |
| -------- | ------------------------------- |
| `int`    | Whole numbers                   |
| `float`  | General floating-point numbers  |
| `nano`   | Precise small fractional values |
| `bool`   | True/false values               |
| `string` | Text                            |
| `void`   | No returned value               |

---

# 16. `int`

`int` represents whole numbers.

```qwic
const age: int = 25
const score: int = 100
const temperature: int = -5
```

Arithmetic is supported:

```qwic
const a: int = 10
const b: int = 20

const result: int = a + b
```

---

# 17. `float`

`float` represents general floating-point numbers.

```qwic
const price: float = 19.95
const ratio: float = 0.75
```

Floating-point values are intended for general numerical calculations where standard floating-point behavior is appropriate.

---

# 18. `nano`

`nano` is one of QwicLang's specialized numeric types.

It is designed for **precise small fractional values**.

For example:

```qwic
const frameTime: nano = 0.016666
```

Another example:

```qwic
const interval: nano = 0.00012121
```

The purpose of `nano` is to make values such as these natural to represent:

```text
0.016666
0.001
0.00012121
```

Typical use cases include:

* game timing
* frame timing
* animation
* simulation
* clocks
* intervals
* interpolation
* precise calculations

---

# 19. Why `nano` Exists

Traditional floating-point values are useful, but some applications repeatedly deal with small fractional quantities where predictable precision is important.

Game development is a good example.

Suppose a game runs at approximately 60 frames per second.

The duration of one frame is approximately:

```text
0.016666
```

A QwicLang program can represent this directly:

```qwic
const frameTime: nano = 0.016666
```

There is deliberately no unit suffix.

QwicLang does **not** use:

```text
16.666ms
0.016666s
```

The value itself is unitless.

The meaning of the value is determined by the API or context in which it is used.

For example:

```qwic
const frameTime: nano = 0.016666
```

A game engine may define this value as seconds.

Another system may use the same numeric representation for another timing purpose.

---

# 20. `bool`

Boolean values are:

```qwic
true
false
```

Example:

```qwic
const loggedIn: bool = true
const gameOver: bool = false
```

Booleans are commonly used with conditions:

```qwic
if loggedIn {
    print("Welcome")
}
```

---

# 21. `string`

Strings contain text.

```qwic
const name: string = "QwicLang"
```

Strings can be passed to functions:

```qwic
print("Hello")
```

String support will expand as the standard library develops.

---

# 22. `void`

`void` indicates that a function does not return a value.

```qwic
public func greet(): void {
    print("Hello")
}
```

The return type can be omitted where the language can infer that no value is returned, depending on the final v0 implementation.

---

# 23. Operators

QwicLang supports familiar operators.

## Arithmetic

```text
+
-
*
/
%
```

Example:

```qwic
const a: int = 20
const b: int = 5

const sum = a + b
const difference = a - b
const product = a * b
const quotient = a / b
const remainder = a % b
```

---

# 24. Comparison Operators

Comparison operators include:

```text
==
!=
>
<
>=
<=
```

Example:

```qwic
const age: int = 25

if age >= 18 {
    print("Adult")
}
```

---

# 25. Boolean Operators

QwicLang uses familiar logical operators.

```text
&&
||
!
```

Example:

```qwic
if age >= 18 && active {
    print("Allowed")
}
```

---

# 26. Functions

Functions are declared using `func`.

Basic function:

```qwic
func greet() {
    print("Hello")
}
```

Functions can accept parameters:

```qwic
func greet(name: string) {
    print(name)
}
```

---

# 27. Function Visibility

QwicLang supports:

```text
public
private
```

A public function:

```qwic
public func greet() {
    print("Hello")
}
```

A private function:

```qwic
private func validateUser() {
}
```

---

# 28. Public Functions

A public function is accessible outside its module.

```qwic
public func createUser() {
}
```

Public functions form part of a module's external API.

---

# 29. Private Functions

A private function is only accessible within its defining module.

```qwic
private func validateUser() {
}
```

Another module cannot call it.

This allows developers to separate:

```text
Public API
    ↓
Implementation details
```

---

# 30. Functions With Return Values

Functions can specify their return type.

```qwic
func add(a: int, b: int): int {
    return a + b
}
```

The function returns an `int`.

Usage:

```qwic
const result = add(10, 20)
print(result)
```

Output:

```text
30
```

---

# 31. The `return` Statement

`return` exits a function.

```qwic
func square(value: int): int {
    return value * value
}
```

A function returning `void` does not need to return a value.

---

# 32. The `main` Function

A QwicLang executable starts from:

```qwic
public func main() {
}
```

For example:

```qwic
public func main() {
    print("Application started")
}
```

The compiler uses `main` as the application entry point.

---

# 33. Conditions

Conditions use `if`.

```qwic
if score > 100 {
    print("High score")
}
```

---

# 34. `else`

An alternative branch can be provided using `else`.

```qwic
if score >= 50 {
    print("Pass")
} else {
    print("Fail")
}
```

---

# 35. Multiple Conditions

Conditions can be chained.

```qwic
if score >= 80 {
    print("Excellent")
} else if score >= 50 {
    print("Pass")
} else {
    print("Fail")
}
```

---

# 36. `while` Loops

QwicLang v0 supports `while`.

```qwic
let count: int = 0

while count < 5 {
    print(count)
    count = count + 1
}
```

The loop continues while the condition is true.

Output:

```text
0
1
2
3
4
```

---

# 37. Scope

Variables exist within the scope where they are declared.

```qwic
public func main() {
    const name: string = "Qwic"

    if true {
        const message: string = "Hello"
        print(message)
    }

    print(name)
}
```

`message` belongs to the inner block.

Scope prevents unrelated parts of a program from accidentally accessing internal variables.

---

# 38. Modules

QwicLang supports modules to divide programs into multiple files.

A module can expose selected functionality while keeping implementation details private.

For example:

```text
project/
    main.qw
    users.qw
```

`users.qw`:

```qwic
public func createUser() {
    validateUser()
}

private func validateUser() {
}
```

The public API is:

```text
createUser
```

while:

```text
validateUser
```

remains internal.

---

# 39. Imports

A module can import another module.

Conceptually:

```qwic
import users
```

Then public functionality can be accessed.

```qwic
import users

public func main() {
    users.createUser()
}
```

Private functionality must remain inaccessible:

```qwic
users.validateUser()
```

This should produce a compile-time error.

---

# 40. Visibility Philosophy

QwicLang treats module boundaries as important.

A module should expose only what other modules need.

This creates a clean distinction:

```text
                 Module
        ┌──────────────────────┐
        │                      │
        │  public API          │
        │                      │
        │  ────────────────    │
        │                      │
        │  private internals   │
        │                      │
        └──────────────────────┘
```

The goal is to make large systems easier to maintain.

---

# 41. `turbo`

QwicLang introduces the concept of a `turbo` function.

Example:

```qwic
turbo func calculate(value: int): int {
    return value * 2
}
```

The developer writes normal QwicLang syntax.

The compiler can identify the function as a candidate for aggressive optimization.

The long-term goal is:

```text
Qwic source
     ↓
turbo function
     ↓
compiler optimization
     ↓
highly optimized native machine code
```

---

# 42. Turbo Is a Compiler Feature

`turbo` is not intended to create a different programming syntax.

The developer still writes:

```qwic
turbo func calculate(value: int): int {
    return value * 2
}
```

rather than writing assembly.

The compiler is responsible for deciding how the function should be optimized.

This keeps the source code readable.

---

# 43. Turbo and Assembly

QwicLang may eventually use architecture-specific assembly in selected situations.

For example:

```text
Qwic
 ↓
Compiler
 ↓
Turbo analysis
 ↓
Optimization
 ↓
Machine code
```

Assembly may exist inside the compiler/runtime implementation where it provides a measurable benefit.

Developers should not need to write assembly for normal turbo functions.

---

# 44. The Runtime

QwicLang uses a runtime layer for functionality that cannot be expressed entirely through generated application code.

The runtime may eventually provide:

* process management
* memory primitives
* I/O
* concurrency
* scheduling
* networking
* timing
* operating-system integration

The runtime should remain small.

---

# 45. Performance Model

QwicLang aims for a simple performance model.

A developer should be able to reason about code without needing to understand a huge runtime.

The compiler should optimize where possible.

The runtime should avoid unnecessary overhead.

The language should make expensive operations identifiable.

---

# 46. Compilation Architecture

The QwicLang compiler is designed around multiple stages.

```text
Source Code
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
Optimization
     ↓
Code Generation
     ↓
Native Executable
```

Each stage has a specific responsibility.

---

# 47. Lexer

The lexer converts characters into tokens.

For:

```qwic
const age: int = 25
```

the lexer conceptually produces:

```text
CONST
IDENTIFIER(age)
COLON
IDENTIFIER(int)
EQUAL
INTEGER(25)
```

The lexer does not decide whether the program is logically correct.

---

# 48. Parser

The parser takes tokens and determines their structure.

For example:

```qwic
const age: int = 25
```

becomes a variable declaration in the AST.

Conceptually:

```text
VariableDeclaration
├── name: age
├── type: int
└── value: 25
```

---

# 49. AST

AST means **Abstract Syntax Tree**.

It represents the structure of the program.

For:

```qwic
const result = add(10, 20)
```

the tree could conceptually look like:

```text
VariableDeclaration
│
├── name: result
│
└── CallExpression
    │
    ├── function: add
    ├── argument: 10
    └── argument: 20
```

---

# 50. Semantic Analysis

Parsing determines whether code has valid structure.

Semantic analysis determines whether the program makes sense.

For example:

```qwic
const age: int = "hello"
```

has valid syntax.

But it has invalid semantics because:

```text
int ≠ string
```

The compiler should reject it.

---

# 51. Type Checking

QwicLang is statically typed.

The compiler checks types before the program runs.

For example:

```qwic
const age: int = 25
```

is valid.

But:

```qwic
const age: int = "twenty five"
```

is invalid.

The compiler should detect this before producing the executable.

---

# 52. Intermediate Representation

The compiler uses an intermediate representation, commonly called IR.

IR separates the language frontend from the machine-code backend.

Conceptually:

```text
Qwic Syntax
     ↓
AST
     ↓
Semantic Model
     ↓
IR
     ↓
Backend
```

This allows QwicLang to eventually support multiple CPU architectures without rewriting the language frontend.

Potential targets include:

```text
x86-64
ARM64
RISC-V
```

---

# 53. Native Compilation

The ultimate output of the compiler is native machine code.

A simplified model:

```text
hello.qw
   ↓
Qwic Compiler
   ↓
Native Object Code
   ↓
Linker
   ↓
hello
```

The resulting application can run directly on the target operating system.

---

# 54. Error Messages

Compiler errors should be useful.

Instead of:

```text
syntax error
```

QwicLang aims for diagnostics such as:

```text
error: cannot assign a string to an int

 --> main.qw:4:5
  |
4 | age = "hello"
  |       ^^^^^^^
```

Good diagnostics are part of the language design.

---

# 55. Error Philosophy

The compiler should tell developers:

1. what went wrong
2. where it happened
3. why it happened
4. where possible, how to fix it

Compiler errors are part of the developer experience.

---

# 56. Example: A Complete Small Program

```qwic
public func add(a: int, b: int): int {
    return a + b
}

public func main() {
    const first: int = 20
    const second: int = 22

    const result = add(first, second)

    print(result)
}
```

Output:

```text
42
```

This small example demonstrates:

* functions
* public visibility
* parameters
* return values
* constants
* type annotations
* type inference
* function calls

---

# 57. Example: Game Timing

`nano` is particularly useful for real-time applications.

```qwic
public func update(delta: nano) {
    // Update game state
}
```

A game loop might conceptually work with:

```qwic
const delta: nano = 0.016666

update(delta)
```

The value does not contain a unit suffix.

The API defines what the value represents.

---

# 58. Example: Mutable State

```qwic
public func main() {
    let score: int = 0

    score = score + 10
    score = score + 20

    print(score)
}
```

Output:

```text
30
```

---

# 59. Example: Constants

```qwic
const maxPlayers: int = 16
const gameName: string = "Arena"
const frameTime: nano = 0.016666
```

Constants are ideal for values that should not change.

---

# 60. Example: Validation

```qwic
private func isValidAge(age: int): bool {
    return age >= 18
}

public func registerUser(age: int) {
    if isValidAge(age) {
        print("User accepted")
    } else {
        print("User rejected")
    }
}
```

The validation function is an implementation detail.

The registration function is part of the public API.

---

# 61. Naming

QwicLang encourages clear names.

Prefer:

```qwic
const playerScore: int = 100
```

over:

```qwic
const x: int = 100
```

unless `x` has an obvious mathematical meaning.

Functions should describe actions:

```qwic
createUser()
calculateScore()
loadGame()
```

Variables should describe data:

```qwic
playerScore
username
frameTime
```

---

# 62. Recommended Formatting

QwicLang code should generally use four spaces for indentation.

Recommended:

```qwic
public func main() {
    const name: string = "QwicLang"

    if name == "QwicLang" {
        print("Hello")
    }
}
```

Although indentation is not syntactically required, consistent formatting is strongly encouraged.

---

# 63. What QwicLang Is Not

QwicLang is not intended to be:

* another JavaScript clone
* another Python clone
* another Go clone
* a replacement for every programming language
* an intentionally complicated systems language
* a language that requires assembly to achieve performance

QwicLang is its own language.

Its inspirations are references, not constraints.

---

# 64. Design Trade-Offs

Every language makes trade-offs.

QwicLang currently prioritizes:

```text
simplicity
    ↓
readability
    ↓
compile speed
    ↓
native performance
    ↓
scalability
```

Not every advanced language feature belongs in the language.

A feature should justify its complexity.

---

# 65. Why Not Put Everything in the Language?

A language can become difficult when every problem is solved by adding syntax.

QwicLang prefers to keep the language core small.

For example:

```text
Language
   ↓
Small core
   ↓
Standard library
   ↓
Libraries
   ↓
Applications
```

This makes the compiler easier to understand and potentially faster to compile.

---

# 66. Scalability

QwicLang is intended for software that grows.

This means scalability is not only about runtime performance.

It also means:

* manageable codebases
* clear module boundaries
* explicit APIs
* fast builds
* predictable behavior
* good tooling

A language should help a project remain understandable as it becomes larger.

---

# 67. Game Development

QwicLang has particular potential for game development.

Important characteristics include:

* native execution
* fast compilation
* predictable numerical representation
* `nano` timing values
* low runtime overhead
* future SIMD support
* future concurrency primitives
* future platform APIs

For example:

```qwic
public func update(delta: nano) {
    player.update(delta)
    enemies.update(delta)
    physics.update(delta)
}
```

The goal is to make real-time code readable without hiding performance characteristics.

---

# 68. Systems Programming

QwicLang can also target systems-oriented applications.

Potential areas include:

```text
network services
databases
command-line tools
servers
infrastructure
system utilities
embedded applications
```

The native compilation model makes these possible without requiring an interpreter.

---

# 69. Concurrency — Future Direction

Concurrency is an important part of QwicLang's long-term design.

Future versions may introduce constructs such as:

```text
spawn
await
channels
```

However, these features are intentionally not part of the minimal v0 language unless their runtime semantics are properly defined.

The philosophy is:

> **Do not add concurrency syntax before there is a sound concurrency model behind it.**

---

# 70. Memory Management — Future Direction

QwicLang's long-term memory model is expected to prioritize:

* predictable performance
* safety
* low overhead
* developer usability

Potential future mechanisms include:

```text
stack allocation
escape analysis
arenas
object pools
thread-local allocation
controlled unsafe operations
```

The exact model should be established through implementation and benchmarking rather than copied directly from another language.

---

# 71. Standard Library

The QwicLang standard library will eventually provide functionality for:

```text
strings
collections
files
networking
processes
time
concurrency
math
randomness
encoding
```

The standard library should remain focused.

Not every functionality needs to become part of the core language.

---

# 72. Tooling

The primary compiler command is:

```bash
qwic
```

The intended command structure includes:

```bash
qwic build
qwic run
qwic check
qwic fmt
qwic clean
```

Future developer tools may include:

```bash
qwic test
qwic debug
qwic doc
```

---

# 73. `qwic build`

Builds a QwicLang program.

```bash
qwic build main.qw
```

The compiler produces a native executable.

---

# 74. `qwic run`

Compiles and runs a program.

```bash
qwic run main.qw
```

This is useful during development.

---

# 75. `qwic check`

Checks source code without necessarily producing a final executable.

```bash
qwic check main.qw
```

This can eventually provide fast feedback during development.

---

# 76. `qwic fmt`

The formatter will eventually provide a canonical QwicLang style.

For example:

```bash
qwic fmt main.qw
```

The formatter should make QwicLang code consistently readable regardless of the developer's personal formatting preferences.

---

# 77. Project Structure

A QwicLang project may eventually look like:

```text
my-project/
│
├── qwic.toml
│
├── src/
│   ├── main.qw
│   ├── users.qw
│   └── database.qw
│
├── tests/
│
└── build/
```

The exact package/project format may evolve during v0.

---

# 78. QwicLang v0 Scope

v0 is intentionally small.

The initial implementation focuses on:

```text
✓ variables
✓ const
✓ let
✓ primitive types
✓ nano
✓ functions
✓ parameters
✓ return values
✓ function calls
✓ if / else
✓ while
✓ expressions
✓ strings
✓ comments
✓ public
✓ private
✓ basic modules
✓ imports
✓ optional semicolons
✓ optional indentation
✓ native compilation
```

---

# 79. Features Beyond v0

The following are future areas rather than requirements for the initial language:

```text
generics
interfaces
advanced ownership
concurrency
async I/O
channels
FFI
SIMD
advanced optimization
versioned package registry and lockfiles
language server
debugger
reflection
macros
GPU programming
self-hosting
```

These features should only be added when they fit QwicLang's philosophy.

---

# 80. Learning QwicLang

A new developer should learn QwicLang in roughly this order:

### Step 1

Understand:

```text
const
let
```

### Step 2

Learn types:

```text
int
float
nano
bool
string
```

### Step 3

Learn expressions:

```text
+
-
*
/
==
!=
>
<
```

### Step 4

Learn functions:

```qwic
func
```

### Step 5

Learn control flow:

```qwic
if
else
while
```

### Step 6

Learn visibility:

```qwic
public
private
```

### Step 7

Learn modules:

```qwic
import
```

### Step 8

Learn specialized features:

```text
nano
turbo
```

This progression intentionally starts with the simplest concepts.

---

# 81. A Beginner's Complete Example

Here is a small program using many of the core concepts:

```qwic
private func isAdult(age: int): bool {
    return age >= 18
}

public func greet(name: string) {
    print("Hello, " + name)
}

public func main() {
    const name: string = "Developer"
    let age: int = 21

    greet(name)

    if isAdult(age) {
        print("You are an adult")
    } else {
        print("You are under 18")
    }
}
```

The program demonstrates:

* private functions
* public functions
* parameters
* return values
* strings
* integers
* constants
* mutable variables
* conditions
* function calls
* the application entry point

---

# 82. QwicLang in One Example

A future QwicLang application should be able to look something like:

```qwic
import users

private func calculateFrameTime(): nano {
    return 0.016666
}

public func update() {
    const delta: nano = calculateFrameTime()

    users.update(delta)
}

public func main() {
    update()
}
```

Notice the design:

* syntax remains readable
* semicolons are unnecessary
* indentation is readable but not syntactically significant
* visibility is explicit
* timing values are represented naturally
* functions look familiar
* the compiler handles native execution

---

# 83. The QwicLang Mental Model

When learning QwicLang, think about the language in four layers.

## Layer 1 — Language

This is what you write:

```qwic
public func add(a: int, b: int): int {
    return a + b
}
```

## Layer 2 — Compiler

The compiler understands:

```text
syntax
types
visibility
control flow
optimization
```

## Layer 3 — Runtime

The runtime provides things that applications need:

```text
I/O
memory
time
processes
concurrency
```

## Layer 4 — Machine

Ultimately the program becomes:

```text
machine code
```

The developer writes the first layer.

QwicLang handles the rest.

---

# 84. The QwicLang Promise

QwicLang is built around a simple promise:

> **The language should stay out of your way.**

Code should be:

```text
clean enough to read
simple enough to learn
fast enough to scale
native enough to perform
```

A developer should not need to understand compiler internals to write QwicLang.

But the compiler itself should remain understandable enough that developers interested in language implementation can study it.

---

# 85. Development Philosophy

QwicLang is being developed from the compiler upward.

The initial compiler is deliberately small.

The development path is:

```text
Lexer
  ↓
Parser
  ↓
AST
  ↓
Semantic Analysis
  ↓
Type System
  ↓
IR
  ↓
Code Generation
  ↓
Native Executable
```

Only after this foundation works should advanced runtime and optimization features be introduced.

---

# 86. Correctness Before Performance

QwicLang is intended to be fast.

But the project does not follow:

> "Make everything complicated because complicated code is faster."

Instead:

```text
Correct
   ↓
Simple
   ↓
Measured
   ↓
Optimized
```

The compiler should first produce correct programs.

Then performance should be measured.

Then bottlenecks should be optimized.

---

# 87. Why QwicLang Exists

There are many excellent programming languages.

QwicLang does not need to exist because existing languages are bad.

It exists as an experiment in answering a different question:

> **How small and clean can a modern native programming language be while still being capable of serious software development?**

That is the central experiment behind QwicLang.

---

# 88. The Long-Term Vision

The long-term goal is a language that can comfortably move between:

```text
Small CLI
    ↓
Web service
    ↓
Distributed backend
    ↓
Real-time application
    ↓
Game
    ↓
Systems software
```

without requiring the developer to switch to an entirely different programming model.

The language should remain recognizable as it grows.

---

# 89. QwicLang's Core Principles

The project can ultimately be summarized by ten principles:

### 1. Keep it simple.

Complexity must justify itself.

### 2. Compile fast.

Developer feedback should be quick.

### 3. Run natively.

Performance should not require a virtual machine.

### 4. Keep binaries lean.

A small program should not require a huge runtime.

### 5. Make safety the default.

The compiler should catch problems early.

### 6. Make APIs explicit.

`public` and `private` should clearly communicate boundaries.

### 7. Make precision available.

`nano` exists for workloads where small precise values matter.

### 8. Optimize intelligently.

`turbo` is a compiler concept, not an invitation to write assembly.

### 9. Scale code, not complexity.

Large systems should remain understandable.

### 10. Measure everything.

Performance claims should be backed by benchmarks.

---

# 90. Final Example

The following represents the spirit of QwicLang v0:

```qwic
private func calculateScore(base: int, bonus: int): int {
    return base + bonus
}

public func main() {
    const baseScore: int = 100
    const bonus: int = 50

    const score = calculateScore(baseScore, bonus)

    const frameTime: nano = 0.016666

    if score >= 100 {
        print("High score")
    }

    print(score)
    print(frameTime)
}
```

It is intentionally ordinary.

That is the point.

QwicLang should not require developers to fight the language in order to write performant software.

---

# 91. Quick Reference

## Variables

```qwic
const name: string = "Qwic"
let score: int = 0
```

## Assignment

```qwic
score = 100
```

## Functions

```qwic
func add(a: int, b: int): int {
    return a + b
}
```

## Public Function

```qwic
public func main() {
}
```

## Private Function

```qwic
private func helper() {
}
```

## Condition

```qwic
if condition {
} else {
}
```

## Loop

```qwic
while condition {
}
```

## Boolean

```qwic
const active: bool = true
```

## String

```qwic
const name: string = "Qwic"
```

## Integer

```qwic
const count: int = 10
```

## Floating Point

```qwic
const ratio: float = 0.75
```

## Nano

```qwic
const delta: nano = 0.016666
```

## Import

```qwic
import users
```

## Turbo

```qwic
turbo func calculate(value: int): int {
    return value * 2
}
```

## Comment

```qwic
// comment
```

## Compile

```bash
qwic build main.qw
```

## Run

```bash
qwic run main.qw
```

---

# 92. Closing

QwicLang is intentionally starting small.

The first objective is not to compete with established languages.

The first objective is to build something that is:

**small enough to understand,**

**fast enough to enjoy,**

**safe enough to trust,**

**and powerful enough to grow.**

The language begins with a compiler.

The compiler becomes a platform.

The platform becomes an ecosystem.

And the ecosystem should never lose the simplicity that made QwicLang worth building in the first place.

---

## QwicLang

**Fast. Native. Precise.**
