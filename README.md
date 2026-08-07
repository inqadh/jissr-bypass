# Jissr Bypass

Jissr Bypass is a native HTTPS transport for networks where ordinary DNS or a
plaintext TLS ClientHello is unreliable. It resolves both address and HTTPS
records over DNS-over-HTTPS and performs TLS 1.3 with Encrypted Client Hello
(ECH).

The same Go core builds for Android, iOS, macOS, Windows, and Linux.

> Status: stable `0.1.x`. Releases follow Semantic Versioning; review upgrade
> notes while the public API is still below `v1.0.0`.

For server setup and complete application examples, follow the
[A-to-Z integration guide](docs/integration-guide.md).

## Go and desktop

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    bypass "github.com/inqadh/jissr-bypass"
)

func main() {
    client := bypass.NewAutoClient()
    request, err := http.NewRequestWithContext(
        context.Background(),
        http.MethodGet,
        "https://api.example.com/v1/profile",
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    request.Header.Set("Authorization", "Bearer <access-token>")
    response, err := client.Do(request)
    if err != nil {
        log.Fatal(err)
    }
    defer response.Body.Close()
    fmt.Printf("HTTP %d\n", response.StatusCode)
}
```

`NewAutoClient` tries the ordinary HTTPS route once, falls back only when the
failure occurs before a request can be sent, and remembers the working route
per host. Use `NewDefaultClient` when every request must use ECH without first
trying the ordinary route. Both clients pool HTTPS connections and cache
DNS/ECH answers; create one long-lived client rather than one per request.

The diagnostic CLI uses the same core:

```sh
go run ./cmd/jissr-bypass
go run ./cmd/jissr-bypass -probe blocked.yusef.ren
```

## Android

Add Maven Central to the application's repositories (Android projects include
it by default), then add the package:

```kotlin
dependencies {
    implementation("org.jissr:jissr-bypass-android:<version>")
}
```

The library requires Android API 26 or newer and contains native libraries for
the four standard Android ABIs. An Android App Bundle delivers only the native
library required by the target device. The AAR has no additional Maven runtime
dependencies.

Call it from a background dispatcher:

```kotlin
import jissrbypass.Jissrbypass
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

val client = Jissrbypass.newAutoClient()
val response = withContext(Dispatchers.IO) {
    client.execute(
        "GET",
        "https://api.example.com/v1/profile",
        ByteArray(0),
        """{"Authorization":"Bearer <access-token>"}"""
    )
}
check(response.statusCode in 200L..299L)
```

Reuse the client for related requests. Call `resetNetworkState()` after the
device changes networks so the next request selects a fresh route.
`execute` returns the status code,
response bytes, and response headers as JSON. Transport failures throw an
exception; non-2xx HTTP responses are returned normally.

Apps can choose the native-route detection budget and observe fallback:

```kotlin
import jissrbypass.Event
import jissrbypass.EventListener
import jissrbypass.Jissrbypass

val client = Jissrbypass.newAutoClientWithTimeout(500)
client.setEventListener(object : EventListener {
    override fun onEvent(event: Event?) {
        if (event?.getType() == "fallback.started") {
            analytics.record(
                "jissr_fallback",
                mapOf(
                    "host" to event.getHost(),
                    "reason" to event.getReason(),
                    "likelyInterference" to event.getLikelyInterference()
                )
            )
        }
    }
})
```

The listener receives decisions, not request content. A
`likelyInterference=true` native failure means fallback was safe before TLS
completed; it is useful diagnostic evidence but is not proof that a particular
network deliberately blocked the host. See [transport events](docs/events.md).

To build the AAR locally on Windows:

```powershell
.\scripts\build-android.ps1
```

See the [complete Android guide](docs/android.md) and
[sample application](examples/android).

Applications already using OkHttp or Retrofit can patch one central client
with the optional adapter:

```kotlin
val patched = JissrBypassInterceptor.patch(
    base = existingOkHttpClient,
    protectedHosts = listOf("api.example.com")
)
```

The adapter will be published separately as
`org.jissr:jissr-bypass-okhttp:<version>`, so the core AAR remains
dependency-free. See [client adapters](docs/adapters.md).

## Apple platforms

On macOS with Xcode installed:

```sh
./scripts/build-apple.sh
```

This creates `dist/apple/JissrBypass.xcframework` containing iOS device,
simulator, and macOS slices. A Swift example is in `examples/apple`.

The public release exposes the XCFramework through the organization-owned
`inqadh/jissr-bypass-swift` package, with a checksum-pinned GitHub Release
artifact. It also includes an optional URLSession adapter that recreates one
central session without changing existing `URLRequest` construction. See the
[Swift package release process](docs/swift-package.md).

## Canary protocol

`blocked.yusef.ren` is a project-hosted connectivity canary that is ready for
SDK consumers to use. It is an example hostname that is blocked on affected
Saudi networks: a normal HTTPS request may fail, while the same request through
Jissr Bypass should succeed. The client generates a random nonce and requests:

```text
GET https://blocked.yusef.ren/v1/canary?nonce=<base64url>
```

The response is accepted only when:

- `ok` is `true`;
- `protocol` is exactly `jissr-bypass/1`;
- the hostname matches the requested hostname;
- the unpredictable nonce is echoed exactly; and
- `issued_at` is recent.

The endpoint uses no cookies, account identifiers, or request body.

## Scope and limitations

Jissr Bypass is an HTTPS client transport, not a VPN and not a browser
extension. It helps when the destination supports ECH and the available DoH
resolvers are reachable. It cannot solve destination-IP blocking, total
network loss, a server without ECH, or a network that blocks every configured
resolver.

See the [integration guide](docs/integration-guide.md),
[threat model](docs/threat-model.md), and [security policy](SECURITY.md) before
integrating it into a production application.

## Distribution plan

- Go: tagged module and documentation on `pkg.go.dev`.
- Android: signed AAR on Maven Central under `org.jissr`.
- Apple: checksum-pinned XCFramework through Swift Package Manager.
- Optional adapters: OkHttp/Retrofit, URLSession, and Go `http.Client`
  host routing.
- Desktop: signed Windows, macOS, and Linux CLI archives with SHA-256 checksums
  and an SBOM.

The canonical source and Go module path are
[`github.com/inqadh/jissr-bypass`](https://github.com/inqadh/jissr-bypass).
Go module versions are immutable once published.

Performance methodology and reproducible Go/Android commands are in
[docs/benchmarks.md](docs/benchmarks.md).

## License

Apache License 2.0. See [LICENSE](LICENSE).
