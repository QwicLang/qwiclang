# QwicLang v0 — Standard Package Library Plan

## 1. Goal

The QwicLang standard library should provide a small, fast, predictable set of packages covering the functionality required to build:

* CLI applications
* Web APIs and HTTP services
* Network services
* Game systems
* File-processing tools
* Concurrent applications
* Secure applications
* Backend services
* Systems utilities

The initial library should **not try to be huge**.

The priority is:

> **Small API surface. Strong primitives. Excellent performance. Easy to learn.**

The first release should focus on eight critical packages:

1. `json`
2. `http`
3. `crypto`
4. `strings`
5. `net`
6. `sync`
7. `fs`
8. `time`

---

# 2. Implementation Priority

The packages should be developed in dependency order rather than simply the order they appear above.

| Priority | Package   | Importance               | Depends On            |
| -------- | --------- | ------------------------ | --------------------- |
| P0       | `strings` | Core language utility    | Runtime               |
| P0       | `time`    | Fundamental primitive    | Runtime               |
| P0       | `fs`      | File and async I/O       | Runtime               |
| P0       | `sync`    | Concurrency primitives   | Runtime               |
| P1       | `net`     | Networking foundation    | `sync`, `time`        |
| P1       | `json`    | Data serialization       | `strings`             |
| P1       | `crypto`  | Security primitives      | `strings`, `fs`       |
| P2       | `http`    | Web/network applications | `net`, `json`, `time` |

The recommended implementation sequence is therefore:

```text
Runtime
   │
   ├── strings
   ├── time
   ├── fs
   └── sync
          │
          └── net
                 │
                 ├── json
                 │
                 └── crypto
                        │
                        └── http
```

`http` should be one of the final packages in the first standard-library milestone because it sits on top of several lower-level capabilities.

---

# 3. Current v0 Implementation Status

This plan is intentionally broader than the compiler can support today. The
current implementation follows the dependency order, but only exposes APIs that
can compile, type check, lower to IR, generate native code, and run end-to-end.

## Implemented Now

### `strings`

The first standard package is implemented as a compiler-known package backed by
runtime C functions. It requires an explicit import:

```qwic
import strings

public func main() {
    const clean = strings.trim("  QwicLang  ")
    print(strings.upper(clean))
}
```

Implemented API:

```text
strings.length(value: string): int
strings.empty(value: string): bool
strings.trim(value: string): string
strings.trimLeft(value: string): string
strings.trimRight(value: string): string
strings.upper(value: string): string
strings.lower(value: string): string
strings.contains(value: string, needle: string): bool
strings.startsWith(value: string, prefix: string): bool
strings.endsWith(value: string, suffix: string): bool
strings.indexOf(value: string, needle: string): int
```

Current `strings` limitations:

* `upper` and `lower` are documented as ASCII-oriented v0 behavior.
* Returning string functions allocate through the runtime; richer ownership and
  lifetime semantics are not designed yet.
* APIs that require arrays, bytes, or richer slicing are still deferred.

### Bootstrap data structures

The `data-structure` branch adds the first runtime-backed collection packages:

```text
lists
sets
dictionaries
tuples
```

Implemented API:

```text
lists.new(): list
lists.push(list: list, value: string): void
lists.get(list: list, index: int): string
lists.length(list: list): int
lists.contains(list: list, value: string): bool

sets.new(): set
sets.add(set: set, value: string): void
sets.contains(set: set, value: string): bool
sets.length(set: set): int

dictionaries.new(): dictionary
dictionaries.set(dictionary: dictionary, key: string, value: string): void
dictionaries.get(dictionary: dictionary, key: string): string
dictionaries.contains(dictionary: dictionary, key: string): bool
dictionaries.length(dictionary: dictionary): int

tuples.new2(first: string, second: string): tuple
tuples.first(tuple: tuple): string
tuples.second(tuple: tuple): string
tuples.length(tuple: tuple): int
```

Current data-structure limitations:

* Collections store strings only. Generic element types require a real generic
  type system or a designed `any` container model.
* Dictionaries use string keys and string values only.
* Tuples currently model fixed pairs only. Literal tuple syntax and typed arity
  are deferred.
* Missing keys and out-of-range indexes return an empty string until Qwic has a
  standard result/error model.
* The runtime uses simple linear storage. Hash tables and more advanced storage
  policies are deferred until correctness and API shape are settled.

## Not Yet Implemented

The remaining packages are deliberately not exposed yet. Adding stubs would make
programs type check without actually working, which violates the project rule
against fake features.

| Package | Status | Reason |
| ------- | ------ | ------ |
| `time` | Planned | Needs a documented `nano`/timestamp representation, platform clock runtime APIs, and a settled return convention for parsed/formatted time values. |
| `fs` | Planned | Needs a result/error model before file failures can be represented cleanly, plus bytes/arrays or another real data-buffer type for file contents. Async variants also require `await`. |
| `sync` | Planned | Needs the concurrency runtime, method/member call support, heap-backed values, and type representations for mutexes, atomics, wait groups, and channels. |
| `net` | Planned | Depends on `sync` and `time`, and needs resource/socket handles, byte buffers, errors, timeouts, and eventually async I/O. |
| `json` | Planned | Needs arrays, objects/structs/maps, field selection on values, and a consistent parse error/result type. It can build on `strings` later. |
| `crypto` | Planned | Must use audited platform/library primitives rather than homemade algorithms, and needs bytes plus explicit error handling for decoding/randomness failures. |
| `http` | Planned | Depends on `net`, `json`, and `time`; also needs request/response structs, handlers/callbacks, routing state, errors, and concurrency support. |

The next practical standard-library step is to finish more of `strings`
(`substring`, `slice`, `replace`, conversions) only after the compiler has the
required value types and runtime ownership story. The next package should be
`time` once the runtime exposes clock primitives and the numeric representation
is documented.

---

# 4. Package Design Principles

Every standard package should follow the same principles.

### 4.1 Small public API

Do not expose internal implementation details.

For example:

```qwic
import strings

const result = strings.trim("  hello  ")
```

The user should not need to know how trimming is implemented.

---

### 4.2 Explicit imports

Packages must be imported explicitly:

```qwic
import json
import http
import crypto
```

No global namespace pollution.

---

### 4.3 Public API vs private implementation

A package can contain private functions and data structures.

Example:

```qwic
public func parse(data)
private func parseObject(data)
private func parseArray(data)
```

Only `parse` is available to consumers.

---

### 4.4 Predictable errors

Packages should use a consistent error model.

For example:

```qwic
const result = fs.read("config.json")

if result.error {
    print(result.error)
}
```

The exact QwicLang error/result mechanism should be standardized before the package APIs are finalized.

Do not create eight different error-handling styles across the standard library.

---

### 4.5 Synchronous and asynchronous APIs

I/O-heavy packages should provide both where appropriate.

Example:

```qwic
const data = fs.read("file.txt")
```

and:

```qwic
const data = await fs.readAsync("file.txt")
```

Async APIs should avoid blocking the main execution thread.

---

# 5. `strings`

## Priority

**P0 — Foundational**

Strings will be used throughout practically every QwicLang application.

## Purpose

Provide efficient string manipulation without requiring developers to implement common operations themselves.

## Initial API

```text
strings.length()
strings.empty()
strings.trim()
strings.trimLeft()
strings.trimRight()

strings.upper()
strings.lower()

strings.contains()
strings.startsWith()
strings.endsWith()

strings.indexOf()
strings.lastIndexOf()

strings.substring()
strings.slice()

strings.replace()
strings.replaceAll()

strings.split()
strings.join()

strings.repeat()

strings.compare()
strings.equal()

strings.toInt()
strings.toFloat()

strings.fromBytes()
strings.toBytes()
```

## Example

```qwic
import strings

const name = "  QwicLang  "

const clean = strings.trim(name)
const upper = strings.upper(clean)

print(upper)
```

Result:

```text
QWICLANG
```

## Performance goals

* Avoid unnecessary allocations.
* Provide efficient byte/string conversion.
* Make UTF-8 handling explicit.
* Keep ASCII operations extremely fast.
* Avoid hidden expensive transformations.

---

# 6. `time`

## Priority

**P0 — Foundational**

Time is required by networking, HTTP, games, logging, databases and synchronization.

## Purpose

Provide clocks, timestamps, date/time manipulation and timing utilities.

## Initial API

```text
time.now()
time.utcNow()

time.timestamp()
time.unix()

time.year()
time.month()
time.day()

time.hour()
time.minute()
time.second()

time.sleep()
time.sleepAsync()

time.since()

time.before()
time.after()
time.equal()

time.format()
time.parse()
```

## Example

```qwic
import time

const start = time.now()

doWork()

const elapsed = time.since(start)

print(elapsed)
```

### `nano`

The `nano` type should integrate with the timing system.

Example:

```qwic
const frameTime: nano = 0.016666
```

The value remains a plain numeric value.

QwicLang should **not** require unit suffixes such as:

```text
16ms
100ms
1s
```

Timing APIs should document the expected numeric representation instead.

---

# 7. `fs`

## Priority

**P0 — Foundational**

## Purpose

Provide filesystem operations and asynchronous file I/O.

## Initial API

### Files

```text
fs.read()
fs.readAsync()

fs.write()
fs.writeAsync()

fs.append()
fs.appendAsync()

fs.exists()

fs.copy()
fs.copyAsync()

fs.move()
fs.moveAsync()

fs.remove()
fs.removeAsync()
```

### Directories

```text
fs.mkdir()
fs.mkdirAsync()

fs.rmdir()
fs.rmdirAsync()

fs.list()
fs.listAsync()
```

### Metadata

```text
fs.stat()
fs.statAsync()

fs.size()
fs.modified()
fs.isFile()
fs.isDirectory()
```

### Paths

```text
fs.path.join()
fs.path.basename()
fs.path.dirname()
fs.path.extension()
fs.path.absolute()
```

## Example

```qwic
import fs

const config = await fs.readAsync("config.json")

print(config)
```

## Design requirement

Filesystem APIs should support both:

```qwic
fs.read()
```

and:

```qwic
await fs.readAsync()
```

The async implementation should integrate with QwicLang's future/event system rather than creating unnecessary threads for every operation.

---

# 8. `sync`

## Priority

**P0 — Foundational**

## Purpose

Provide synchronization primitives required for concurrent applications.

## Initial API

```text
sync.Mutex
sync.RWMutex
sync.Atomic
sync.WaitGroup
sync.Once
sync.Channel
```

### Mutex

```qwic
import sync

const lock = sync.Mutex()

lock.lock()

// critical section

lock.unlock()
```

### RWMutex

```qwic
const lock = sync.RWMutex()

lock.readLock()
lock.readUnlock()

lock.writeLock()
lock.writeUnlock()
```

### Atomic

```qwic
const counter = sync.Atomic(0)

counter.increment()
counter.decrement()

const value = counter.load()
counter.store(100)
```

### WaitGroup

```qwic
const workers = sync.WaitGroup()

workers.add(2)

// workers...

workers.wait()
```

### Channel

Channels can provide communication between concurrent tasks:

```qwic
const messages = sync.Channel(10)

messages.send("hello")

const message = messages.receive()
```

## Design goal

Concurrency should be safe by default.

The package should provide primitives without forcing developers to understand low-level OS synchronization mechanisms.

---

# 9. `net`

## Priority

**P1 — Core infrastructure**

## Purpose

Provide low-level networking primitives.

## Initial API

### Addresses

```text
net.IP
net.Address
net.Port
```

### TCP

```text
net.tcp.connect()
net.tcp.listen()
net.tcp.accept()
net.tcp.read()
net.tcp.write()
net.tcp.close()
```

### UDP

```text
net.udp.bind()
net.udp.send()
net.udp.receive()
net.udp.close()
```

### DNS

```text
net.dns.lookup()
net.dns.resolve()
```

## Example TCP server

```qwic
import net

const server = net.tcp.listen(":9000")

while true {
    const connection = server.accept()

    connection.write("Hello from QwicLang")
    connection.close()
}
```

## Example client

```qwic
import net

const connection = net.tcp.connect("example.com", 9000)

connection.write("Hello")

const response = connection.read()

print(response)

connection.close()
```

## Design goals

* Efficient socket abstraction.
* IPv4 and IPv6.
* TCP and UDP.
* DNS.
* Timeouts.
* Non-blocking operations.
* Integration with async execution.
* Minimal abstraction over underlying networking primitives.

---

# 10. `json`

## Priority

**P1 — Essential application package**

JSON will be heavily used by HTTP APIs, configuration systems and backend applications.

## Purpose

Provide fast JSON encoding and decoding.

## Initial API

```text
json.parse()
json.parseAsync()

json.stringify()
json.stringifyPretty()

json.encode()
json.decode()
```

## Example

```qwic
import json

const data = json.parse("""
{
    "name": "QwicLang",
    "version": "0.1"
}
""")

print(data.name)
```

Encoding:

```qwic
const user = {
    name: "Tanaka",
    age: 30
}

const output = json.stringify(user)

print(output)
```

## Requirements

The JSON implementation should support:

```text
string
number
boolean
null
array
object
```

It should also provide clear errors for malformed JSON.

## Performance goal

JSON should be optimized for API/server workloads.

Avoid unnecessary conversion between:

```text
bytes → string → parser buffer → object
```

where direct byte parsing can be used.

---

# 11. `crypto`

## Priority

**P1 — Security-critical**

## Purpose

Provide secure cryptographic primitives.

QwicLang should **not implement cryptographic algorithms from scratch**.

The implementation should use audited, mature cryptographic implementations underneath the QwicLang API.

## Initial API

### Hashing

```text
crypto.sha256()
crypto.sha512()
crypto.sha3()
```

### HMAC

```text
crypto.hmac()
```

### Randomness

```text
crypto.randomBytes()
crypto.randomInt()
```

Cryptographic randomness must come from the operating system's secure random source.

### Encoding

```text
crypto.hexEncode()
crypto.hexDecode()

crypto.base64Encode()
crypto.base64Decode()
```

### Password hashing

A future password package/API should use an intentionally slow password-hashing algorithm rather than generic SHA hashing.

Possible implementations:

```text
crypto.argon2()
crypto.scrypt()
```

### Encryption

Future versions can expose authenticated encryption such as:

```text
crypto.aesGCM()
crypto.chacha20Poly1305()
```

## Security requirements

The package must:

* avoid insecure random generators for security operations
* avoid silently insecure defaults
* zero sensitive buffers where practical
* provide authenticated encryption
* expose secure hashing APIs
* document security guarantees
* rely on established cryptographic implementations

---

# 12. `http`

## Priority

**P2 — Major application package**

`http` should sit on top of `net`.

It will likely become one of the most-used QwicLang packages.

## Purpose

Provide HTTP client and server functionality.

## HTTP client

```text
http.get()
http.post()
http.put()
http.patch()
http.delete()
http.request()
```

Example:

```qwic
import http

const response = http.get("https://example.com")

print(response.status)
print(response.body)
```

POST:

```qwic
const response = http.post(
    "https://api.example.com/users",
    {
        name: "QwicLang"
    }
)
```

## HTTP request

Provide:

```text
http.Request
```

with:

```text
method
url
headers
body
query
```

## HTTP response

Provide:

```text
http.Response
```

with:

```text
status
headers
body
```

---

# 13. HTTP Server

QwicLang should make HTTP servers extremely simple.

```qwic
import http

public func main() {
    const server = http.server(":8080")

    server.get("/", home)

    server.listen()
}

func home(request: http.Request, response: http.Response) {
    response.text("Hello from QwicLang")
}
```

JSON API:

```qwic
func users(request, response) {
    response.json({
        users: []
    })
}
```

Routing should support:

```text
GET
POST
PUT
PATCH
DELETE
OPTIONS
HEAD
```

and eventually:

```text
/users/:id
```

style parameters.

---

# 14. HTTP Features

The initial implementation should support:

### Client

* GET
* POST
* PUT
* PATCH
* DELETE
* custom headers
* query parameters
* request body
* JSON
* timeouts
* redirects
* TLS/HTTPS

### Server

* routing
* request parsing
* response handling
* headers
* cookies
* JSON responses
* middleware
* graceful shutdown
* concurrent requests

### Future

```text
WebSocket
HTTP/2
HTTP/3
multipart/form-data
streaming
server-sent events
```

These should not block the initial HTTP implementation.

---

# 15. Package Dependency Rules

The standard library should maintain strict dependency boundaries.

Recommended:

```text
strings
   ↓
time
   ↓
fs
   ↓
sync
   ↓
net
   ↓
json
   ↓
crypto
   ↓
http
```

However, these should not become a strict one-way chain internally.

For example, `strings` should **not** import `http`.

The rule should be:

> Lower-level packages must never depend on higher-level application packages.

Therefore:

```text
strings  → runtime
time     → runtime
fs       → runtime
sync     → runtime
net      → runtime + sync/time
json     → runtime + strings
crypto   → runtime
http     → net + json + time
```

This keeps the standard library maintainable.

---

# 16. Standard Package Layout

The repository should separate package implementations cleanly.

```text
stdlib/
├── strings/
│   ├── strings.qw
│   └── strings_test.qw
│
├── time/
│   ├── time.qw
│   └── time_test.qw
│
├── fs/
│   ├── fs.qw
│   └── fs_test.qw
│
├── sync/
│   ├── sync.qw
│   └── sync_test.qw
│
├── net/
│   ├── net.qw
│   ├── tcp.qw
│   ├── udp.qw
│   └── net_test.qw
│
├── json/
│   ├── json.qw
│   ├── parser.qw
│   ├── encoder.qw
│   └── json_test.qw
│
├── crypto/
│   ├── crypto.qw
│   └── crypto_test.qw
│
└── http/
    ├── http.qw
    ├── client.qw
    ├── server.qw
    ├── router.qw
    └── http_test.qw
```

Low-level implementations that require native code should live behind the package API.

For example:

```text
stdlib/net/
        │
        ├── QwicLang API
        │
        └── native/
              ├── tcp
              ├── udp
              └── dns
```

Users should not need to interact with those internals.

---

# 17. Testing Strategy

Every package must have tests before being considered complete.

### Unit tests

Test individual functions:

```text
strings.trim()
json.parse()
time.format()
```

### Integration tests

Test packages working together:

```text
net + http
json + http
fs + json
crypto + http
sync + net
```

### Stress tests

Especially important for:

```text
net
http
sync
fs
json
```

### Security tests

Especially:

```text
crypto
http
json
```

Test malformed input, invalid encodings, oversized payloads and unexpected network data.

---

# 18. Performance Benchmarks

QwicLang's standard library should be benchmarked from the beginning.

Important benchmarks:

```text
JSON parsing
JSON encoding
string operations
file read/write
TCP throughput
HTTP requests/sec
HTTP latency
mutex contention
channel throughput
crypto hashing
```

The benchmark suite should become part of the language's regression testing.

A standard-library optimization should not be accepted merely because it "looks faster."

Measure it.

---

# 19. v0 Milestones

## Milestone 1 — Foundation

Implement:

```text
strings
time
fs
sync
```

Deliverable:

Basic QwicLang applications can manipulate data, access files and perform concurrency/timing operations.

---

## Milestone 2 — Networking

Implement:

```text
net
```

Deliverable:

QwicLang can create TCP/UDP clients and servers.

Example:

```qwic
import net

public func main() {
    const server = net.tcp.listen(":8080")
    // accept connections
}
```

---

## Milestone 3 — Data

Implement:

```text
json
```

Deliverable:

QwicLang can parse and generate JSON efficiently.

---

## Milestone 4 — Security

Implement:

```text
crypto
```

Deliverable:

QwicLang applications have access to secure hashing, randomness, encoding and cryptographic primitives.

---

## Milestone 5 — HTTP

Implement:

```text
http
```

Deliverable:

A complete HTTP client and basic production-capable HTTP server.

Example:

```qwic
import http

public func main() {
    const server = http.server(":8080")

    server.get("/", home)

    server.listen()
}

func home(request, response) {
    response.text("Hello QwicLang")
}
```

---

# 20. Definition of Done

A package is not considered complete merely because its functions compile.

Each package must have:

* public API documentation
* unit tests
* integration tests where applicable
* error handling
* benchmarks where performance matters
* examples
* platform considerations
* memory/allocation review
* concurrency review where applicable
* security review where applicable

The package must also follow QwicLang's naming, visibility and error-handling conventions.

---

# 21. v0 Standard Library Target

The first QwicLang release should aim to make this possible:

```qwic
import http
import json
import fs
import crypto
import strings
import time
import sync
import net

public func main() {
    const config = await fs.readAsync("config.json")

    const data = json.parse(config)

    const token = crypto.randomBytes(32)

    const name = strings.trim(data.name)

    print(name)

    const server = http.server(":8080")

    server.get("/", home)

    server.listen()
}

func home(request, response) {
    response.json({
        message: "Hello from QwicLang",
        time: time.now()
    })
}
```

If QwicLang can compile and execute applications like this reliably, the standard library has achieved its v0 objective.

---

# 22. What Comes After These Eight

Do **not** add dozens of packages before these eight are mature.

After v0, the next candidates should be:

```text
os
process
log
math
regex
url
encoding
database
testing
cli
compress
crypto/tls
```

Eventually:

```text
websocket
smtp
email
xml
yaml
sqlite
grpc
template
image
```

The priority should remain:

> **Mature the core before expanding the library.**

QwicLang should have a small standard library that developers can trust rather than a huge standard library full of inconsistent APIs.

# Final v0 Priority

```text
P0  strings
P0  time
P0  fs
P0  sync

P1  net
P1  json
P1  crypto

P2  http
```

**The critical v0 target is therefore:**

```text
QwicLang Runtime
       ↓
strings + time + fs + sync
       ↓
      net
       ↓
json + crypto
       ↓
      http
       ↓
Production-capable QwicLang applications
```

This gives QwicLang a strong foundation for its intended use cases: **fast backend services, networking, systems tooling and game development**, without bloating the first release.
