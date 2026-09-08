# Qwic Standard Library

The Qwic standard library is intentionally small during v0. Packages are added
only when the compiler and runtime can support their public APIs end-to-end.

Implemented packages:

- `strings`

Planned packages from `PACKAGE-LIBRARY-PLAN.md`:

- `time`
- `fs`
- `sync`
- `net`
- `json`
- `crypto`
- `http`

Those planned packages are not exposed yet because they require language
features that are still deferred, including structs, arrays, bytes, async/await,
methods, and a settled result/error model.
