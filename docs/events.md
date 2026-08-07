# Transport events

Jissr Bypass can report major routing decisions to application code. Events
are disabled until a listener is attached and the library does not write them
to the system log automatically.

## Event contract

| Type | Meaning |
| --- | --- |
| `native.attempt` | Auto policy is checking the ordinary HTTPS route. |
| `native.selected` | Native is selected, either after success or from cache. |
| `native.failed` | The native request failed. Inspect `safeToFallback`. |
| `fallback.started` | A pre-request native failure activated ECH fallback. |
| `bypass.selected` | Bypass-only policy, cached routing, or fallback selected ECH. |
| `bypass.failed` | The ECH fallback also failed. |
| `resolution.selected` | A fresh mutually consistent address/ECH answer was cached. |
| `network.reset` | Route, resolution, and connection state was reset. |
| `connections.idle_closed` | The application closed pooled idle connections. |

Every event can include:

- `host`: canonical destination hostname;
- `route`: `native` or `bypass`;
- `stage`: `route`, `resolve`, `request`, or `lifecycle`;
- `reason`: a stable machine-readable reason;
- `detail`: a bounded diagnostic error without response data;
- `resolver`: the selected DoH endpoint for a fresh resolution;
- `cached`: whether a cached route decision was used;
- `safeToFallback`: whether the native failure occurred before a request could
  be transmitted; and
- `likelyInterference`: evidence consistent with DNS, TCP, or pre-TLS
  interference.

`likelyInterference` is deliberately not named `blocked`. A timeout, offline
network, firewall, broken router, and deliberate filtering can look identical
from an HTTPS client. Applications should display wording such as “ordinary
connection unavailable; protected route used,” not identify a network
operator without independent evidence.

Events never include URL paths, query strings, request or response headers,
bodies, tokens, cookies, or response payloads. The diagnostic `detail` is
limited to 512 characters and has line breaks removed.

## Android

```kotlin
import jissrbypass.Event
import jissrbypass.EventListener
import jissrbypass.Jissrbypass

private val client = Jissrbypass.newAutoClientWithTimeout(750)

fun installTransportEvents() {
    client.setEventListener(object : EventListener {
        override fun onEvent(event: Event?) {
            event ?: return
            when (event.getType()) {
                "fallback.started" -> {
                    // Update diagnostics or enqueue privacy-reviewed telemetry.
                }
                "bypass.failed" -> {
                    // Record transport availability without request content.
                }
            }
        }
    })
}
```

## Apple platforms

The generated Swift protocol follows the same mobile interface. Implement
`JissrbypassEventListener`, attach it with `setEventListener`, and read the
event getters. Use `JissrbypassNewAutoClientWithTimeout` when the default
750-millisecond budget does not fit the application.

Treat the generated framework headers as the source of truth for Swift
optional values and integer types.

## Go

```go
config := jissrbypass.DefaultConfig()
config.NativeFirst = true
config.NativeAttemptTimeout = 500 * time.Millisecond
config.EventHandler = func(event jissrbypass.Event) {
    if event.Type == jissrbypass.EventFallbackStarted {
        log.Printf(
            "protected route selected host=%s reason=%s likely_interference=%t",
            event.Host,
            event.Reason,
            event.LikelyInterference,
        )
    }
}
client, err := jissrbypass.NewClient(config)
```

The handler is synchronous so related events remain ordered. It must return
quickly and must not make another request with the same client. Send analytics
to an application-owned queue rather than performing network work directly in
the callback. Panics from a Go handler are recovered and cannot break the
transport request.

Passing `nil` to `SetEventHandler` or `setEventListener` disables delivery.
