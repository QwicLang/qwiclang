#include "qwic_runtime.h"

#include <stdlib.h>
#include <stdio.h>
#include <ctype.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <fcntl.h>
#include <pthread.h>

static int qwic_exit_code = 0;

typedef struct {
    size_t length;
    size_t capacity;
    char **items;
} qwic_list;

typedef struct {
    qwic_list values;
} qwic_set;

typedef struct {
    char *key;
    char *value;
} qwic_dictionary_entry;

typedef struct {
    size_t length;
    size_t capacity;
    qwic_dictionary_entry *entries;
} qwic_dictionary;

typedef struct {
    size_t length;
    char **items;
} qwic_tuple;

static const char *qwic_trim_left_start(const char *value);
static const char *qwic_trim_right_end(const char *value);
static char *qwic_copy_range(const char *start, const char *end);
static char *qwic_ascii_map(const char *value, bool uppercase);
static char *qwic_copy_string(const char *value);
static void qwic_list_init(qwic_list *list);
static void qwic_list_push_copy(qwic_list *list, const char *value);
static bool qwic_list_contains(qwic_list *list, const char *value);
static void qwic_list_reserve(qwic_list *list, size_t needed);
static int64_t qwic_dictionary_find(qwic_dictionary *dictionary, const char *key);
static void qwic_dictionary_reserve(qwic_dictionary *dictionary, size_t needed);

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

// ... (previous functions)

int64_t qwic_time_now(void) {
    return (int64_t)time(NULL);
}

void qwic_time_sleep(int64_t ms) {
    usleep(ms * 1000);
}

int64_t qwic_time_duration(int64_t start, int64_t end) {
    return end - start;
}

// ... (previous functions)


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
    return access(path, F_OK) == 0;
}

void *qwic_sync_mutex_new(void) {
    pthread_mutex_t *mutex = qwic_alloc(sizeof(pthread_mutex_t));
    pthread_mutex_init(mutex, NULL);
    return mutex;
}

void qwic_sync_mutex_lock(void *mutex) {
    if (mutex != NULL) {
        pthread_mutex_lock((pthread_mutex_t *)mutex);
    }
}

void qwic_sync_mutex_unlock(void *mutex) {
    if (mutex != NULL) {
        pthread_mutex_unlock((pthread_mutex_t *)mutex);
    }
}

void qwic_sync_mutex_free(void *mutex) {
    if (mutex != NULL) {
        pthread_mutex_destroy((pthread_mutex_t *)mutex);
        qwic_free(mutex);
    }
}

void *qwic_lists_new(void) {
    qwic_list *list = qwic_alloc(sizeof(qwic_list));
    qwic_list_init(list);
    return list;
}

void qwic_lists_push(void *list, const char *value) {
    if (list == NULL) {
        return;
    }
    qwic_list_push_copy((qwic_list *)list, value);
}

const char *qwic_lists_get(void *list, int64_t index) {
    if (list == NULL || index < 0) {
        return "";
    }
    qwic_list *typed = (qwic_list *)list;
    if ((size_t)index >= typed->length) {
        return "";
    }
    return typed->items[index];
}

int64_t qwic_lists_length(void *list) {
    if (list == NULL) {
        return 0;
    }
    return (int64_t)((qwic_list *)list)->length;
}

bool qwic_lists_contains(void *list, const char *value) {
    if (list == NULL) {
        return false;
    }
    return qwic_list_contains((qwic_list *)list, value);
}

void *qwic_sets_new(void) {
    qwic_set *set = qwic_alloc(sizeof(qwic_set));
    qwic_list_init(&set->values);
    return set;
}

void qwic_sets_add(void *set, const char *value) {
    if (set == NULL) {
        return;
    }
    qwic_set *typed = (qwic_set *)set;
    if (!qwic_list_contains(&typed->values, value)) {
        qwic_list_push_copy(&typed->values, value);
    }
}

bool qwic_sets_contains(void *set, const char *value) {
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

void qwic_dictionaries_set(void *dictionary, const char *key, const char *value) {
    if (dictionary == NULL) {
        return;
    }
    qwic_dictionary *typed = (qwic_dictionary *)dictionary;
    int64_t existing = qwic_dictionary_find(typed, key);
    if (existing >= 0) {
        typed->entries[existing].value = qwic_copy_string(value);
        return;
    }
    qwic_dictionary_reserve(typed, typed->length + 1);
    typed->entries[typed->length].key = qwic_copy_string(key);
    typed->entries[typed->length].value = qwic_copy_string(value);
    typed->length++;
}

const char *qwic_dictionaries_get(void *dictionary, const char *key) {
    if (dictionary == NULL) {
        return "";
    }
    qwic_dictionary *typed = (qwic_dictionary *)dictionary;
    int64_t index = qwic_dictionary_find(typed, key);
    if (index < 0) {
        return "";
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

void *qwic_tuples_new2(const char *first, const char *second) {
    qwic_tuple *tuple = qwic_alloc(sizeof(qwic_tuple));
    tuple->length = 2;
    tuple->items = qwic_alloc(sizeof(char *) * tuple->length);
    tuple->items[0] = qwic_copy_string(first);
    tuple->items[1] = qwic_copy_string(second);
    return tuple;
}

const char *qwic_tuples_first(void *tuple) {
    if (tuple == NULL) {
        return "";
    }
    return ((qwic_tuple *)tuple)->items[0];
}

const char *qwic_tuples_second(void *tuple) {
    if (tuple == NULL) {
        return "";
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

static void qwic_list_push_copy(qwic_list *list, const char *value) {
    qwic_list_reserve(list, list->length + 1);
    list->items[list->length] = qwic_copy_string(value);
    list->length++;
}

static bool qwic_list_contains(qwic_list *list, const char *value) {
    if (value == NULL) {
        value = "";
    }
    for (size_t index = 0; index < list->length; index++) {
        if (strcmp(list->items[index], value) == 0) {
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
    char **items = qwic_alloc(sizeof(char *) * next_capacity);
    if (list->items != NULL) {
        memcpy(items, list->items, sizeof(char *) * list->length);
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
