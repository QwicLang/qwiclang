#include "qwic_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

// Simple JSON-like parser for v0. 
// In a production language, we would use a robust library like cJSON or JSMN.
// For v0, we implement basic stringification and a dummy parser that 
// identifies if a string is valid JSON.

// JSON: Stringify
// In a real implementation, this would traverse Qwic's internal Dict/List types.
// For now, since Qwic's runtime types are opaque, we provide a basic 
// utility to wrap a string in quotes (simulating a simple JSON string).
char *qwic_json_stringify(const char *value) {
    if (!value) return qwic_alloc(4); // "null"
    
    size_t len = strlen(value);
    char *res = qwic_alloc(len + 3); // quotes + null terminator + 1
    sprintf(res, "\"%s\"", value);
    return res;
}

// JSON: Parse
// For v0, this is a placeholder that returns the input if it starts 
// with { or [ (simulating "parsing" by returning the raw string as the 
// internal representation).
const char *qwic_json_parse(const char *json) {
    if (!json) return NULL;
    
    // Basic validation: must start with { or [
    if (json[0] != '{' && json[0] != '[') {
        return NULL; 
    }
    
    return json; // In v0, we return the string as the "parsed" object
}

// JSON: Encode
// Alias for stringify in v0
char *qwic_json_encode(const char *value) {
    return qwic_json_stringify(value);
}

// JSON: Decode
// Alias for parse in v0
const char *qwic_json_decode(const char *json) {
    return qwic_json_parse(json);
}
