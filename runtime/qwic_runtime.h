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

#endif
