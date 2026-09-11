#ifndef QWIC_RUNTIME_H
#define QWIC_RUNTIME_H

#include <stdbool.h>
#include <stdint.h>
#include <stddef.h>

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
char *qwic_strings_trim_left(const char *value);
char *qwic_strings_trim_right(const char *value);
char *qwic_strings_upper(const char *value);
char *qwic_strings_lower(const char *value);
bool qwic_strings_contains(const char *value, const char *needle);
bool qwic_strings_starts_with(const char *value, const char *prefix);
bool qwic_strings_ends_with(const char *value, const char *suffix);
int64_t qwic_strings_index_of(const char *value, const char *needle);

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

void *qwic_lists_new(void);
void qwic_lists_push(void *list, const char *value);
const char *qwic_lists_get(void *list, int64_t index);
int64_t qwic_lists_length(void *list);
bool qwic_lists_contains(void *list, const char *value);

void *qwic_sets_new(void);
void qwic_sets_add(void *set, const char *value);
bool qwic_sets_contains(void *set, const char *value);
int64_t qwic_sets_length(void *set);

void *qwic_dictionaries_new(void);
void qwic_dictionaries_set(void *dictionary, const char *key, const char *value);
const char *qwic_dictionaries_get(void *dictionary, const char *key);
bool qwic_dictionaries_contains(void *dictionary, const char *key);
int64_t qwic_dictionaries_length(void *dictionary);

void *qwic_tuples_new2(const char *first, const char *second);
const char *qwic_tuples_first(void *tuple);
const char *qwic_tuples_second(void *tuple);
int64_t qwic_tuples_length(void *tuple);

#endif
