#include "qwic_runtime.h"

#include <stdlib.h>
#include <stdio.h>

static int qwic_exit_code = 0;

void qwic_runtime_init(void) {
    qwic_exit_code = 0;
}

int qwic_runtime_exit_code(void) {
    return qwic_exit_code;
}

const char *qwic_platform_name(void) {
#if defined(_WIN32)
    return "windows";
#elif defined(__APPLE__)
    return "darwin";
#elif defined(__linux__)
    return "linux";
#else
    return "unknown";
#endif
}

void *qwic_alloc(size_t size) {
    return malloc(size);
}

void qwic_free(void *ptr) {
    free(ptr);
}

void qwic_print_bool(bool value) {
    printf("%s\n", value ? "true" : "false");
}

void qwic_print_float(double value) {
    printf("%.6f\n", value);
}

void qwic_print_int(int64_t value) {
    printf("%lld\n", (long long)value);
}

void qwic_print_string(const char *value) {
    printf("%s\n", value);
}
