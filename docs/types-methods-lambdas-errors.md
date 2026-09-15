# Types, Methods, Lambdas, and Errors

Qwic supports record types, type-owned functions, typed anonymous functions,
and recoverable string errors. These features are compiled through the normal
AST, semantic analysis, IR, and native C backend pipeline.

## Record Types

Declare a type with named, statically checked fields:

```qwic
public type User {
    name: string
    age: int
}
```

Construct a value by providing every field exactly once:

```qwic
const user = User {
    name: "Ada"
    age: 36
}
```

Type names used in literals must begin with an uppercase letter. Missing,
duplicate, unknown, and incorrectly typed fields are compile-time errors.
Record values are currently heap allocated; automatic record deallocation is
not implemented yet.

## Methods

A method declares its receiver before the method name:

```qwic
func (user: User) greeting(): string {
    return f"Hello, {user.name}"
}
```

The receiver name is explicit and local to the method. Qwic does not define an
implicit `this` keyword. A type-qualified function has no receiver and is
static; `new` is the constructor convention:

```qwic
func User.new(name: string, age: int): User {
    return User { name: name, age: age }
}
```

Method visibility follows the same `public` and `private` module rules as
ordinary functions. Method references are not first-class yet; invoke methods
directly on a receiver.

## Lambdas

Lambdas use `=>` after their parameter list and optional return type:

```qwic
const double = (value: int): int => {
    return value * 2
}

print(double(21))
```

The compiler lowers each lambda to a private native callback and a portable
closure containing the callback and its captured-value context. Parameters
default to `any`; an omitted return type is `any` so callbacks can return
dynamic values.

Lambdas may capture surrounding local variables. Captures currently use
by-value snapshot semantics when the closure is created. Closure environments
are runtime allocated and are not automatically reclaimed in the bootstrap
runtime.

## Try, Catch, Throw

Throw a string and handle it with `try` / `catch`:

```qwic
func validate(ok: bool) {
    if !ok {
        throw "validation failed"
    }
}

func main() {
    try {
        validate(false)
    } catch (error) {
        print(error)
    }
}
```

Catch variables are immutable strings scoped to the catch block. An uncaught
error is written to standard error and exits the process with status 1.

The bootstrap runtime implements stack-like exception frames with standard C
`setjmp` and `longjmp`. The exception stack is process-global because Qwic does
not yet have a concurrency runtime. Typed error values, `finally`, stack traces,
and cross-thread propagation are not implemented.

## Platform Notes

Generated code uses C11-compatible declarations, a common callback ABI, and
standard C exception primitives. The CI build exercises GCC on Linux, Apple
Clang on macOS, and MinGW-w64 GCC on Windows.
