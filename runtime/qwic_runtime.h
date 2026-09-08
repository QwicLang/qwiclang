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

#endif
