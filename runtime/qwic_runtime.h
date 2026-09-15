#ifndef QWIC_RUNTIME_H
#define QWIC_RUNTIME_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <setjmp.h>

typedef struct qwic_value qwic_value;
typedef struct qwic_closure qwic_closure;
typedef qwic_value *(*qwic_callback)(void *context, qwic_value **args, size_t count);

struct qwic_closure {
    qwic_callback callback;
    void *context;
};

typedef struct qwic_try_frame {
    jmp_buf environment;
    struct qwic_try_frame *previous;
    const char *message;
} qwic_try_frame;

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
void qwic_print_list(void *value);
void qwic_print_set(void *value);
void qwic_print_dictionary(void *value);
void qwic_print_tuple(void *value);
void qwic_print_any(qwic_value *value);

qwic_value *qwic_any_null(void);
qwic_value *qwic_any_bool(bool value);
qwic_value *qwic_any_int(int64_t value);
qwic_value *qwic_any_float(double value);
qwic_value *qwic_any_string(const char *value);
qwic_value *qwic_any_list(void *value);
qwic_value *qwic_any_set(void *value);
qwic_value *qwic_any_dictionary(void *value);
qwic_value *qwic_any_tuple(void *value);
qwic_value *qwic_any_function(qwic_closure *value);
qwic_value *qwic_any_pointer(void *value);
bool qwic_any_as_bool(qwic_value *value);
int64_t qwic_any_as_int(qwic_value *value);
double qwic_any_as_float(qwic_value *value);
const char *qwic_any_as_string(qwic_value *value);
void *qwic_any_as_list(qwic_value *value);
void *qwic_any_as_set(qwic_value *value);
void *qwic_any_as_dictionary(qwic_value *value);
void *qwic_any_as_tuple(qwic_value *value);
qwic_closure *qwic_any_as_function(qwic_value *value);
void *qwic_any_as_pointer(qwic_value *value);
const char *qwic_any_type(qwic_value *value);
char *qwic_any_to_string(qwic_value *value);
bool qwic_any_equal(qwic_value *left, qwic_value *right);
int qwic_any_compare(qwic_value *left, qwic_value *right);
qwic_value *qwic_any_add(qwic_value *left, qwic_value *right);
qwic_value *qwic_any_field(qwic_value *value, const char *field);
void qwic_any_set_field(qwic_value *value, const char *field, qwic_value *item);
qwic_value *qwic_any_index(qwic_value *value, qwic_value *index);
void qwic_any_set_index(qwic_value *value, qwic_value *index, qwic_value *item);

qwic_closure *qwic_closure_new(qwic_callback callback, void *context);
qwic_value *qwic_closure_call(qwic_closure *closure, qwic_value **args, size_t count);

int64_t qwic_strings_length(const char *value);
bool qwic_strings_empty(const char *value);
char *qwic_strings_trim(const char *value);
char *qwic_strings_trim_left(const char *value);
char *qwic_strings_trim_right(const char *value);
char *qwic_strings_upper(const char *value);
char *qwic_strings_lower(const char *value);
bool qwic_strings_contains(const char *value, const char *needle);
bool qwic_strings_starts_with(const char *value, const char *prefix);
bool qwic_strings_ends_with(const char *value, const char *suffix);
int64_t qwic_strings_index_of(const char *value, const char *needle);
char *qwic_strings_concat(const char *left, const char *right);
void *qwic_strings_split(const char *value, const char *separator);
char *qwic_strings_substring(const char *value, int64_t start, int64_t end);
int64_t qwic_strings_to_int(const char *value);

// Math
double qwic_math_abs(double x);
double qwic_math_min(double a, double b);
double qwic_math_max(double a, double b);
double qwic_math_sqrt(double x);
double qwic_math_pow(double base, double exp);
double qwic_math_floor(double x);
double qwic_math_ceil(double x);
double qwic_math_round(double x);

int64_t qwic_time_now(void);
void qwic_time_sleep(int64_t ms);
int64_t qwic_time_duration(int64_t start, int64_t end);

const char *qwic_fs_read_file(const char *path);
bool qwic_fs_write_file(const char *path, const char *content);
bool qwic_fs_exists(const char *path);

void *qwic_sync_mutex_new(void);
void qwic_sync_mutex_lock(void *mutex);
void qwic_sync_mutex_unlock(void *mutex);
void qwic_sync_mutex_free(void *mutex);

// Crypto
char *qwic_crypto_sha256(const char *value);
char *qwic_crypto_sha512(const char *value);
char *qwic_crypto_random_bytes(int len);
int64_t qwic_crypto_random_int(void);
char *qwic_crypto_hex_encode(const char *bytes, int len);
char *qwic_crypto_base64_encode(const char *bytes, int len);

// Net
void *qwic_net_tcp_connect(const char *host, int64_t port);
void *qwic_net_tcp_listen(const char *port_str);
void *qwic_net_tcp_accept(void *listen_conn);
const char *qwic_net_tcp_read(void *conn, int64_t max_len);
int64_t qwic_net_tcp_write(void *conn, const char *data);
void qwic_net_tcp_close(void *conn);
const char *qwic_net_dns_lookup(const char *host);

// JSON
char *qwic_json_stringify(qwic_value *value);
qwic_value *qwic_json_parse(const char *json);
char *qwic_json_encode(qwic_value *value);
qwic_value *qwic_json_decode(const char *json);

// HTTP
void *qwic_http_request_new(const char *url);
void qwic_http_request_set_method(void *req, const char *method);
void qwic_http_request_set_body(void *req, const char *body);
void *qwic_http_send(void *req);
void qwic_http_request_free(void *req);

int64_t qwic_http_get_status(void *response);
const char *qwic_http_get_body(void *response);
void qwic_http_free_response(void *response);
void qwic_http_listen(int64_t port, qwic_closure *handler);

void *qwic_lists_new(void);
void qwic_lists_push(void *list, qwic_value *value);
qwic_value *qwic_lists_get(void *list, int64_t index);
int64_t qwic_lists_length(void *list);
bool qwic_lists_contains(void *list, qwic_value *value);
void *qwic_lists_slice(void *list, int64_t start, int64_t end);
void *qwic_lists_concat(void *left, void *right);

void *qwic_sets_new(void);
void qwic_sets_add(void *set, qwic_value *value);
bool qwic_sets_contains(void *set, qwic_value *value);
int64_t qwic_sets_length(void *set);

void *qwic_dictionaries_new(void);
void qwic_dictionaries_set(void *dictionary, const char *key, qwic_value *value);
qwic_value *qwic_dictionaries_get(void *dictionary, const char *key);
bool qwic_dictionaries_contains(void *dictionary, const char *key);
int64_t qwic_dictionaries_length(void *dictionary);
void *qwic_dictionaries_keys(void *dictionary);

void *qwic_tuples_new2(qwic_value *first, qwic_value *second);
qwic_value *qwic_tuples_first(void *tuple);
qwic_value *qwic_tuples_second(void *tuple);
int64_t qwic_tuples_length(void *tuple);

#endif
