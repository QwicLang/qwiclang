# Runtime

Phase 10 adds a small C runtime used by the bootstrap C backend.

## Responsibilities

- Runtime initialization.
- Process exit code handling.
- Printing helpers for supported v0 values.
- Basic allocation and free helpers.
- Runtime-backed string collections for the bootstrap data-structure packages.
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

int64_t qwic_strings_length(const char *value);
bool qwic_strings_empty(const char *value);
char *qwic_strings_trim(const char *value);

void *qwic_lists_new(void);
void qwic_lists_push(void *list, const char *value);
const char *qwic_lists_get(void *list, int64_t index);
int64_t qwic_lists_length(void *list);

void *qwic_sets_new(void);
void qwic_sets_add(void *set, const char *value);
bool qwic_sets_contains(void *set, const char *value);
int64_t qwic_sets_length(void *set);

void *qwic_dictionaries_new(void);
void qwic_dictionaries_set(void *dictionary, const char *key, const char *value);
const char *qwic_dictionaries_get(void *dictionary, const char *key);

void *qwic_tuples_new2(const char *first, const char *second);
const char *qwic_tuples_first(void *tuple);
const char *qwic_tuples_second(void *tuple);
```

## Current Limits

- The runtime is intentionally tiny.
- String formatting and bootstrap data structures allocate through the runtime;
  ownership and reclamation are still intentionally simple in v0.
- `nano` printing uses the same floating-point helper as `float` until the
  fixed-point representation is implemented.
- Lists, sets, dictionaries, and tuples store strings only until the language
  has generics or a designed `any` container model.
