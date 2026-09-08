# CLI

Phase 11 provides the first practical compiler commands.

## Commands

```bash
qwic build <source.qw> [-o output]
qwic run <source.qw>
qwic check <source.qw>
qwic fmt <source.qw>
qwic clean [source.qw]
qwic --help
```

## Behavior

- `build` compiles a source file and same-directory imports into a native
  executable.
- `run` builds into a temporary executable, executes it, prints program output,
  and returns the program exit status.
- `check` runs parsing, semantic analysis, and IR generation without producing
  an executable.
- `fmt` rewrites one source file with conservative token-based formatting.
- `clean` removes the default executable for a source file, or `.qwic-cache`
  when no source is provided.

## Exit Codes

- `0`: success.
- `1`: compiler, file, runtime, or external tool failure.
- `2`: command-line usage error.

## Current Limits

- Import resolution is same-directory only.
- The formatter is intentionally conservative and does not preserve comments.
- `clean` only removes known v0 build artifacts.
