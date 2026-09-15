# Packages

Qwic uses a shared package directory so dependencies are downloaded once and
can be reused by every project owned by the same user.

## Installing a Package

The official package catalog is hosted at:

```text
https://github.com/QwicLang/packages
```

Install a package by its catalog directory name:

```bash
qwic install signal
```

The command copies the package into:

```text
~/qwic/packages/signal
```

Running the command again reuses the existing installation and does not
download another copy.

## Package Layout

Each catalog and project package uses this layout:

```text
signal/
├── README.md
└── src/
    ├── app.qw
    ├── request.qw
    └── response.qw
```

All `.qw` files beneath `src` belong to the package named by the directory. A
source-level `module` declaration is optional for these files. When present, it
must match the package directory name.

Only `public` functions are accessible to importing projects. Private functions
can be shared by files within the package.

## Import Resolution

Given:

```qwic
import signal
```

the compiler searches in this order:

1. Compiler-provided standard packages.
2. `~/qwic/packages/signal/src`.
3. `<project>/packages/signal/src`.
4. `<project>/signal/src`.
5. The legacy same-directory `signal.qw` module.

The first matching package wins. This means an installed shared package takes
precedence over a project package with the same name.

The project root is the nearest parent containing `qwic.toml`. When no manifest
exists, the compiler uses the current working directory if the entry source is
inside it; otherwise it uses the entry source directory.

## Configuration

The defaults can be overridden for testing or private mirrors:

```text
QWIC_PACKAGES_DIR
QWIC_PACKAGES_REPOSITORY
```

`QWIC_PACKAGES_DIR` replaces `~/qwic/packages`.

The default uses the operating system's user-home directory and native path
separator:

| Platform | Default location |
| --- | --- |
| Linux | `/home/<user>/qwic/packages` |
| macOS | `/Users/<user>/qwic/packages` |
| Windows | `C:\Users\<user>\qwic\packages` |

`qwic install` creates the full directory tree when it does not exist. The
implementation uses Go's `os.UserHomeDir`, `filepath.Join`, and `os.MkdirAll`
instead of assuming Unix environment variables or separators.
`QWIC_PACKAGES_REPOSITORY` replaces the official Git repository URL.

## Current Limits

- Installation requires `git` when using a remote package repository.
- Packages are installed from the current default branch snapshot.
- Package versions, lockfiles, updates, removal, and checksums are not yet
  implemented.
- Installation validates package layout but does not compile the package.
- Package source must use language features supported by the current compiler.
- The current package identity is its import name; globally unique package IDs
  are deferred until versioned manifests are introduced.
- Symbolic links inside catalog packages are rejected.

The first version intentionally compiles package source together with the
application. Precompiled package artifacts and a stable binary ABI are deferred.
