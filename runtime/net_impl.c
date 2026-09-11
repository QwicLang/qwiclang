#include "qwic_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netdb.h>
#include <errno.h>

// Opaque handle for TCP connections
typedef struct {
    int fd;
} qwic_tcp_conn;

// TCP: Connect
void *qwic_net_tcp_connect(const char *host, int64_t port) {
    struct addrinfo hints, *res;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;

    char port_str[10];
    sprintf(port_str, "%lld", (long long)port);

    if (getaddrinfo(host, port_str, &hints, &res) != 0) return NULL;

    int sock = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
    if (sock == -1) {
        freeaddrinfo(res);
        return NULL;
    }

    if (connect(sock, res->ai_addr, res->ai_addrlen) == -1) {
        close(sock);
        freeaddrinfo(res);
        return NULL;
    }

    freeaddrinfo(res);
    qwic_tcp_conn *conn = qwic_alloc(sizeof(qwic_tcp_conn));
    conn->fd = sock;
    return conn;
}

// TCP: Listen
void *qwic_net_tcp_listen(const char *port_str) {
    struct addrinfo hints, *res;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_flags = AI_PASSIVE;

    if (getaddrinfo(NULL, port_str, &hints, &res) != 0) return NULL;

    int sock = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
    if (sock == -1) {
        freeaddrinfo(res);
        return NULL;
    }

    int opt = 1;
    setsockopt(sock, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    if (bind(sock, res->ai_addr, res->ai_addrlen) == -1) {
        close(sock);
        freeaddrinfo(res);
        return NULL;
    }

    if (listen(sock, 10) == -1) {
        close(sock);
        freeaddrinfo(res);
        return NULL;
    }

    freeaddrinfo(res);
    qwic_tcp_conn *conn = qwic_alloc(sizeof(qwic_tcp_conn));
    conn->fd = sock;
    return conn;
}

// TCP: Accept
void *qwic_net_tcp_accept(void *listen_conn) {
    qwic_tcp_conn *lconn = (qwic_tcp_conn *)listen_conn;
    int client_fd = accept(lconn->fd, NULL, NULL);
    if (client_fd == -1) return NULL;

    qwic_tcp_conn *conn = qwic_alloc(sizeof(qwic_tcp_conn));
    conn->fd = client_fd;
    return conn;
}

// TCP: Read
const char *qwic_net_tcp_read(void *conn, int64_t max_len) {
    qwic_tcp_conn *cconn = (qwic_tcp_conn *)conn;
    char *buf = qwic_alloc(max_len + 1);
    ssize_t n = read(cconn->fd, buf, max_len);
    if (n <= 0) {
        qwic_free(buf);
        return NULL;
    }
    buf[n] = '\0';
    return buf;
}

// TCP: Write
int64_t qwic_net_tcp_write(void *conn, const char *data) {
    qwic_tcp_conn *cconn = (qwic_tcp_conn *)conn;
    ssize_t n = write(cconn->fd, data, strlen(data));
    return (int64_t)n;
}

// TCP: Close
void qwic_net_tcp_close(void *conn) {
    qwic_tcp_conn *cconn = (qwic_tcp_conn *)conn;
    close(cconn->fd);
    qwic_free(cconn);
}

// DNS: Lookup
const char *qwic_net_dns_lookup(const char *host) {
    struct addrinfo hints, *res;
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_INET;
    hints.ai_socktype = SOCK_STREAM;

    if (getaddrinfo(host, NULL, &hints, &res) != 0) return NULL;

    struct sockaddr_in *ipv4 = (struct sockaddr_in *)res->ai_addr;
    char *ip = qwic_alloc(INET_ADDRSTRLEN);
    inet_ntop(AF_INET, &(ipv4->sin_addr), ip, INET_ADDRSTRLEN);
    
    freeaddrinfo(res);
    return ip;
}
