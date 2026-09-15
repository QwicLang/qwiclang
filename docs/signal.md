# Signal Package Compatibility

Qwic's bootstrap compiler supports the language and runtime features needed by
the Signal HTTP package:

- Explicit named method receivers.
- Qualified package types such as `signal.Request`.
- Heterogeneous list and map values.
- Dynamic field access, indexing, calls, and `.type` inspection.
- By-value lambda captures.
- String splitting, path matching, slicing, and integer conversion.
- JSON objects, arrays, strings, numbers, booleans, and null.
- A synchronous HTTP/1.1 listener on Linux, macOS, and Windows.

The executable compatibility fixture is under
`tests/integration/testdata/signal`. It builds the complete server application
and executes Signal's request-body validation path.

## Receiver Syntax

Signal methods must use Qwic's explicit receiver syntax:

```qwic
public func (app: App) routes(routeList: list) {
    app.routes = app.routes + routeList
}
```

Type-qualified declarations without a receiver are static functions:

```qwic
public func App.new(): App {
    return App { routes: [] }
}
```

The compiler intentionally does not provide an implicit `this` binding.

## Current HTTP Limits

`http.listen` currently serves synchronous HTTP/1.1 requests. It supports
request method, path, headers, body text, JSON responses, response status, and
`Content-Type`. TLS, chunked request bodies, streaming, keep-alive, concurrent
request handling, middleware, and graceful shutdown remain future runtime work.

The JSON parser is suitable for ordinary request payloads but does not yet
provide detailed parse diagnostics or full Unicode escape decoding. Signal
validation throws string errors, matching Qwic's current error model.

