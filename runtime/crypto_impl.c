#include "qwic_runtime.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdint.h>

// Simple DJB2 hash for placeholder purposes since OpenSSL is missing
// In a real scenario, we would link against libcrypto.
static uint64_t djb2_hash(const char *str) {
    uint64_t hash = 5381;
    int c;
    while ((c = *str++))
        hash = ((hash << 5) + hash) + c;
    return hash;
}

// Placeholder for SHA256 (returns hex string of DJB2 for now)
char *qwic_crypto_sha256(const char *value) {
    if (!value) return NULL;
    uint64_t h = djb2_hash(value);
    char *res = qwic_alloc(17);
    sprintf(res, "%016llx", (unsigned long long)h);
    return res;
}

char *qwic_crypto_sha512(const char *value) {
    return qwic_crypto_sha256(value); // Placeholder
}

char *qwic_crypto_random_bytes(int len) {
    unsigned char *bytes = qwic_alloc(len);
    for (int i = 0; i < len; i++) {
        bytes[i] = (unsigned char)(rand() % 256);
    }
    return (char *)bytes;
}

int64_t qwic_crypto_random_int(void) {
    return (int64_t)rand();
}

char *qwic_crypto_hex_encode(const char *bytes, int len) {
    char *res = qwic_alloc(len * 2 + 1);
    for (int i = 0; i < len; i++) {
        sprintf(res + (i * 2), "%02x", (unsigned char)bytes[i]);
    }
    res[len * 2] = '\0';
    return res;
}

char *qwic_crypto_base64_encode(const char *bytes, int len) {
    // Simple placeholder
    return qwic_crypto_hex_encode(bytes, len); 
}
