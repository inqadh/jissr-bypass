# Client adapters

Jissr Bypass can be installed once at an application's HTTP composition root.
Existing endpoint definitions and request call sites can then continue using
their normal client API.

The router applies Jissr Bypass only to an explicit set of exact hostnames:

```text
configured API host -> Jissr Bypass
every other host    -> existing HTTP transport
```

Exact matching is deliberate. Configuring `example.com` does not silently
include `api.example.com` or other subdomains.

## Compatibility contract

The first adapter release supports ordinary bounded HTTPS requests used by
JSON and similar APIs:

- HTTPS on port 443;
- request and response headers;
- bounded in-memory request and response bodies;
- every HTTP status code; and
- ordinary application authentication.

It does not claim transparent support for:

- WebSockets;
- bidirectional, duplex, or streaming request bodies;
- streaming or very large responses;
- background transfer services; or
- mutating a client that is already being used concurrently.

Create the routed client once, replace the application-level client reference,
and allow old pooled connections to be discarded with the old client.

## Go

Clone an existing `http.Client` without changing its callers:

```go
base := &http.Client{}
routed, err := jissrbypass.NewHostRoutingHTTPClient(
    base,
    jissrbypass.NewDefaultClient(),
    "api.example.com",
)
if err != nil {
    log.Fatal(err)
}
```

The returned client preserves the base client's timeout, cookie jar, redirect
policy, and fallback behavior. The original client is not mutated.

## Android

The optional OkHttp adapter is intended to be installed as an application
interceptor on the central `OkHttpClient.Builder`. Retrofit can continue using
that OkHttp client without changing service interfaces.

The adapter buffers request and response bodies. Cancellation of the outer
OkHttp call does not yet guarantee immediate interruption of native work.

## Apple platforms

The optional URL loading adapter is installed on a dedicated
`URLSessionConfiguration` and used to recreate the application's central
session. Existing `URLRequest` construction remains unchanged.

Global `URLProtocol` registration, background sessions, WebSockets, and
`WKWebView` interception are outside the supported contract.
