#if !defined(_WIN32)
#define _POSIX_C_SOURCE 200809L
#endif

#include "qwic_runtime.h"

#include <stdlib.h>
#include <stdio.h>
#include <ctype.h>
#include <string.h>
#include <time.h>
#include <math.h>

#if defined(_WIN32)
#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#include <io.h>
#else
#include <arpa/inet.h>
#include <netinet/in.h>
#include <pthread.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>
#endif

#if defined(_WIN32)
typedef SOCKET qwic_socket;
#define QWIC_INVALID_SOCKET INVALID_SOCKET
#define qwic_close_socket closesocket
#else
typedef int qwic_socket;
#define QWIC_INVALID_SOCKET (-1)
#define qwic_close_socket close
#endif

static int qwic_exit_code = 0;
static qwic_try_frame *qwic_active_try_frame = NULL;

typedef enum {
    QWIC_VALUE_NULL,
    QWIC_VALUE_BOOL,
    QWIC_VALUE_INT,
    QWIC_VALUE_FLOAT,
    QWIC_VALUE_STRING,
    QWIC_VALUE_LIST,
    QWIC_VALUE_SET,
    QWIC_VALUE_DICTIONARY,
    QWIC_VALUE_TUPLE,
    QWIC_VALUE_FUNCTION,
    QWIC_VALUE_POINTER
} qwic_value_kind;

struct qwic_value {
    qwic_value_kind kind;
    union {
        bool boolean;
        int64_t integer;
        double floating;
        char *string;
        void *pointer;
        qwic_closure *function;
    } as;
};

typedef struct {
    size_t length;
    size_t capacity;
    qwic_value **items;
} qwic_list;

typedef struct {
    qwic_list values;
} qwic_set;

typedef struct {
    char *key;
    qwic_value *value;
} qwic_dictionary_entry;

typedef struct {
    size_t length;
    size_t capacity;
    qwic_dictionary_entry *entries;
} qwic_dictionary;

typedef struct {
    size_t length;
    qwic_value **items;
} qwic_tuple;

typedef struct {
    const char *cursor;
} qwic_json_reader;

typedef struct {
    char *data;
    size_t length;
    size_t capacity;
} qwic_string_builder;

static const char *qwic_trim_left_start(const char *value);
static const char *qwic_trim_right_end(const char *value);
static char *qwic_copy_range(const char *start, const char *end);
static char *qwic_ascii_map(const char *value, bool uppercase);
static char *qwic_copy_string(const char *value);
static void qwic_list_init(qwic_list *list);
static void qwic_list_push_value(qwic_list *list, qwic_value *value);
static bool qwic_list_contains(qwic_list *list, qwic_value *value);
static void qwic_list_reserve(qwic_list *list, size_t needed);
static int64_t qwic_dictionary_find(qwic_dictionary *dictionary, const char *key);
static void qwic_dictionary_reserve(qwic_dictionary *dictionary, size_t needed);
static qwic_value *qwic_new_value(qwic_value_kind kind);
static qwic_value *qwic_json_read_value(qwic_json_reader *reader);
static void qwic_json_write_value(qwic_string_builder *builder, qwic_value *value);
static void qwic_builder_append(qwic_string_builder *builder, const char *value);
static void qwic_builder_append_char(qwic_string_builder *builder, char value);

void qwic_runtime_init(void) {
    qwic_exit_code = 0;
    qwic_active_try_frame = NULL;
#if defined(_WIN32)
    WSADATA socket_data;
    if (WSAStartup(MAKEWORD(2, 2), &socket_data) != 0) {
        fprintf(stderr, "fatal: unable to initialize Winsock\n");
        exit(1);
    }
#endif
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

static qwic_value *qwic_new_value(qwic_value_kind kind) {
    qwic_value *value = qwic_alloc(sizeof(qwic_value));
    if (value == NULL) {
        fprintf(stderr, "fatal: unable to allocate runtime value\n");
        exit(1);
    }
    memset(value, 0, sizeof(qwic_value));
    value->kind = kind;
    return value;
}

qwic_value *qwic_any_null(void) {
    static qwic_value null_value = {QWIC_VALUE_NULL, {0}};
    return &null_value;
}

qwic_value *qwic_any_bool(bool raw) {
    qwic_value *value = qwic_new_value(QWIC_VALUE_BOOL);
    value->as.boolean = raw;
    return value;
}

qwic_value *qwic_any_int(int64_t raw) {
    qwic_value *value = qwic_new_value(QWIC_VALUE_INT);
    value->as.integer = raw;
    return value;
}

qwic_value *qwic_any_float(double raw) {
    qwic_value *value = qwic_new_value(QWIC_VALUE_FLOAT);
    value->as.floating = raw;
    return value;
}

qwic_value *qwic_any_string(const char *raw) {
    qwic_value *value = qwic_new_value(QWIC_VALUE_STRING);
    value->as.string = qwic_copy_string(raw);
    return value;
}

static qwic_value *qwic_any_container(qwic_value_kind kind, void *raw) {
    qwic_value *value = qwic_new_value(kind);
    value->as.pointer = raw;
    return value;
}

qwic_value *qwic_any_list(void *raw) { return qwic_any_container(QWIC_VALUE_LIST, raw); }
qwic_value *qwic_any_set(void *raw) { return qwic_any_container(QWIC_VALUE_SET, raw); }
qwic_value *qwic_any_dictionary(void *raw) { return qwic_any_container(QWIC_VALUE_DICTIONARY, raw); }
qwic_value *qwic_any_tuple(void *raw) { return qwic_any_container(QWIC_VALUE_TUPLE, raw); }
qwic_value *qwic_any_pointer(void *raw) { return qwic_any_container(QWIC_VALUE_POINTER, raw); }

qwic_value *qwic_any_function(qwic_closure *raw) {
    qwic_value *value = qwic_new_value(QWIC_VALUE_FUNCTION);
    value->as.function = raw;
    return value;
}

bool qwic_any_as_bool(qwic_value *value) {
    if (value == NULL || value->kind == QWIC_VALUE_NULL) return false;
    if (value->kind == QWIC_VALUE_BOOL) return value->as.boolean;
    if (value->kind == QWIC_VALUE_INT) return value->as.integer != 0;
    if (value->kind == QWIC_VALUE_FLOAT) return value->as.floating != 0.0;
    if (value->kind == QWIC_VALUE_STRING) return value->as.string[0] != '\0';
    return true;
}

int64_t qwic_any_as_int(qwic_value *value) {
    if (value == NULL) return 0;
    if (value->kind == QWIC_VALUE_INT) return value->as.integer;
    if (value->kind == QWIC_VALUE_FLOAT) return (int64_t)value->as.floating;
    if (value->kind == QWIC_VALUE_BOOL) return value->as.boolean ? 1 : 0;
    if (value->kind == QWIC_VALUE_STRING) return strtoll(value->as.string, NULL, 10);
    return 0;
}

double qwic_any_as_float(qwic_value *value) {
    if (value == NULL) return 0.0;
    if (value->kind == QWIC_VALUE_FLOAT) return value->as.floating;
    if (value->kind == QWIC_VALUE_INT) return (double)value->as.integer;
    if (value->kind == QWIC_VALUE_STRING) return strtod(value->as.string, NULL);
    return 0.0;
}

const char *qwic_any_as_string(qwic_value *value) {
    if (value != NULL && value->kind == QWIC_VALUE_STRING) return value->as.string;
    return qwic_any_to_string(value);
}

static void *qwic_any_as_container(qwic_value *value, qwic_value_kind expected) {
    if (value == NULL || value->kind != expected) return NULL;
    return value->as.pointer;
}

void *qwic_any_as_list(qwic_value *value) { return qwic_any_as_container(value, QWIC_VALUE_LIST); }
void *qwic_any_as_set(qwic_value *value) { return qwic_any_as_container(value, QWIC_VALUE_SET); }
void *qwic_any_as_dictionary(qwic_value *value) { return qwic_any_as_container(value, QWIC_VALUE_DICTIONARY); }
void *qwic_any_as_tuple(qwic_value *value) { return qwic_any_as_container(value, QWIC_VALUE_TUPLE); }

qwic_closure *qwic_any_as_function(qwic_value *value) {
    if (value == NULL || value->kind != QWIC_VALUE_FUNCTION) return NULL;
    return value->as.function;
}

void *qwic_any_as_pointer(qwic_value *value) {
    if (value == NULL || value->kind != QWIC_VALUE_POINTER) return NULL;
    return value->as.pointer;
}

const char *qwic_any_type(qwic_value *value) {
    if (value == NULL) return "null";
    switch (value->kind) {
    case QWIC_VALUE_NULL: return "null";
    case QWIC_VALUE_BOOL: return "bool";
    case QWIC_VALUE_INT: return "int";
    case QWIC_VALUE_FLOAT: return "float";
    case QWIC_VALUE_STRING: return "string";
    case QWIC_VALUE_LIST: return "list";
    case QWIC_VALUE_SET: return "set";
    case QWIC_VALUE_DICTIONARY: return "map";
    case QWIC_VALUE_TUPLE: return "tuple";
    case QWIC_VALUE_FUNCTION: return "function";
    case QWIC_VALUE_POINTER: return "pointer";
    }
    return "unknown";
}

char *qwic_any_to_string(qwic_value *value) {
    char buffer[64];
    if (value == NULL || value->kind == QWIC_VALUE_NULL) return qwic_copy_string("null");
    switch (value->kind) {
    case QWIC_VALUE_BOOL: return qwic_copy_string(value->as.boolean ? "true" : "false");
    case QWIC_VALUE_INT:
        snprintf(buffer, sizeof(buffer), "%lld", (long long)value->as.integer);
        return qwic_copy_string(buffer);
    case QWIC_VALUE_FLOAT:
        snprintf(buffer, sizeof(buffer), "%.6f", value->as.floating);
        return qwic_copy_string(buffer);
    case QWIC_VALUE_STRING: return qwic_copy_string(value->as.string);
    case QWIC_VALUE_LIST: return qwic_copy_string("[list]");
    case QWIC_VALUE_SET: return qwic_copy_string("{set}");
    case QWIC_VALUE_DICTIONARY: return qwic_copy_string("{map}");
    case QWIC_VALUE_TUPLE: return qwic_copy_string("(tuple)");
    case QWIC_VALUE_FUNCTION: return qwic_copy_string("<function>");
    case QWIC_VALUE_POINTER: return qwic_copy_string("<pointer>");
    default: return qwic_copy_string("");
    }
}

bool qwic_any_equal(qwic_value *left, qwic_value *right) {
    if (left == right) return true;
    if (left == NULL || right == NULL) return false;
    if ((left->kind == QWIC_VALUE_INT || left->kind == QWIC_VALUE_FLOAT) &&
        (right->kind == QWIC_VALUE_INT || right->kind == QWIC_VALUE_FLOAT)) {
        return qwic_any_as_float(left) == qwic_any_as_float(right);
    }
    if (left->kind != right->kind) return false;
    switch (left->kind) {
    case QWIC_VALUE_NULL: return true;
    case QWIC_VALUE_BOOL: return left->as.boolean == right->as.boolean;
    case QWIC_VALUE_INT: return left->as.integer == right->as.integer;
    case QWIC_VALUE_FLOAT: return left->as.floating == right->as.floating;
    case QWIC_VALUE_STRING: return strcmp(left->as.string, right->as.string) == 0;
    case QWIC_VALUE_FUNCTION: return left->as.function == right->as.function;
    default: return left->as.pointer == right->as.pointer;
    }
}

int qwic_any_compare(qwic_value *left, qwic_value *right) {
    if ((left != NULL && (left->kind == QWIC_VALUE_INT || left->kind == QWIC_VALUE_FLOAT)) &&
        (right != NULL && (right->kind == QWIC_VALUE_INT || right->kind == QWIC_VALUE_FLOAT))) {
        double a = qwic_any_as_float(left), b = qwic_any_as_float(right);
        return (a > b) - (a < b);
    }
    return strcmp(qwic_any_as_string(left), qwic_any_as_string(right));
}

qwic_value *qwic_any_add(qwic_value *left, qwic_value *right) {
    if ((left != NULL && left->kind == QWIC_VALUE_STRING) || (right != NULL && right->kind == QWIC_VALUE_STRING)) {
        const char *a = qwic_any_as_string(left), *b = qwic_any_as_string(right);
        size_t length = strlen(a) + strlen(b) + 1;
        char *joined = qwic_alloc(length);
        snprintf(joined, length, "%s%s", a, b);
        return qwic_any_string(joined);
    }
    if ((left != NULL && left->kind == QWIC_VALUE_FLOAT) || (right != NULL && right->kind == QWIC_VALUE_FLOAT)) {
        return qwic_any_float(qwic_any_as_float(left) + qwic_any_as_float(right));
    }
    return qwic_any_int(qwic_any_as_int(left) + qwic_any_as_int(right));
}

qwic_closure *qwic_closure_new(qwic_callback callback, void *context) {
    qwic_closure *closure = qwic_alloc(sizeof(qwic_closure));
    closure->callback = callback;
    closure->context = context;
    return closure;
}

qwic_value *qwic_closure_call(qwic_closure *closure, qwic_value **args, size_t count) {
    if (closure == NULL || closure->callback == NULL) return qwic_any_null();
    return closure->callback(closure->context, args, count);
}

qwic_try_frame *qwic_try_new(void) {
    qwic_try_frame *frame = qwic_alloc(sizeof(qwic_try_frame));
    if (frame == NULL) {
        fprintf(stderr, "fatal: unable to allocate exception frame\n");
        exit(1);
    }
    frame->previous = NULL;
    frame->message = NULL;
    return frame;
}

void qwic_try_push(qwic_try_frame *frame) {
    frame->previous = qwic_active_try_frame;
    qwic_active_try_frame = frame;
}

void qwic_try_end(qwic_try_frame *frame) {
    if (qwic_active_try_frame == frame) {
        qwic_active_try_frame = frame->previous;
    }
    qwic_free(frame);
}

const char *qwic_try_message(qwic_try_frame *frame) {
    if (frame == NULL || frame->message == NULL) {
        return "";
    }
    return frame->message;
}

void qwic_try_caught(qwic_try_frame *frame) {
    qwic_free(frame);
}

void qwic_throw(const char *message) {
    qwic_try_frame *frame = qwic_active_try_frame;
    if (frame == NULL) {
        fprintf(stderr, "uncaught error: %s\n", message != NULL ? message : "");
        exit(1);
    }
    frame->message = message != NULL ? message : "";
    qwic_active_try_frame = frame->previous;
    longjmp(frame->environment, 1);
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
    printf("%s\n", value != NULL ? value : "");
}

void qwic_print_list(void *value) {
    if (value == NULL) {
        printf("[]\n");
        return;
    }
    qwic_list *list = (qwic_list *)value;
    putchar('[');
    for (size_t i = 0; i < list->length; i++) {
        if (i > 0) {
            printf(", ");
        }
        printf("%s", qwic_any_to_string(list->items[i]));
    }
    putchar(']');
    putchar('\n');
}

void qwic_print_set(void *value) {
    if (value == NULL) {
        printf("{}\n");
        return;
    }
    qwic_set *set = (qwic_set *)value;
    putchar('{');
    for (size_t i = 0; i < set->values.length; i++) {
        if (i > 0) {
            printf(", ");
        }
        printf("%s", qwic_any_to_string(set->values.items[i]));
    }
    putchar('}');
    putchar('\n');
}

void qwic_print_dictionary(void *value) {
    if (value == NULL) {
        printf("{}\n");
        return;
    }
    qwic_dictionary *dict = (qwic_dictionary *)value;
    putchar('{');
    for (size_t i = 0; i < dict->length; i++) {
        if (i > 0) {
            printf(", ");
        }
        printf("%s: %s", dict->entries[i].key != NULL ? dict->entries[i].key : "", qwic_any_to_string(dict->entries[i].value));
    }
    putchar('}');
    putchar('\n');
}

void qwic_print_tuple(void *value) {
    if (value == NULL) {
        printf("()\n");
        return;
    }
    qwic_tuple *tuple = (qwic_tuple *)value;
    putchar('(');
    for (size_t i = 0; i < tuple->length; i++) {
        if (i > 0) {
            printf(", ");
        }
        printf("%s", qwic_any_to_string(tuple->items[i]));
    }
    putchar(')');
    putchar('\n');
}

void qwic_print_any(qwic_value *value) {
    printf("%s\n", qwic_any_to_string(value));
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

char *qwic_strings_concat(const char *left, const char *right) {
    if (left == NULL) left = "";
    if (right == NULL) right = "";
    size_t length = strlen(left) + strlen(right) + 1;
    char *result = qwic_alloc(length);
    snprintf(result, length, "%s%s", left, right);
    return result;
}

void *qwic_strings_split(const char *value, const char *separator) {
    qwic_list *result = qwic_lists_new();
    if (value == NULL) return result;
    if (separator == NULL || separator[0] == '\0') {
        for (const char *cursor = value; *cursor != '\0'; cursor++) {
            char part[2] = {*cursor, '\0'};
            qwic_lists_push(result, qwic_any_string(part));
        }
        return result;
    }
    const char *start = value;
    size_t separator_length = strlen(separator);
    for (;;) {
        const char *match = strstr(start, separator);
        if (match == NULL) {
            qwic_lists_push(result, qwic_any_string(start));
            break;
        }
        char *part = qwic_copy_range(start, match);
        qwic_lists_push(result, qwic_any_string(part));
        start = match + separator_length;
    }
    return result;
}

char *qwic_strings_substring(const char *value, int64_t start, int64_t end) {
    if (value == NULL) return qwic_copy_string("");
    int64_t length = (int64_t)strlen(value);
    if (start < 0) start = 0;
    if (end < start) end = start;
    if (start > length) start = length;
    if (end > length) end = length;
    return qwic_copy_range(value + start, value + end);
}

int64_t qwic_strings_to_int(const char *value) {
    if (value == NULL) return 0;
    return strtoll(value, NULL, 10);
}

static void qwic_json_skip_space(qwic_json_reader *reader) {
    while (isspace((unsigned char)*reader->cursor)) reader->cursor++;
}

static char *qwic_json_read_string(qwic_json_reader *reader) {
    qwic_string_builder builder = {0};
    if (*reader->cursor != '"') return qwic_copy_string("");
    reader->cursor++;
    while (*reader->cursor != '\0' && *reader->cursor != '"') {
        char current = *reader->cursor++;
        if (current == '\\' && *reader->cursor != '\0') {
            char escaped = *reader->cursor++;
            switch (escaped) {
            case 'n': current = '\n'; break;
            case 'r': current = '\r'; break;
            case 't': current = '\t'; break;
            case 'b': current = '\b'; break;
            case 'f': current = '\f'; break;
            default: current = escaped; break;
            }
        }
        qwic_builder_append_char(&builder, current);
    }
    if (*reader->cursor == '"') reader->cursor++;
    if (builder.data == NULL) return qwic_copy_string("");
    return builder.data;
}

static qwic_value *qwic_json_read_array(qwic_json_reader *reader) {
    void *list = qwic_lists_new();
    reader->cursor++;
    qwic_json_skip_space(reader);
    if (*reader->cursor == ']') {
        reader->cursor++;
        return qwic_any_list(list);
    }
    for (;;) {
        qwic_lists_push(list, qwic_json_read_value(reader));
        qwic_json_skip_space(reader);
        if (*reader->cursor == ']') {
            reader->cursor++;
            break;
        }
        if (*reader->cursor != ',') break;
        reader->cursor++;
        qwic_json_skip_space(reader);
    }
    return qwic_any_list(list);
}

static qwic_value *qwic_json_read_object(qwic_json_reader *reader) {
    void *dictionary = qwic_dictionaries_new();
    reader->cursor++;
    qwic_json_skip_space(reader);
    if (*reader->cursor == '}') {
        reader->cursor++;
        return qwic_any_dictionary(dictionary);
    }
    for (;;) {
        qwic_json_skip_space(reader);
        char *key = qwic_json_read_string(reader);
        qwic_json_skip_space(reader);
        if (*reader->cursor == ':') reader->cursor++;
        qwic_json_skip_space(reader);
        qwic_dictionaries_set(dictionary, key, qwic_json_read_value(reader));
        qwic_json_skip_space(reader);
        if (*reader->cursor == '}') {
            reader->cursor++;
            break;
        }
        if (*reader->cursor != ',') break;
        reader->cursor++;
    }
    return qwic_any_dictionary(dictionary);
}

static qwic_value *qwic_json_read_value(qwic_json_reader *reader) {
    qwic_json_skip_space(reader);
    if (*reader->cursor == '"') return qwic_any_string(qwic_json_read_string(reader));
    if (*reader->cursor == '{') return qwic_json_read_object(reader);
    if (*reader->cursor == '[') return qwic_json_read_array(reader);
    if (strncmp(reader->cursor, "true", 4) == 0) {
        reader->cursor += 4;
        return qwic_any_bool(true);
    }
    if (strncmp(reader->cursor, "false", 5) == 0) {
        reader->cursor += 5;
        return qwic_any_bool(false);
    }
    if (strncmp(reader->cursor, "null", 4) == 0) {
        reader->cursor += 4;
        return qwic_any_null();
    }
    const char *start = reader->cursor;
    char *end = NULL;
    double number = strtod(start, &end);
    if (end != start) {
        bool floating = false;
        for (const char *cursor = start; cursor < end; cursor++) {
            if (*cursor == '.' || *cursor == 'e' || *cursor == 'E') floating = true;
        }
        reader->cursor = end;
        return floating ? qwic_any_float(number) : qwic_any_int((int64_t)number);
    }
    return qwic_any_null();
}

qwic_value *qwic_json_parse(const char *json) {
    qwic_json_reader reader = {json != NULL ? json : "null"};
    return qwic_json_read_value(&reader);
}

static void qwic_json_write_string(qwic_string_builder *builder, const char *value) {
    qwic_builder_append_char(builder, '"');
    if (value != NULL) {
        for (const char *cursor = value; *cursor != '\0'; cursor++) {
            switch (*cursor) {
            case '"': qwic_builder_append(builder, "\\\""); break;
            case '\\': qwic_builder_append(builder, "\\\\"); break;
            case '\n': qwic_builder_append(builder, "\\n"); break;
            case '\r': qwic_builder_append(builder, "\\r"); break;
            case '\t': qwic_builder_append(builder, "\\t"); break;
            default: qwic_builder_append_char(builder, *cursor); break;
            }
        }
    }
    qwic_builder_append_char(builder, '"');
}

static void qwic_json_write_value(qwic_string_builder *builder, qwic_value *value) {
    char number[64];
    if (value == NULL || value->kind == QWIC_VALUE_NULL) {
        qwic_builder_append(builder, "null");
        return;
    }
    switch (value->kind) {
    case QWIC_VALUE_BOOL:
        qwic_builder_append(builder, value->as.boolean ? "true" : "false");
        break;
    case QWIC_VALUE_INT:
        snprintf(number, sizeof(number), "%lld", (long long)value->as.integer);
        qwic_builder_append(builder, number);
        break;
    case QWIC_VALUE_FLOAT:
        snprintf(number, sizeof(number), "%.15g", value->as.floating);
        qwic_builder_append(builder, number);
        break;
    case QWIC_VALUE_STRING:
        qwic_json_write_string(builder, value->as.string);
        break;
    case QWIC_VALUE_LIST: {
        qwic_list *list = value->as.pointer;
        qwic_builder_append_char(builder, '[');
        if (list != NULL) for (size_t index = 0; index < list->length; index++) {
            if (index > 0) qwic_builder_append_char(builder, ',');
            qwic_json_write_value(builder, list->items[index]);
        }
        qwic_builder_append_char(builder, ']');
        break;
    }
    case QWIC_VALUE_DICTIONARY: {
        qwic_dictionary *dictionary = value->as.pointer;
        qwic_builder_append_char(builder, '{');
        if (dictionary != NULL) for (size_t index = 0; index < dictionary->length; index++) {
            if (index > 0) qwic_builder_append_char(builder, ',');
            qwic_json_write_string(builder, dictionary->entries[index].key);
            qwic_builder_append_char(builder, ':');
            qwic_json_write_value(builder, dictionary->entries[index].value);
        }
        qwic_builder_append_char(builder, '}');
        break;
    }
    default:
        qwic_builder_append(builder, "null");
        break;
    }
}

char *qwic_json_stringify(qwic_value *value) {
    qwic_string_builder builder = {0};
    qwic_json_write_value(&builder, value);
    if (builder.data == NULL) return qwic_copy_string("null");
    return builder.data;
}

char *qwic_json_encode(qwic_value *value) { return qwic_json_stringify(value); }
qwic_value *qwic_json_decode(const char *json) { return qwic_json_parse(json); }

static void qwic_builder_reserve(qwic_string_builder *builder, size_t needed) {
    if (builder->capacity >= needed) return;
    size_t capacity = builder->capacity == 0 ? 64 : builder->capacity * 2;
    while (capacity < needed) capacity *= 2;
    char *data = qwic_alloc(capacity);
    if (builder->data != NULL) {
        memcpy(data, builder->data, builder->length + 1);
        qwic_free(builder->data);
    } else {
        data[0] = '\0';
    }
    builder->data = data;
    builder->capacity = capacity;
}

static void qwic_builder_append(qwic_string_builder *builder, const char *value) {
    size_t length = strlen(value);
    qwic_builder_reserve(builder, builder->length + length + 1);
    memcpy(builder->data + builder->length, value, length + 1);
    builder->length += length;
}

static void qwic_builder_append_char(qwic_string_builder *builder, char value) {
    qwic_builder_reserve(builder, builder->length + 2);
    builder->data[builder->length++] = value;
    builder->data[builder->length] = '\0';
}

static bool qwic_ascii_case_prefix(const char *value, const char *prefix) {
    while (*prefix != '\0') {
        if (*value == '\0' || tolower((unsigned char)*value) != tolower((unsigned char)*prefix)) return false;
        value++;
        prefix++;
    }
    return true;
}

static const char *qwic_http_reason(int64_t status) {
    switch (status) {
    case 200: return "OK";
    case 201: return "Created";
    case 204: return "No Content";
    case 400: return "Bad Request";
    case 404: return "Not Found";
    case 500: return "Internal Server Error";
    default: return "OK";
    }
}

static bool qwic_socket_send_all(qwic_socket socket_value, const char *data, size_t length) {
    size_t sent = 0;
    while (sent < length) {
#if defined(_WIN32)
        int count = send(socket_value, data + sent, (int)(length - sent), 0);
#else
        ssize_t count = send(socket_value, data + sent, length - sent, 0);
#endif
        if (count <= 0) return false;
        sent += (size_t)count;
    }
    return true;
}

static qwic_value *qwic_http_read_request(qwic_socket client) {
    const size_t capacity = 1024 * 1024;
    char *buffer = qwic_alloc(capacity + 1);
    size_t length = 0;
    char *headers_end = NULL;
    int64_t content_length = 0;
    while (length < capacity) {
#if defined(_WIN32)
        int count = recv(client, buffer + length, (int)(capacity - length), 0);
#else
        ssize_t count = recv(client, buffer + length, capacity - length, 0);
#endif
        if (count <= 0) break;
        length += (size_t)count;
        buffer[length] = '\0';
        headers_end = strstr(buffer, "\r\n\r\n");
        if (headers_end != NULL) {
            const char *header = buffer;
            while ((header = strstr(header, "\r\n")) != NULL && header < headers_end) {
                header += 2;
                if (header >= headers_end) break;
                if (qwic_ascii_case_prefix(header, "Content-Length:")) content_length = strtoll(header + 15, NULL, 10);
            }
            size_t header_length = (size_t)(headers_end + 4 - buffer);
            if (length >= header_length + (size_t)content_length) break;
        }
    }

    void *headers = qwic_dictionaries_new();
    void *params = qwic_dictionaries_new();
    const char *method = "GET";
    const char *path = "/";
    const char *body = "";
    if (headers_end != NULL) {
        *headers_end = '\0';
        char *line_end = strstr(buffer, "\r\n");
        if (line_end != NULL) {
            *line_end = '\0';
            char *first_space = strchr(buffer, ' ');
            if (first_space != NULL) {
                *first_space = '\0';
                method = buffer;
                char *second_space = strchr(first_space + 1, ' ');
                if (second_space != NULL) {
                    *second_space = '\0';
                    path = first_space + 1;
                    char *query = strchr((char *)path, '?');
                    if (query != NULL) *query = '\0';
                }
            }
            char *line = line_end + 2;
            while (*line != '\0') {
                char *next = strstr(line, "\r\n");
                if (next != NULL) *next = '\0';
                char *colon = strchr(line, ':');
                if (colon != NULL) {
                    *colon = '\0';
                    char *header_value = colon + 1;
                    while (*header_value == ' ' || *header_value == '\t') header_value++;
                    qwic_dictionaries_set(headers, line, qwic_any_string(header_value));
                }
                if (next == NULL) break;
                line = next + 2;
            }
        }
        body = headers_end + 4;
    }

    void *body_value = qwic_dictionaries_new();
    qwic_dictionaries_set(body_value, "raw", qwic_any_string(body));
    void *request = qwic_dictionaries_new();
    qwic_dictionaries_set(request, "method", qwic_any_string(method));
    qwic_dictionaries_set(request, "path", qwic_any_string(path));
    qwic_dictionaries_set(request, "params", qwic_any_dictionary(params));
    qwic_dictionaries_set(request, "body", qwic_any_dictionary(body_value));
    qwic_dictionaries_set(request, "headers", qwic_any_dictionary(headers));
    return qwic_any_dictionary(request);
}

static void qwic_http_write_response(qwic_socket client, qwic_value *response) {
    int64_t status = qwic_any_as_int(qwic_any_field(response, "status"));
    if (status == 0) status = 200;
    qwic_value *body_value = qwic_any_field(response, "body");
    char *body = qwic_json_stringify(body_value);
    const char *content_type = "application/json";
    qwic_value *headers = qwic_any_field(response, "headers");
    qwic_value *configured_type = qwic_any_field(headers, "Content-Type");
    if (configured_type != NULL && strcmp(qwic_any_type(configured_type), "string") == 0) {
        content_type = qwic_any_as_string(configured_type);
    }
    char header[1024];
    int header_length = snprintf(header, sizeof(header),
        "HTTP/1.1 %lld %s\r\nContent-Type: %s\r\nContent-Length: %llu\r\nConnection: close\r\n\r\n",
        (long long)status, qwic_http_reason(status), content_type, (unsigned long long)strlen(body));
    if (header_length > 0) {
        qwic_socket_send_all(client, header, (size_t)header_length);
        qwic_socket_send_all(client, body, strlen(body));
    }
}

void qwic_http_listen(int64_t port, qwic_closure *handler) {
    qwic_socket server = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (server == QWIC_INVALID_SOCKET) {
        fprintf(stderr, "http.listen: unable to create socket\n");
        qwic_exit_code = 1;
        return;
    }
    int reuse = 1;
#if defined(_WIN32)
    setsockopt(server, SOL_SOCKET, SO_REUSEADDR, (const char *)&reuse, sizeof(reuse));
#else
    setsockopt(server, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));
#endif
    struct sockaddr_in address;
    memset(&address, 0, sizeof(address));
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = htonl(INADDR_ANY);
    address.sin_port = htons((uint16_t)port);
    if (bind(server, (struct sockaddr *)&address, sizeof(address)) != 0 || listen(server, 128) != 0) {
        fprintf(stderr, "http.listen: unable to listen on port %lld\n", (long long)port);
        qwic_close_socket(server);
        qwic_exit_code = 1;
        return;
    }
    for (;;) {
        qwic_socket client = accept(server, NULL, NULL);
        if (client == QWIC_INVALID_SOCKET) continue;
        qwic_value *request = qwic_http_read_request(client);
        qwic_value *arguments[1] = {request};
        qwic_value *response = qwic_closure_call(handler, arguments, 1);
        qwic_http_write_response(client, response);
        qwic_close_socket(client);
    }
}

double qwic_math_abs(double x) {
    return fabs(x);
}

double qwic_math_min(double a, double b) {
    return fmin(a, b);
}

double qwic_math_max(double a, double b) {
    return fmax(a, b);
}

double qwic_math_sqrt(double x) {
    return sqrt(x);
}

double qwic_math_pow(double base, double exp) {
    return pow(base, exp);
}

double qwic_math_floor(double x) {
    return floor(x);
}

double qwic_math_ceil(double x) {
    return ceil(x);
}

double qwic_math_round(double x) {
    return round(x);
}

int64_t qwic_time_now(void) {
    return (int64_t)time(NULL);
}

void qwic_time_sleep(int64_t ms) {
#if defined(_WIN32)
    if (ms > 0) {
        Sleep(ms > MAXDWORD ? MAXDWORD : (DWORD)ms);
    }
#else
    if (ms > 0) {
        struct timespec duration;
        duration.tv_sec = (time_t)(ms / 1000);
        duration.tv_nsec = (long)((ms % 1000) * 1000000);
        nanosleep(&duration, NULL);
    }
#endif
}

int64_t qwic_time_duration(int64_t start, int64_t end) {
    return end - start;
}

const char *qwic_fs_read_file(const char *path) {
    if (path == NULL) return "";
    
    FILE *file = fopen(path, "r");
    if (file == NULL) return "";

    fseek(file, 0, SEEK_END);
    long length = ftell(file);
    fseek(file, 0, SEEK_SET);

    char *buffer = qwic_alloc(length + 1);
    if (buffer == NULL) {
        fclose(file);
        return "";
    }
    
    size_t read_bytes = fread(buffer, 1, length, file);
    buffer[read_bytes] = '\0';
    fclose(file);
    return buffer;
}

bool qwic_fs_write_file(const char *path, const char *content) {
    if (path == NULL || content == NULL) return false;
    
    FILE *file = fopen(path, "w");
    if (file == NULL) return false;
    
    size_t written = fwrite(content, 1, strlen(content), file);
    fclose(file);
    return written == strlen(content);
}

bool qwic_fs_exists(const char *path) {
    if (path == NULL) return false;
#if defined(_WIN32)
    return _access(path, 0) == 0;
#else
    return access(path, F_OK) == 0;
#endif
}

void *qwic_sync_mutex_new(void) {
#if defined(_WIN32)
    CRITICAL_SECTION *mutex = qwic_alloc(sizeof(CRITICAL_SECTION));
    if (mutex != NULL) {
        InitializeCriticalSection(mutex);
    }
#else
    pthread_mutex_t *mutex = qwic_alloc(sizeof(pthread_mutex_t));
    if (mutex != NULL) {
        pthread_mutex_init(mutex, NULL);
    }
#endif
    return mutex;
}

void qwic_sync_mutex_lock(void *mutex) {
    if (mutex != NULL) {
#if defined(_WIN32)
        EnterCriticalSection((CRITICAL_SECTION *)mutex);
#else
        pthread_mutex_lock((pthread_mutex_t *)mutex);
#endif
    }
}

void qwic_sync_mutex_unlock(void *mutex) {
    if (mutex != NULL) {
#if defined(_WIN32)
        LeaveCriticalSection((CRITICAL_SECTION *)mutex);
#else
        pthread_mutex_unlock((pthread_mutex_t *)mutex);
#endif
    }
}

void qwic_sync_mutex_free(void *mutex) {
    if (mutex != NULL) {
#if defined(_WIN32)
        DeleteCriticalSection((CRITICAL_SECTION *)mutex);
#else
        pthread_mutex_destroy((pthread_mutex_t *)mutex);
#endif
        qwic_free(mutex);
    }
}

void *qwic_lists_new(void) {
    qwic_list *list = qwic_alloc(sizeof(qwic_list));
    qwic_list_init(list);
    return list;
}

void qwic_lists_push(void *list, qwic_value *value) {
    if (list == NULL) {
        return;
    }
    qwic_list_push_value((qwic_list *)list, value);
}

qwic_value *qwic_lists_get(void *list, int64_t index) {
    if (list == NULL || index < 0) {
        return qwic_any_null();
    }
    qwic_list *typed = (qwic_list *)list;
    if ((size_t)index >= typed->length) {
        return qwic_any_null();
    }
    return typed->items[index];
}

int64_t qwic_lists_length(void *list) {
    if (list == NULL) {
        return 0;
    }
    return (int64_t)((qwic_list *)list)->length;
}

bool qwic_lists_contains(void *list, qwic_value *value) {
    if (list == NULL) {
        return false;
    }
    return qwic_list_contains((qwic_list *)list, value);
}

void *qwic_lists_slice(void *list, int64_t start, int64_t end) {
    qwic_list *result = qwic_lists_new();
    if (list == NULL) return result;
    qwic_list *source = list;
    if (start < 0) start = 0;
    if (end > (int64_t)source->length) end = (int64_t)source->length;
    if (end < start) end = start;
    for (int64_t index = start; index < end; index++) {
        qwic_list_push_value(result, source->items[index]);
    }
    return result;
}

void *qwic_lists_concat(void *left, void *right) {
    qwic_list *result = qwic_lists_new();
    qwic_list *first = left, *second = right;
    if (first != NULL) {
        for (size_t index = 0; index < first->length; index++) qwic_list_push_value(result, first->items[index]);
    }
    if (second != NULL) {
        for (size_t index = 0; index < second->length; index++) qwic_list_push_value(result, second->items[index]);
    }
    return result;
}

void *qwic_sets_new(void) {
    qwic_set *set = qwic_alloc(sizeof(qwic_set));
    qwic_list_init(&set->values);
    return set;
}

void qwic_sets_add(void *set, qwic_value *value) {
    if (set == NULL) {
        return;
    }
    qwic_set *typed = (qwic_set *)set;
    if (!qwic_list_contains(&typed->values, value)) {
        qwic_list_push_value(&typed->values, value);
    }
}

bool qwic_sets_contains(void *set, qwic_value *value) {
    if (set == NULL) {
        return false;
    }
    return qwic_list_contains(&((qwic_set *)set)->values, value);
}

int64_t qwic_sets_length(void *set) {
    if (set == NULL) {
        return 0;
    }
    return (int64_t)((qwic_set *)set)->values.length;
}

void *qwic_dictionaries_new(void) {
    qwic_dictionary *dictionary = qwic_alloc(sizeof(qwic_dictionary));
    dictionary->length = 0;
    dictionary->capacity = 0;
    dictionary->entries = NULL;
    return dictionary;
}

void qwic_dictionaries_set(void *dictionary, const char *key, qwic_value *value) {
    if (dictionary == NULL) {
        return;
    }
    qwic_dictionary *typed = (qwic_dictionary *)dictionary;
    int64_t existing = qwic_dictionary_find(typed, key);
    if (existing >= 0) {
        typed->entries[existing].value = value != NULL ? value : qwic_any_null();
        return;
    }
    qwic_dictionary_reserve(typed, typed->length + 1);
    typed->entries[typed->length].key = qwic_copy_string(key);
    typed->entries[typed->length].value = value != NULL ? value : qwic_any_null();
    typed->length++;
}

qwic_value *qwic_dictionaries_get(void *dictionary, const char *key) {
    if (dictionary == NULL) {
        return qwic_any_null();
    }
    qwic_dictionary *typed = (qwic_dictionary *)dictionary;
    int64_t index = qwic_dictionary_find(typed, key);
    if (index < 0) {
        return qwic_any_null();
    }
    return typed->entries[index].value;
}

bool qwic_dictionaries_contains(void *dictionary, const char *key) {
    if (dictionary == NULL) {
        return false;
    }
    return qwic_dictionary_find((qwic_dictionary *)dictionary, key) >= 0;
}

int64_t qwic_dictionaries_length(void *dictionary) {
    if (dictionary == NULL) {
        return 0;
    }
    return (int64_t)((qwic_dictionary *)dictionary)->length;
}

void *qwic_dictionaries_keys(void *dictionary) {
    qwic_list *keys = qwic_lists_new();
    if (dictionary == NULL) return keys;
    qwic_dictionary *typed = dictionary;
    for (size_t index = 0; index < typed->length; index++) {
        qwic_list_push_value(keys, qwic_any_string(typed->entries[index].key));
    }
    return keys;
}

qwic_value *qwic_any_field(qwic_value *value, const char *field) {
    if (value == NULL || value->kind != QWIC_VALUE_DICTIONARY) return qwic_any_null();
    return qwic_dictionaries_get(value->as.pointer, field);
}

void qwic_any_set_field(qwic_value *value, const char *field, qwic_value *item) {
    if (value == NULL || value->kind != QWIC_VALUE_DICTIONARY) return;
    qwic_dictionaries_set(value->as.pointer, field, item);
}

qwic_value *qwic_any_index(qwic_value *value, qwic_value *index) {
    if (value == NULL) return qwic_any_null();
    if (value->kind == QWIC_VALUE_DICTIONARY) {
        return qwic_dictionaries_get(value->as.pointer, qwic_any_as_string(index));
    }
    if (value->kind == QWIC_VALUE_LIST) {
        return qwic_lists_get(value->as.pointer, qwic_any_as_int(index));
    }
    if (value->kind == QWIC_VALUE_TUPLE) {
        int64_t position = qwic_any_as_int(index);
        qwic_tuple *tuple = value->as.pointer;
        if (tuple == NULL || position < 0 || (size_t)position >= tuple->length) return qwic_any_null();
        return tuple->items[position];
    }
    if (value->kind == QWIC_VALUE_STRING) {
        int64_t position = qwic_any_as_int(index);
        size_t length = strlen(value->as.string);
        if (position < 0 || (size_t)position >= length) return qwic_any_string("");
        char character[2] = {value->as.string[position], '\0'};
        return qwic_any_string(character);
    }
    return qwic_any_null();
}

void qwic_any_set_index(qwic_value *value, qwic_value *index, qwic_value *item) {
    if (value == NULL) return;
    if (value->kind == QWIC_VALUE_DICTIONARY) {
        qwic_dictionaries_set(value->as.pointer, qwic_any_as_string(index), item);
        return;
    }
    if (value->kind == QWIC_VALUE_LIST) {
        qwic_list *list = value->as.pointer;
        int64_t position = qwic_any_as_int(index);
        if (list != NULL && position >= 0 && (size_t)position < list->length) list->items[position] = item;
    }
}

void *qwic_tuples_new2(qwic_value *first, qwic_value *second) {
    qwic_tuple *tuple = qwic_alloc(sizeof(qwic_tuple));
    tuple->length = 2;
    tuple->items = qwic_alloc(sizeof(qwic_value *) * tuple->length);
    tuple->items[0] = first != NULL ? first : qwic_any_null();
    tuple->items[1] = second != NULL ? second : qwic_any_null();
    return tuple;
}

qwic_value *qwic_tuples_first(void *tuple) {
    if (tuple == NULL) {
        return qwic_any_null();
    }
    return ((qwic_tuple *)tuple)->items[0];
}

qwic_value *qwic_tuples_second(void *tuple) {
    if (tuple == NULL) {
        return qwic_any_null();
    }
    return ((qwic_tuple *)tuple)->items[1];
}

int64_t qwic_tuples_length(void *tuple) {
    if (tuple == NULL) {
        return 0;
    }
    return (int64_t)((qwic_tuple *)tuple)->length;
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

static char *qwic_copy_string(const char *value) {
    if (value == NULL) {
        value = "";
    }
    return qwic_copy_range(value, value + strlen(value));
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

static void qwic_list_init(qwic_list *list) {
    list->length = 0;
    list->capacity = 0;
    list->items = NULL;
}

static void qwic_list_push_value(qwic_list *list, qwic_value *value) {
    qwic_list_reserve(list, list->length + 1);
    list->items[list->length] = value != NULL ? value : qwic_any_null();
    list->length++;
}

static bool qwic_list_contains(qwic_list *list, qwic_value *value) {
    for (size_t index = 0; index < list->length; index++) {
        if (qwic_any_equal(list->items[index], value)) {
            return true;
        }
    }
    return false;
}

static void qwic_list_reserve(qwic_list *list, size_t needed) {
    if (list->capacity >= needed) {
        return;
    }
    size_t next_capacity = list->capacity == 0 ? 4 : list->capacity * 2;
    while (next_capacity < needed) {
        next_capacity *= 2;
    }
    qwic_value **items = qwic_alloc(sizeof(qwic_value *) * next_capacity);
    if (list->items != NULL) {
        memcpy(items, list->items, sizeof(qwic_value *) * list->length);
    }
    list->items = items;
    list->capacity = next_capacity;
}

static int64_t qwic_dictionary_find(qwic_dictionary *dictionary, const char *key) {
    if (key == NULL) {
        key = "";
    }
    for (size_t index = 0; index < dictionary->length; index++) {
        if (strcmp(dictionary->entries[index].key, key) == 0) {
            return (int64_t)index;
        }
    }
    return -1;
}

static void qwic_dictionary_reserve(qwic_dictionary *dictionary, size_t needed) {
    if (dictionary->capacity >= needed) {
        return;
    }
    size_t next_capacity = dictionary->capacity == 0 ? 4 : dictionary->capacity * 2;
    while (next_capacity < needed) {
        next_capacity *= 2;
    }
    qwic_dictionary_entry *entries = qwic_alloc(sizeof(qwic_dictionary_entry) * next_capacity);
    if (dictionary->entries != NULL) {
        memcpy(entries, dictionary->entries, sizeof(qwic_dictionary_entry) * dictionary->length);
    }
    dictionary->entries = entries;
    dictionary->capacity = next_capacity;
}
