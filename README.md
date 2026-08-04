# Jissr Bypass

Jissr Bypass is a native HTTPS transport for networks where ordinary DNS or a
plaintext TLS ClientHello is unreliable. It resolves both address and HTTPS
records over DNS-over-HTTPS and performs TLS 1.3 with Encrypted Client Hello
(ECH).

The same Go core builds for Android, iOS, macOS, Windows, and Linux.

> Status: pre-release. The public API may change before `v1.0.0`.

## Why wire-format DNS-over-HTTPS?

An ECH configuration is binary data carried in the `ech` service parameter of
an HTTPS DNS record. Jissr Bypass requests `application/dns-message`, the wire
format standardized by RFC 8484. This preserves the exact record bytes.

JSON DNS APIs are retained only as a compatibility fallback because providers
may represent the same HTTPS record as presentation text or RFC 3597 raw data.
The library understands both representations, but it does not depend on either
one for its primary path.

Each resolver must return both a valid address and a valid ECH configuration.
The library never combines an address from one resolver with ECH data from
another.

## Go and desktop

```go
package main

import (
    "fmt"
    "log"

    bypass "github.com/jissr-org/jissr-bypass"
)

func main() {
    result, err := bypass.NewDefaultClient().
        VerifyCanary("https://blocked.yusef.ren/v1/canary")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("verified %s at %s\n", result.Host, result.IssuedAt)
}
```

The diagnostic CLI uses the same core:

```sh
go run ./cmd/jissr-bypass
go run ./cmd/jissr-bypass -probe blocked.yusef.ren
```

## Android

Build the AAR on Windows:

```powershell
.\scripts\build-android.ps1
```

Add `dist/android/jissr-bypass.aar` to the application's dependencies, then:

```kotlin
import jissrbypass.Jissrbypass

val resultJson = Jissrbypass.newClient()
    .verifyCanary("https://blocked.yusef.ren/v1/canary")
```

Network calls are blocking and must run on a background dispatcher. The
complete sample is in `examples/android`.

The planned public coordinate is:

```kotlin
implementation("org.jissr:jissr-bypass-android:0.1.0")
```

## Apple platforms

On macOS with Xcode installed:

```sh
./scripts/build-apple.sh
```

This creates `dist/apple/JissrBypass.xcframework` containing iOS device,
simulator, and macOS slices. A Swift example is in `examples/apple`.

The public release will expose the XCFramework through Swift Package Manager
with a checksum-pinned GitHub Release artifact.

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

See [docs/threat-model.md](docs/threat-model.md) and [SECURITY.md](SECURITY.md)
before integrating it into a production application.

## Distribution plan

- Go: tagged module and documentation on `pkg.go.dev`.
- Android: signed AAR on Maven Central under `org.jissr`.
- Apple: checksum-pinned XCFramework through Swift Package Manager.
- Desktop: signed Windows, macOS, and Linux CLI archives with SHA-256 checksums
  and an SBOM.

No public tag should be created until the final GitHub organization and module
path have been confirmed. Go module versions are immutable once published.

## License

Apache License 2.0. See [LICENSE](LICENSE).
