# Qwic Standard Library

The Qwic standard library is intentionally small during v0. Packages are added
only when the compiler and runtime can support their public APIs end-to-end.

Implemented packages:

- `strings`
- `lists`
- `sets`
- `dictionaries`
- `tuples`

Planned packages from `PACKAGE-LIBRARY-PLAN.md`:

- `time`
- `fs`
- `sync`
- `net`
- `json`
- `crypto`
- `http`

The bootstrap collection packages are string-oriented because Qwic does not yet
have generics, `any` containers, array literals, structs, methods, or a settled
ownership model.

The remaining planned packages are not exposed yet because they require language
features that are still deferred, including bytes, async/await, resource
handles, callbacks, and a settled result/error model.
