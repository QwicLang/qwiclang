![Qwic](https://github.com/QwicLang/qwiclang/blob/main/image/readme-banner.png)

<div align="center">

# QwicLang

**A lightweight, readable, statically typed language that compiles to fast native executables.**

[![Version](https://img.shields.io/badge/version-v0.1.0--alpha-blue.svg?style=flat-square)](https://github.com/QwicLang/qwiclang)
[![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen.svg?style=flat-square)](https://github.com/QwicLang/qwiclang)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg?style=flat-square)](https://github.com/QwicLang/qwiclang)
[![Tests](https://img.shields.io/badge/tests-passing-success.svg?style=flat-square)](https://github.com/QwicLang/qwiclang)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-00ADD8.svg?style=flat-square&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENCE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/QwicLang/qwiclang/pulls)

[Quick Start](#-quick-start) •
[Features](#-key-features) •
[Examples](#-code-examples) •
[CLI Reference](#-cli-reference) •
[Architecture](#-architecture) •
[Contributing](#-contributing)

</div>

---

## ⚡ What is QwicLang?

**QwicLang** is an open-source, compiled programming language built on a simple philosophy:

> **Write clean code. Compile quickly. Run natively. Scale without unnecessary complexity.**

High-performance native software shouldn't require cumbersome tooling or convoluted syntax. QwicLang gives you a familiar, expressive syntax inspired by TypeScript and Python, while compiling down to compact, native machine binaries suited for systems programming, networking, simulations, and game development.

```text
Qwic Source (.qw) ➔ Lexer ➔ Parser / AST ➔ Semantic Analysis ➔ IR ➔ Native Executable
```

---

## ✨ Key Features

- 🚀 **Native Machine Code:** Zero VM or interpreter overhead. Programs compile directly to fast, standalone executables.
- 🎯 **Clean & Familiar Syntax:** Elegant declarations with `let` and `const`, block scoping, and optional semicolons.
- 🔬 **High-Precision `nano` Type:** Native fixed-point numeric type tailored for physics, animations, delta times, and simulation math.
- 🔤 **Python-Style F-Strings:** String interpolation with `{expression}` syntax out of the box.
- 📦 **Built-in Collections:** First-class literals and operations for `list`, `set`, `dictionary`, and `tuple`.
- 🌐 **Rich Standard Library:** Built-in modules including `strings`, `time`, `fs`, `sync`, `crypto`, `net`, `json`, and `http` (with TLS support).
- 🔒 **Explicit Visibility:** Module-level boundaries and clear `public` / `private` encapsulation.
- 🛠️ **Batteries-Included CLI:** Everything you need via `qwic build`, `qwic run`, `qwic check`, and `qwic fmt`.

---

## 🚀 Quick Start

### Prerequisites

- [Go 1.22+](https://golang.org/dl/)
- A C compiler (`cc`, `gcc`, or `clang`) available on your `PATH`

### 1. Installation

Clone and compile the `qwic` CLI:

```bash
git clone https://github.com/QwicLang/qwiclang.git
cd qwiclang
go build -o qwic ./cmd/qwic
```

*(Optional)* Move `qwic` to your `PATH`:

```bash
sudo mv qwic /usr/local/bin/
# or: mv qwic ~/go/bin/
```

Verify your installation:

```bash
qwic --help
```

### 2. Hello, World!

Create `hello.qw`:

```qwic
public func main() {
    print("Hello, Qwic!")
}
```

Run directly:

```bash
qwic run hello.qw
```

Or compile to a standalone executable:

```bash
qwic build hello.qw -o hello
./hello
```

---

## 💡 Code Examples

### Functions & Expressions

```qwic
func add(a: int, b: int): int {
    return a + b
}

func multiply(a: int, b: int): int {
    return a * b
}

public func main() {
    const sum = add(20, 22)
    const product = multiply(6, 7)
    print(f"Sum: {sum}, Product: {product}")
}
```

### Control Flow & Loops

QwicLang supports intuitive `if` / `else`, `while` loops, and `for ... in` collection iteration:

```qwic
import lists

public func main() {
    // While loop
    let count: int = 0
    while count < 3 {
        print(f"Count: {count}")
        count = count + 1
    }

    // For-in iteration
    const fruits = ["apple", "banana", "cherry"]
    for fruit in fruits {
        print(f"Fruit: {fruit}")
    }
}
```

### High-Precision `nano` Numbers

The `nano` type provides exact numerical representation for frame deltas, game logic, and physics:

```qwic
public func main() {
    const delta: nano = 0.016666
    const rate: nano = 2.5
    const step: nano = delta * rate

    print(f"Simulation step: {step}")
}
```

### Data Structures: Literal Declarations (No Imports Needed)

QwicLang supports first-class literal syntax for lists, dictionaries, and tuples without requiring any module imports:

```qwic
public func main() {
    // Lists: [elem, ...]
    const tags = ["systems", "compiler", "native"]
    print(f"First tag: {tags[0]}")
    print(f"Second tag: {tags[1]}")

    // Dictionaries: {key: value, ...}
    const config = {"env": "production", "port": "8080"}
    const env = config["env"]
    print(f"Environment: {env}")

    // Tuples: (item1, item2)
    const coord = ("127.0.0.1", "8080")
    print(f"Host: {coord[0]}, Port: {coord[1]}")
}
```

### Standard Library: HTTP & JSON

```qwic
import http
import json
import time

public func main() {
    const req = http.request("https://api.github.com")
    req.set_header("User-Agent", "QwicLang")

    const resp = http.send(req)
    print(f"HTTP Status: {resp.status}")

    const now = time.now()
    print(f"Timestamp: {now}")
}
```

### Multi-File Modules & Encapsulation

```qwic
// math_utils.qw
module math_utils

private func helper(x: int): int {
    return x * x
}

public func square(x: int): int {
    return helper(x)
}
```

```qwic
// main.qw
import math_utils

public func main() {
    const result = math_utils.square(8)
    print(f"Square: {result}")
}
```

---

## 💻 CLI Reference

The `qwic` binary comes with built-in commands for the complete development workflow:

| Command | Usage | Description |
| :--- | :--- | :--- |
| `run` | `qwic run <file.qw>` | Compiles to a temporary binary and executes it immediately |
| `build` | `qwic build <file.qw> [-o output]` | Compiles source file and dependencies into an executable |
| `check` | `qwic check <file.qw>` | Runs lexical, parsing, semantic, and IR checks without building |
| `fmt` | `qwic fmt <file.qw>` | Formats Qwic source code with consistent indentation |
| `clean` | `qwic clean [file.qw]` | Cleans up compiler build artifacts and temporary files |
| `help` | `qwic --help` | Displays available commands and flags |

---

## 🏗️ Architecture

The Qwic compiler is engineered with strict separation of concerns across phases:

```text
Source Code (.qw)
       │
       ▼
   [ Lexer ]          Transforms raw text into tokens with position tracking
       │
       ▼
  [ Parser ]          Constructs the Abstract Syntax Tree (AST)
       │
       ▼
[ Sema / Types ]      Performs symbol resolution, scope checking, and type validation
       │
       ▼
    [ IR ]            Builds an intermediate representation decoupled from the AST
       │
       ▼
  [ Codegen ]         Emits low-level C / LLVM IR with target optimizations
       │
       ▼
 [ Native Binary ]    Linked with minimal C runtime into a standalone executable
```

---

## 🧪 Running Tests

Ensure all unit and integration tests pass:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run sample programs in examples/
go run ./cmd/qwic run examples/hello.qw
go run ./cmd/qwic run examples/functions.qw
go run ./cmd/qwic run examples/control_flow.qw
go run ./cmd/qwic run examples/data_structures.qw
go run ./cmd/qwic run examples/nano.qw
```

---

## 🗺️ Roadmap (v0 ➔ v1)

- [x] Lexer, parser, AST, and semantic analysis
- [x] Primitive types (`int`, `float`, `nano`, `string`, `bool`, `void`)
- [x] Control flow (`if`/`else`, `while`, `for ... in`)
- [x] Built-in collections (`list`, `dict`, `set`, `tuple`)
- [x] Standard library (`strings`, `time`, `fs`, `sync`, `crypto`, `net`, `json`, `http`)
- [x] Native executable generation via bootstrap backend
- [ ] Direct LLVM IR code generator
- [ ] Concurrency model (`spawn` & `await`)
- [ ] Struct declarations and methods
- [ ] Package manager & registry
- [ ] Self-hosting compiler in QwicLang

---

## 🤝 Contributing

Contributions are very welcome! Whether you are reporting an issue, proposing language syntax, improving the docs, or submitting a pull request:

1. **Fork** the repository.
2. **Create a branch** for your feature: `git checkout -b feature/my-feature`
3. **Write tests** covering your changes.
4. **Ensure all tests pass**: `go test ./...`
5. **Commit your changes**: `git commit -m "sema: add feature XYZ"`
6. **Push** to your fork and submit a **Pull Request**.

Please check [AGENTS.md](AGENTS.md) and [PHASE.md](PHASE.md) for architecture guidelines, code standards, and phase milestones.

---

## 📄 License

QwicLang is distributed under the open-source **[MIT License](LICENCE)**.
