#include "qwic_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netdb.h>

// Opaque handles for HTTP
typedef struct {
    char *url;
    char *method;
    char *body;
    char *headers;
} qwic_http_request;

typedef struct {
    int64_t status_code;
    char *body;
    char *headers;
} qwic_http_response;

// HTTP: Simple GET request (Synchronous)
void *qwic_http_get(const char *url) {
    // v0 implementation: simple TCP socket to port 80
    char host[256];
    int port = 80;
    
    // Extremely simple URL parsing (http://host/path)
    const char *start = url;
    if (strncmp(url, "http://", 7) == 0) start += 7;
    
    const char *slash = strchr(start, '/');
    if (slash) {
        strncpy(host, start, slash - start);
        host[slash - start] = '\0';
    } else {
        strcpy(host, start);
    }

    void *conn = qwic_net_tcp_connect(host, port);
    if (!conn) return NULL;

    char request[1024];
    sprintf(request, "GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", 
            slash ? slash : "/", host);
    
    qwic_net_tcp_write(conn, request);
    
    // Read response body (simplified: read everything)
    char *response_data = qwic_alloc(8192);
    int64_t total_read = 0;
    while (total_read < 8191) {
        const char *chunk = qwic_net_tcp_read(conn, 1024);
        if (!chunk) break;
        size_t len = strlen(chunk);
        memcpy(response_data + total_read, chunk, len);
        total_read += len;
        qwic_free((void*)chunk);
    }
    response_data[total_read] = '\0';
    
    qwic_net_tcp_close(conn);

    qwic_http_response *res = qwic_alloc(sizeof(qwic_http_response));
    res->status_code = 200; // Placeholder
    res->body = response_data;
    res->headers = NULL;
    return res;
}

// HTTP: Post
void *qwic_http_post(const char *url, const char *body) {
    // Simplified version of GET with body
    char host[256];
    const char *start = url;
    if (strncmp(url, "http://", 7) == 0) start += 7;
    const char *slash = strchr(start, '/');
    if (slash) {
        strncpy(host, start, slash - start);
        host[slash - start] = '\0';
    } else {
        strcpy(host, start);
    }

    void *conn = qwic_net_tcp_connect(host, 80);
    if (!conn) return NULL;

    char request[2048];
    sprintf(request, "POST %s HTTP/1.1\r\nHost: %s\r\nContent-Length: %ld\r\nConnection: close\r\n\r\n%s", 
            slash ? slash : "/", host, strlen(body), body);
    
    qwic_net_tcp_write(conn, request);
    
    char *response_data = qwic_alloc(8192);
    int64_t total_read = 0;
    while (total_read < 8191) {
        const char *chunk = qwic_net_tcp_read(conn, 1024);
        if (!chunk) break;
        size_t len = strlen(chunk);
        memcpy(response_data + total_read, chunk, len);
        total_read += len;
        qwic_free((void*)chunk);
    }
    response_data[total_read] = '\0';
    
    qwic_net_tcp_close(conn);

    qwic_http_response *res = qwic_alloc(sizeof(qwic_http_response));
    res->status_code = 200;
    res->body = response_data;
    res->headers = NULL;
    return res;
}

// Response helpers
int64_t qwic_http_get_status(void *response) {
    return ((qwic_http_response *)response)->status_code;
}

const char *qwic_http_get_body(void *response) {
    return ((qwic_http_response *)response)->body;
}

void qwic_http_free_response(void *response) {
    qwic_http_response *res = (qwic_http_response *)response;
    qwic_free(res->body);
    qwic_free(res->headers);
    qwic_free(res);
}
