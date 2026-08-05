# Jissr Bypass

Jissr Bypass is a native HTTPS transport for networks where ordinary DNS or a
plaintext TLS ClientHello is unreliable. It resolves both address and HTTPS
records over DNS-over-HTTPS and performs TLS 1.3 with Encrypted Client Hello
(ECH).

The same Go core builds for Android, iOS, macOS, Windows, and Linux.

> Status: pre-release. The public API may change before `v1.0.0`.

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
    response, err := bypass.NewDefaultClient().Do(request)
    if err != nil {
        log.Fatal(err)
    }
    defer response.Body.Close()
    fmt.Printf("HTTP %d\n", response.StatusCode)
}
```

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
    implementation("org.jissr:jissr-bypass-android:0.1.0-rc.2")
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

val response = withContext(Dispatchers.IO) {
    Jissrbypass.newClient().execute(
        "GET",
        "https://api.example.com/v1/profile",
        ByteArray(0),
        """{"Authorization":"Bearer <access-token>"}"""
    )
}
check(response.statusCode in 200L..299L)
```

Reuse the client for related requests. `execute` returns the status code,
response bytes, and response headers as JSON. Transport failures throw an
exception; non-2xx HTTP responses are returned normally.

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

`blocked.yusef.ren` is a controlled Worker used only as a connectivity canary.
The client generates a random nonce and requests:

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

## License

Apache License 2.0. See [LICENSE](LICENSE).
