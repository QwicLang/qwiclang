# Runtime

Phase 10 adds a small C runtime used by the bootstrap C backend.

## Responsibilities

- Runtime initialization.
- Process exit code handling.
- Printing helpers for supported v0 values.
- Basic allocation and free helpers.
- A small platform-name helper for future platform-specific runtime work.

## Files

- `runtime/qwic_runtime.h`
- `runtime/qwic_runtime.c`

Generated C includes `qwic_runtime.h` and links with `qwic_runtime.c`.

## Current API

```c
void qwic_runtime_init(void);
int qwic_runtime_exit_code(void);
const char *qwic_platform_name(void);

void *qwic_alloc(size_t size);
void qwic_free(void *ptr);

void qwic_print_bool(bool value);
void qwic_print_float(double value);
void qwic_print_int(int64_t value);
void qwic_print_string(const char *value);
```

## Current Limits

- The runtime is intentionally tiny.
- Allocation helpers are present for a future memory model, but generated v0
  programs do not rely on dynamic allocation yet.
- `nano` printing uses the same floating-point helper as `float` until the
  fixed-point representation is implemented.
