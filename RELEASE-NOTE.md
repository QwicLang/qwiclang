# QwicLang v0.1.0-alpha Release Notes

We are thrilled to announce the first pre-release of **QwicLang (`v0.1.0-alpha`)**!

QwicLang is a lightweight, readable, statically typed, and natively compiled programming language designed for fast compilation, developer ergonomic simplicity, and predictable native runtime performance.

---

## 🚀 Highlights & Features in this Release

### 1. End-to-End Native Compilation Pipeline
- High-speed compiler transforming `.qw` source into native executables via an internal IR and C code generator.
- Single command execution:
  - `qwic run <file.qw>`: One-step compile and execute.
  - `qwic build <file.qw> -o <binary>`: Compile to standalone machine executable.

### 2. Modern, Clean Syntax & Static Type System
- **Variables & Mutability**: Explicit `const` (immutable) and `let` (mutable) declarations with static type validation and inference.
- **Primitive Types**: `int`, `float`, `string`, `bool`, and `void`.
- **Specialized `nano` Type**: High-precision numeric representation tailored for animation, physics, simulation, and game delta calculations.
- **Functions**: Typed parameters, explicit return types, and visibility controls (`public` and `private`).
- **Control Flow**: Conditionals (`if` / `else`), `while` loops, and `for ... in` collection iteration.
- **String Interpolation**: Readable f-strings (`f"Count: {count}"`).
- **Flexible Syntax**: Optional semicolons and indentation-independent brace scoping.

### 3. First-Class Literal Data Structures
Native literal syntax for collections without requiring any explicit module imports:
- **Lists**: `["item1", "item2"]` with indexing `list[0]` and slicing `list[start:end]`.
- **Dictionaries**: `{"key": "value"}` with key-based indexing `dict["key"]`.
- **Tuples**: `("alpha", "beta")` with indexed element access `pair[0]`, `pair[1]`.
- **Direct Collection Printing**: `print(data)` outputs formatted collections natively:
  - Lists: `["item1", "item2"]`
  - Dictionaries: `{"key": "value"}`
  - Tuples: `("item1", "item2")`

### 4. `turbo` Performance Marker
- The `turbo` keyword marks latency-critical functions (`turbo func calculate(...)`).
- Automatically emits compiler optimization directives (`inline`, `hot`, `optimize("O3")`) in the native backend.

### 5. Builtin Standard Library
- **`strings`**: `trim`, `upper`, `lower`, `contains`, `starts_with`, `ends_with`, `index_of`, `slice`.
- **`math`**: `abs`, `min`, `max`, `sqrt`, `pow`, `floor`, `ceil`, `round`.
- **`time`**: Millisecond/microsecond/nanosecond timestamps and `sleep`.
- **`http` & `json`**: Native HTTP client requests, headers, and basic JSON value querying.
- **`lists`, `sets`, `dictionaries`, `tuples`**: Comprehensive standard functions for collections.

### 6. Developer Experience & CLI Tooling
- `qwic build`: Build native binaries.
- `qwic run`: Direct execution of `.qw` scripts.
- `qwic check`: Semantic analysis and type checking without compiling.
- `qwic fmt`: Source code formatter.
- `qwic clean`: Build cache and temporary artifact cleanup.

---

## 📦 Quick Start

### Installation
Clone and build the compiler using Go (>= 1.22):

```bash
git clone https://github.com/QwicLang/qwiclang.git
cd qwiclang
go build -o qwic ./cmd/qwic
```

### Running Your First Program

Create `hello.qw`:

```qwic
public func main() {
    print("Hello, Qwic!")
}
```

Run:

```bash
./qwic run hello.qw
```

---

## 🔮 What's Next

Future milestones for upcoming releases:
- LLVM direct code generation backend
- Structured types (`struct`)
- Multi-file module import system
- Concurrency primitives (`spawn` and `await`)
