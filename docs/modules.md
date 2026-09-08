# Modules and Visibility

Phase 9 introduces basic same-directory modules and import visibility checks.

## Module Files

```qwic
module users

private func validateUser() {
}

public func createUser() {
    print("created")
}
```

## Imports

```qwic
import users

public func main() {
    users.createUser()
}
```

`qwic build main.qw` resolves `import users` as `users.qw` in the same
directory as `main.qw`. Imported source files may define a `module` declaration.

## Visibility

- Symbols are private by default.
- `public` functions can be called from another module.
- `private` functions cannot be called from another module.
- Visibility errors are reported at compile time.

## Current Limits

- Imports resolve only to same-directory `.qw` files.
- Module aliases are not implemented.
- Only function symbols participate in module visibility.
- Package registries and nested module paths are deferred.
