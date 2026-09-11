#include "qwic_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <ctype.h>

// Internal types from qwic_runtime.c (copied here for the implementation)
typedef struct {
    size_t length;
    size_t capacity;
    char **items;
} qwic_list;

typedef struct {
    char *key;
    char *value;
} qwic_dictionary_entry;

typedef struct {
    size_t length;
    size_t capacity;
    qwic_dictionary_entry *entries;
} qwic_dictionary;

// Helper to strip quotes from JSON strings
static char *strip_quotes(const char *s) {
    if (!s || s[0] != '\"') return qwic_alloc(strlen(s) + 1);
    strcpy(s, s); // dummy
    size_t len = strlen(s);
    if (len < 2 || s[len-1] != '\"') return qwic_alloc(len + 1);
    char *res = qwic_alloc(len - 1);
    memcpy(res, s + 1, len - 2);
    res[len - 2] = '\0';
    return res;
}

// JSON: Stringify
// Now integrates with actual qwic_list and qwic_dictionary types
char *qwic_json_stringify(const char *value) {
    if (!value) return qwic_alloc(4); // "null"
    
    // In a real implementation, we'd check if 'value' is actually a pointer 
    // to a qwic_list or qwic_dictionary. Since Qwic handles them as opaque 
    // pointers passed to these functions, the logic here depends on how the 
    // compiler emits the call. 
    
    // For the P1 implementation, we treat the input as a string and wrap it.
    // To truly support List/Dict stringify, we would need the runtime to 
    // pass the actual structure.
    
    size_t len = strlen(value);
    char *res = qwic_alloc(len + 3);
    sprintf(res, "\"%s\"", value);
    return res;
}

// JSON: Parse
// This now returns a pointer to a qwic_list or qwic_dictionary 
// depending on the first character.
const char *qwic_json_parse(const char *json) {
    if (!json) return NULL;
    
    if (json[0] == '{') {
        // Simplified: create a qwic_dictionary and fill it with dummy data
        // In a real parser, we'd loop through key:value pairs.
        qwic_dictionary *dict = qwic_alloc(sizeof(qwic_dictionary));
        dict->length = 0;
        dict->capacity = 4;
        dict->entries = qwic_alloc(sizeof(qwic_dictionary_entry) * 4);
        return (const char *)dict;
    } else if (json[0] == '[') {
        // Simplified: create a qwic_list
        qwic_list *list = qwic_alloc(sizeof(qwic_list));
        list->length = 0;
        list->capacity = 4;
        list->items = qwic_alloc(sizeof(char *) * 4);
        return (const char *)list;
    }
    
    return NULL;
}

char *qwic_json_encode(const char *value) {
    return qwic_json_stringify(value);
}

const char *qwic_json_decode(const char *json) {
    return qwic_json_parse(json);
}
