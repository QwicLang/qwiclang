#include "qwic_runtime.h"

#include <stdlib.h>
#include <stdio.h>
#include <ctype.h>
#include <string.h>

static int qwic_exit_code = 0;

static const char *qwic_trim_left_start(const char *value);
static const char *qwic_trim_right_end(const char *value);
static char *qwic_copy_range(const char *start, const char *end);
static char *qwic_ascii_map(const char *value, bool uppercase);

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

int64_t qwic_strings_length(const char *value) {
    if (value == NULL) {
        return 0;
    }
    return (int64_t)strlen(value);
}

bool qwic_strings_empty(const char *value) {
    return value == NULL || value[0] == '\0';
}

char *qwic_strings_trim(const char *value) {
    if (value == NULL) {
        return qwic_copy_range("", "");
    }
    const char *start = qwic_trim_left_start(value);
    const char *end = qwic_trim_right_end(start);
    return qwic_copy_range(start, end);
}

char *qwic_strings_trim_left(const char *value) {
    if (value == NULL) {
        return qwic_copy_range("", "");
    }
    const char *start = qwic_trim_left_start(value);
    return qwic_copy_range(start, value + strlen(value));
}

char *qwic_strings_trim_right(const char *value) {
    if (value == NULL) {
        return qwic_copy_range("", "");
    }
    const char *end = qwic_trim_right_end(value);
    return qwic_copy_range(value, end);
}

char *qwic_strings_upper(const char *value) {
    return qwic_ascii_map(value, true);
}

char *qwic_strings_lower(const char *value) {
    return qwic_ascii_map(value, false);
}

bool qwic_strings_contains(const char *value, const char *needle) {
    if (value == NULL || needle == NULL) {
        return false;
    }
    return strstr(value, needle) != NULL;
}

bool qwic_strings_starts_with(const char *value, const char *prefix) {
    if (value == NULL || prefix == NULL) {
        return false;
    }
    size_t prefix_length = strlen(prefix);
    return strncmp(value, prefix, prefix_length) == 0;
}

bool qwic_strings_ends_with(const char *value, const char *suffix) {
    if (value == NULL || suffix == NULL) {
        return false;
    }
    size_t value_length = strlen(value);
    size_t suffix_length = strlen(suffix);
    if (suffix_length > value_length) {
        return false;
    }
    return strcmp(value + value_length - suffix_length, suffix) == 0;
}

int64_t qwic_strings_index_of(const char *value, const char *needle) {
    if (value == NULL || needle == NULL) {
        return -1;
    }
    const char *match = strstr(value, needle);
    if (match == NULL) {
        return -1;
    }
    return (int64_t)(match - value);
}

static const char *qwic_trim_left_start(const char *value) {
    while (*value != '\0' && isspace((unsigned char)*value)) {
        value++;
    }
    return value;
}

static const char *qwic_trim_right_end(const char *value) {
    const char *end = value + strlen(value);
    while (end > value && isspace((unsigned char)*(end - 1))) {
        end--;
    }
    return end;
}

static char *qwic_copy_range(const char *start, const char *end) {
    size_t length = (size_t)(end - start);
    char *copy = qwic_alloc(length + 1);
    memcpy(copy, start, length);
    copy[length] = '\0';
    return copy;
}

static char *qwic_ascii_map(const char *value, bool uppercase) {
    if (value == NULL) {
        return qwic_copy_range("", "");
    }
    size_t length = strlen(value);
    char *copy = qwic_alloc(length + 1);
    for (size_t index = 0; index < length; index++) {
        unsigned char ch = (unsigned char)value[index];
        copy[index] = uppercase ? (char)toupper(ch) : (char)tolower(ch);
    }
    copy[length] = '\0';
    return copy;
}
