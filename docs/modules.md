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

`qwic build main.qw` can resolve `import users` from an installed shared
package, a project package directory, or the legacy `users.qw` file beside the
importing source. See [Packages](packages.md) for the complete resolution order.

Standard-library packages such as `strings` are resolved by the compiler and do
not require a local `strings.qw` file.

## Visibility

- Symbols are private by default.
- `public` functions can be called from another module.
- `private` functions cannot be called from another module.
- Visibility errors are reported at compile time.

## Current Limits

- Module aliases are not implemented.
- Public type and function symbols participate in module visibility.
- Nested module paths are deferred.
- Package versions and lockfiles are deferred.
