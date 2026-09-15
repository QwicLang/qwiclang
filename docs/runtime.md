# Runtime

Phase 10 adds a small C runtime used by the bootstrap C backend.

## Responsibilities

- Runtime initialization.
- Process exit code handling.
- Printing helpers for supported v0 values.
- Basic allocation and free helpers.
- Tagged dynamic values and heterogeneous bootstrap collections.
- Portable closure callbacks with captured-value contexts.
- JSON parsing and serialization for dynamic values.
- A synchronous cross-platform HTTP/1.1 listener used by Signal.
- Exception-frame helpers for `try`, `catch`, and `throw`.
- Platform-specific sleep, file-existence, and mutex implementations.
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

qwic_try_frame *qwic_try_new(void);
void qwic_try_push(qwic_try_frame *frame);
void qwic_try_end(qwic_try_frame *frame);
const char *qwic_try_message(qwic_try_frame *frame);
void qwic_try_caught(qwic_try_frame *frame);
void qwic_throw(const char *message);

void qwic_print_bool(bool value);
void qwic_print_float(double value);
void qwic_print_int(int64_t value);
void qwic_print_string(const char *value);

int64_t qwic_strings_length(const char *value);
bool qwic_strings_empty(const char *value);
char *qwic_strings_trim(const char *value);

void *qwic_lists_new(void);
void qwic_lists_push(void *list, qwic_value *value);
qwic_value *qwic_lists_get(void *list, int64_t index);
int64_t qwic_lists_length(void *list);

void *qwic_sets_new(void);
void qwic_sets_add(void *set, qwic_value *value);
bool qwic_sets_contains(void *set, qwic_value *value);
int64_t qwic_sets_length(void *set);

void *qwic_dictionaries_new(void);
void qwic_dictionaries_set(void *dictionary, const char *key, qwic_value *value);
qwic_value *qwic_dictionaries_get(void *dictionary, const char *key);

void *qwic_tuples_new2(qwic_value *first, qwic_value *second);
qwic_value *qwic_tuples_first(void *tuple);
qwic_value *qwic_tuples_second(void *tuple);

qwic_closure *qwic_closure_new(qwic_callback callback, void *context);
qwic_value *qwic_closure_call(qwic_closure *closure, qwic_value **args, size_t count);

qwic_value *qwic_json_parse(const char *json);
char *qwic_json_stringify(qwic_value *value);
void qwic_http_listen(int64_t port, qwic_closure *handler);
```

## Current Limits

- The runtime is intentionally tiny.
- String formatting and bootstrap data structures allocate through the runtime;
  ownership and reclamation are still intentionally simple in v0.
- `nano` printing uses the same floating-point helper as `float` until the
  fixed-point representation is implemented.
- Tagged values and closure contexts are runtime allocated without tracing or
  automatic reclamation in v0.
- The HTTP listener is synchronous HTTP/1.1. TLS, streaming, keep-alive,
  chunked request bodies, and concurrency are not implemented.
- Exception frames are process-global until Qwic has a concurrency and
  thread-local runtime model.
- Windows synchronization uses `CRITICAL_SECTION`; Linux and macOS use
  `pthread_mutex_t`. Generated programs do not require pthread symbols on
  Windows.
