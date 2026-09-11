#include "qwic_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netdb.h>
#include <openssl/ssl.h>
#include <openssl/err.h>

typedef struct {
    char *name;
    char *value;
} qwic_cookie;

typedef struct {
    qwic_cookie *cookies;
    int count;
    int capacity;
} qwic_cookie_jar;

typedef struct {
    char *url;
    char *method;
    char *body;
    char *headers;
    qwic_cookie_jar *jar;
    bool secure;
} qwic_http_request;

typedef struct {
    int64_t status_code;
    char *body;
    char *headers;
    qwic_cookie_jar *jar;
} qwic_http_response;

// --- Cookie Jar Helpers ---

qwic_cookie_jar *qwic_cookie_jar_new() {
    qwic_cookie_jar *jar = qwic_alloc(sizeof(qwic_cookie_jar));
    jar->count = 0;
    jar->capacity = 10;
    jar->cookies = qwic_alloc(sizeof(qwic_cookie) * jar->capacity);
    return jar;
}

void qwic_cookie_jar_add(qwic_cookie_jar *jar, const char *name, const char *value) {
    if (jar->count >= jar->capacity) {
        jar->capacity *= 2;
        qwic_cookie *new_cookies = qwic_alloc(sizeof(qwic_cookie) * jar->capacity);
        memcpy(new_cookies, jar->cookies, sizeof(qwic_cookie) * jar->count);
        qwic_free(jar->cookies);
        jar->cookies = new_cookies;
    }
    jar->cookies[jar->count].name = qwic_alloc(strlen(name) + 1);
    strcpy(jar->cookies[jar->count].name, name);
    jar->cookies[jar->count].value = qwic_alloc(strlen(value) + 1);
    strcpy(jar->cookies[jar->count].value, value);
    jar->count++;
}

char *qwic_cookie_jar_get(qwic_cookie_jar *jar, const char *name) {
    for (int i = 0; i < jar->count; i++) {
        if (strcmp(jar->cookies[i].name, name) == 0) return jar->cookies[i].value;
    }
    return NULL;
}

// --- HTTP Core ---

static void build_request_string(qwic_http_request *req, char *out_buf) {
    char host[256];
    const char *start = req->url;
    if (strncmp(req->url, "http://", 7) == 0) start += 7;
    else if (strncmp(req->url, "https://", 8) == 0) start += 8;
    
    const char *slash = strchr(start, '/');
    if (slash) {
        strncpy(host, start, slash - start);
        host[slash - start] = '\0';
    } else {
        strcpy(host, start);
    }

    sprintf(out_buf, "%s %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n", 
            req->method, slash ? slash : "/", host);

    if (req->jar) {
        strcat(out_buf, "Cookie: ");
        for (int i = 0; i < req->jar->count; i++) {
            strcat(out_buf, req->jar->cookies[i].name);
            strcat(out_buf, "=");
            strcat(out_buf, req->jar->cookies[i].value);
            if (i < req->jar->count - 1) strcat(out_buf, "; ");
        }
        strcat(out_buf, "\r\n");
    }

    if (req->body) {
        char len_hdr[64];
        sprintf(len_hdr, "Content-Length: %ld\r\n", strlen(req->body));
        strcat(out_buf, len_hdr);
    }

    strcat(out_buf, "\r\n");
    if (req->body) {
        strcat(out_buf, req->body);
    }
}

void *qwic_http_send(qwic_http_request *req) {
    char host[256];
    const char *start = req->url;
    if (strncmp(req->url, "http://", 7) == 0) start += 7;
    else if (strncmp(req->url, "https://", 8) == 0) start += 8;
    const char *slash = strchr(start, '/');
    if (slash) {
        strncpy(host, start, slash - start);
        host[slash - start] = '\0';
    } else {
        strcpy(host, start);
    }

    int port = req->secure ? 443 : 80;
    void *conn = qwic_net_tcp_connect(host, port);
    if (!conn) return NULL;

    typedef struct { int fd; } qwic_tcp_conn;
    int fd = ((qwic_tcp_conn*)conn)->fd;

    SSL_CTX *ctx = NULL;
    SSL *ssl = NULL;
    if (req->secure) {
        SSL_library_init();
        OpenSSL_add_all_algorithms();
        SSL_load_error_strings();
        ctx = SSL_CTX_new(TLS_client_method());
        ssl = SSL_new(ctx);
        SSL_set_fd(ssl, fd);
        if (SSL_connect(ssl) <= 0) {
            SSL_free(ssl);
            SSL_CTX_free(ctx);
            qwic_net_tcp_close(conn);
            return NULL;
        }
    }

    char req_str[4096];
    build_request_string(req, req_str);

    if (req->secure) {
        SSL_write(ssl, req_str, strlen(req_str));
    } else {
        write(fd, req_str, strlen(req_str));
    }

    char *res_body = qwic_alloc(16384);
    int total_read = 0;
    char buf[1024];
    while (total_read < 16383) {
        int n = req->secure ? SSL_read(ssl, buf, 1023) : read(fd, buf, 1023);
        if (n <= 0) break;
        memcpy(res_body + total_read, buf, n);
        total_read += n;
    }
    res_body[total_read] = '\0';

    if (req->secure) {
        SSL_shutdown(ssl);
        SSL_free(ssl);
        SSL_CTX_free(ctx);
    }
    qwic_net_tcp_close(conn);

    qwic_http_response *res = qwic_alloc(sizeof(qwic_http_response));
    res->status_code = 200; 
    res->body = res_body;
    res->headers = NULL;
    res->jar = req->jar; 
    return res;
}

void *qwic_http_request_new(const char *url) {
    qwic_http_request *req = qwic_alloc(sizeof(qwic_http_request));
    req->url = qwic_alloc(strlen(url) + 1);
    strcpy(req->url, url);
    req->method = qwic_alloc(8);
    strcpy(req->method, "GET");
    req->body = NULL;
    req->headers = NULL;
    req->jar = qwic_cookie_jar_new();
    req->secure = (strncmp(url, "https://", 8) == 0);
    return req;
}

void qwic_http_request_set_method(void *req, const char *method) {
    qwic_http_request *r = (qwic_http_request *)req;
    qwic_free(r->method);
    r->method = qwic_alloc(strlen(method) + 1);
    strcpy(r->method, method);
}

void qwic_http_request_set_body(void *req, const char *body) {
    qwic_http_request *r = (qwic_http_request *)req;
    if (r->body) qwic_free(r->body);
    r->body = qwic_alloc(strlen(body) + 1);
    strcpy(r->body, body);
}

void qwic_http_request_free(void *req) {
    qwic_http_request *r = (qwic_http_request *)req;
    qwic_free(r->url);
    qwic_free(r->method);
    if (r->body) qwic_free(r->body);
    // jar is usually shared with response or managed separately
    qwic_free(r);
}

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
