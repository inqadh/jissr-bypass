# Integration guide

This guide takes an application from an ordinary API hostname to
ECH-protected requests with Jissr Bypass on Android, Apple platforms, and
desktop systems.

Jissr Bypass is an HTTPS client transport. It is not a VPN, proxy, DNS host,
CDN, or origin server. Your API must be served through an edge provider that
terminates Encrypted Client Hello (ECH) and publishes a valid ECH configuration
in its HTTPS DNS record.

## 1. Understand the request path

An ordinary HTTPS client normally asks the operating system for an address and
then sends a TLS ClientHello containing the destination hostname. Jissr Bypass
instead:

1. asks a DNS-over-HTTPS resolver for the hostname's address record;
2. asks the same resolver for its HTTPS record and binary ECH configuration;
3. connects to the returned edge address on TCP port 443;
4. performs TLS 1.3 with the real hostname inside ECH;
5. verifies the server certificate against the real hostname; and
6. sends the application's normal HTTP method, headers, and body inside TLS.

The application still talks directly to its own API. Authentication,
authorization, cookies, JSON, status codes, and server behavior do not change.
Only DNS resolution and TLS connection establishment are replaced.

```text
application
    |
    | Jissr Bypass: DoH + ECH + HTTPS
    v
ECH-capable edge provider
    |
    | provider-to-origin HTTPS
    v
your API origin
```

## 2. Confirm that this design fits

Use Jissr Bypass when:

- the application controls or has permission to call the destination API;
- the API can be placed behind an ECH-capable edge provider;
- requests use HTTPS on port 443; and
- each protected request can be made through the library instead of the
  platform's ordinary HTTP client.

It does not solve destination-IP blocking, total loss of connectivity, or a
provider that does not publish a usable ECH configuration. It does not carry
arbitrary TCP, UDP, WebSocket, or browser traffic.

## 3. Prepare the API hostname

Use a dedicated, stable hostname such as `api.example.com`. The hostname should
serve the complete API contract needed by the app:

- a valid public TLS certificate;
- a lightweight unauthenticated health endpoint such as `GET /health`;
- authenticated application endpoints;
- responses no larger than 4 MiB; and
- no required redirect to a different hostname.

Keep the origin inaccessible except through your edge provider where practical.
Configure provider-to-origin TLS with certificate verification. Do not put API
keys, access tokens, or session data in a URL or query string.

### Cloudflare setup

1. Add the domain to Cloudflare and make its assigned nameservers authoritative.
2. Create an `A`, `AAAA`, or `CNAME` record for `api.example.com` that points to
   the origin.
3. Enable the orange-cloud **Proxied** state for that record. A DNS-only record
   exposes the origin and cannot use Cloudflare's edge ECH configuration.
4. In **SSL/TLS**, use **Full (strict)** and install a valid certificate at the
   origin.
5. Confirm that Universal SSL is active and that HTTP/2 or HTTP/3 is enabled.
6. In **SSL/TLS → Edge Certificates**, confirm that Encrypted Client Hello is
   enabled. Cloudflare documents ECH as enabled by default for Free zones; on
   other plans, use this control to enable it.
7. Configure WAF and rate-limit rules for the API contract. Do not depend on a
   particular Jissr Bypass User-Agent for authorization or allowlisting.

For a proxied Cloudflare hostname, Cloudflare generates the HTTPS DNS record
when the required edge features are active. Do not create a conflicting manual
HTTPS record.

Cloudflare references:

- [Encrypted Client Hello](https://developers.cloudflare.com/ssl/edge-certificates/ech/)
- [HTTPS DNS record generation](https://developers.cloudflare.com/ssl/reference/browser-compatibility/)
- [How proxied DNS records work](https://developers.cloudflare.com/fundamentals/concepts/how-cloudflare-works/)

### Another edge provider

An equivalent provider is compatible only if it:

- terminates TLS 1.3 ECH for your hostname;
- publishes an HTTPS or SVCB DNS record containing a valid `ech` parameter;
- returns an address that belongs to the same edge configuration;
- presents a publicly trusted certificate for your API hostname; and
- forwards the request to your origin without changing the API contract.

Verify these capabilities in the provider's current documentation. Do not
assume that a service supports ECH merely because it offers a CDN, TLS 1.3, or
DNS-over-HTTPS. Google Public DNS in the default Jissr Bypass configuration is
a resolver fallback; it does not make a Google-hosted origin ECH-capable.

## 4. Verify the server before integrating the app

First verify ordinary server health:

```sh
curl --fail-with-body https://api.example.com/health
```

Then inspect the public HTTPS DNS record:

```sh
dig @1.1.1.1 api.example.com HTTPS
```

The answer should contain an `ech=` service parameter. Finally, use the Jissr
Bypass diagnostic command, which validates the binary record and performs an
ECH handshake:

```sh
go run ./cmd/jissr-bypass -probe api.example.com
```

A successful ordinary `curl` request alone does not prove that ECH is
available. A failed probe should be fixed at the DNS or edge-provider layer
before mobile integration.

## 5. Install the library

### Android

Build the binding:

```powershell
.\scripts\build-android.ps1
```

Copy `dist/android/jissr-bypass.aar` into the app's `app/libs` directory and
add it to the app module:

```kotlin
dependencies {
    implementation(files("libs/jissr-bypass.aar"))
}
```

Ensure the manifest contains:

```xml
<uses-permission android:name="android.permission.INTERNET" />
```

The planned public Maven coordinate is
`org.jissr:jissr-bypass-android:<version>`. Until it is published, pin and
checksum the AAR supplied with the release.

### iOS and macOS

Build on macOS with Xcode installed:

```sh
./scripts/build-apple.sh
```

Drag `dist/apple/JissrBypass.xcframework` into the Xcode project, select the app
target, and use **Embed & Sign**. Add `import Jissrbypass` in Swift sources that
call the binding.

The planned public distribution is a checksum-pinned binary target through
Swift Package Manager. Until it is published, pin and checksum the XCFramework
from the release.

### Go applications on Windows, macOS, and Linux

Add the module:

```sh
go get github.com/jissr-org/jissr-bypass@<version>
```

Do not use `@latest` in a reproducible release. Pin a reviewed semantic version.

## 6. Replace calls to protected endpoints

Every request to the protected hostname must use Jissr Bypass. If one screen
uses the library but another uses `URLSession`, OkHttp, Retrofit's default
transport, or the operating system resolver, the latter request does not gain
the same DNS and ECH behavior.

The mobile `Execute` API accepts:

- an HTTP method;
- a complete `https://` URL on port 443;
- body bytes, or an empty byte array;
- request headers encoded as a JSON object; and
- returns status, body, and response headers.

Header values may be a string or an array of strings:

```json
{
  "Authorization": "Bearer <access-token>",
  "Accept": "application/json",
  "X-Trace-ID": ["first-value", "second-value"]
}
```

`Host`, `Content-Length`, `Transfer-Encoding`, and `Connection` are managed by
the transport and cannot be supplied by the application.

### Android: authenticated GET

Run network calls on a background dispatcher:

```kotlin
import jissrbypass.Jissrbypass
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

suspend fun loadProfile(accessToken: String): String =
    withContext(Dispatchers.IO) {
        val headers = """
            {
              "Authorization": "Bearer $accessToken",
              "Accept": "application/json"
            }
        """.trimIndent()

        val response = Jissrbypass.newClient().execute(
            "GET",
            "https://api.example.com/v1/profile",
            ByteArray(0),
            headers
        )
        if (response.statusCode !in 200L..299L) {
            error("API returned HTTP ${response.statusCode}")
        }
        response.body.toString(Charsets.UTF_8)
    }
```

### Android: JSON POST

```kotlin
val payload = """{"message":"hello"}""".toByteArray(Charsets.UTF_8)
val headers = """
    {
      "Authorization": "Bearer $accessToken",
      "Content-Type": "application/json",
      "Accept": "application/json"
    }
""".trimIndent()

val response = Jissrbypass.newClient().execute(
    "POST",
    "https://api.example.com/v1/messages",
    payload,
    headers
)
```

### Apple: authenticated GET

The binding is synchronous, so call it outside the main actor:

```swift
import Foundation
import Jissrbypass

struct BypassHTTPResponse {
    let statusCode: Int
    let body: Data
    let headersJSON: String
}

func loadProfile(accessToken: String) async throws -> BypassHTTPResponse {
    try await Task.detached(priority: .userInitiated) {
        let headers = """
        {"Authorization":"Bearer \(accessToken)","Accept":"application/json"}
        """
        var error: NSError?
        let response = JissrbypassNewClient().execute(
            "GET",
            url: "https://api.example.com/v1/profile",
            body: Data(),
            headersJSON: headers,
            error: &error
        )
        if let error {
            throw error
        }
        guard let response else {
            throw NSError(domain: "JissrBypass", code: -1)
        }
        return BypassHTTPResponse(
            statusCode: Int(response.getStatusCode()),
            body: response.getBody(),
            headersJSON: response.getHeadersJSON()
        )
    }.value
}
```

Generated Swift names can vary slightly by the gomobile version. Treat the
generated framework header as the source of truth if Xcode suggests a different
argument label.

### Go: authenticated request

The root Go package accepts a normal `http.Request`, which is the most flexible
desktop API:

```go
package api

import (
    "context"
    "fmt"
    "io"
    "net/http"

    bypass "github.com/jissr-org/jissr-bypass"
)

func LoadProfile(ctx context.Context, token string) ([]byte, error) {
    request, err := http.NewRequestWithContext(
        ctx,
        http.MethodGet,
        "https://api.example.com/v1/profile",
        nil,
    )
    if err != nil {
        return nil, err
    }
    request.Header.Set("Authorization", "Bearer "+token)
    request.Header.Set("Accept", "application/json")

    response, err := bypass.NewDefaultClient().Do(request)
    if err != nil {
        return nil, err
    }
    defer response.Body.Close()
    if response.StatusCode < 200 || response.StatusCode > 299 {
        return nil, fmt.Errorf("API returned %s", response.Status)
    }
    return io.ReadAll(response.Body)
}
```

Reuse one client for multiple requests. `Client` is safe for concurrent use.

## 7. Handle responses and redirects

`Execute` returns non-success HTTP responses without turning them into transport
errors. The application must handle its own API statuses, for example:

- `200`–`299`: process the response;
- `401` or `403`: refresh credentials or ask the user to sign in;
- `429`: follow the server's retry policy;
- `500`–`599`: use bounded retries appropriate for the operation.

Response bodies are limited to 4 MiB. Use ordinary API endpoints that return
bounded JSON; the current mobile API is not a streaming file-transfer client.

Redirects are deliberately not followed automatically. A `3xx` response is
returned to the application. If redirects are required:

1. read `Location` from `headersJSON`;
2. resolve the location against the original URL;
3. require `https`, port 443, and an application allowlisted hostname; and
4. issue a new request through Jissr Bypass.

Never hand the redirect to the platform's default HTTP client.

## 8. Add health checks without making them a dependency

At application startup or before an important operation, `Probe` can diagnose
the protected hostname:

```kotlin
val probeJSON = Jissrbypass.newClient().probe("api.example.com")
```

The result identifies the stage:

- `resolve`: no mutually consistent address and ECH answer;
- `tcp`: the selected edge address was unreachable;
- `ech`: TLS or ECH negotiation failed;
- `complete`: ECH negotiation succeeded.

Use the probe for diagnostics, not as a mandatory preflight before every API
request. A preflight doubles latency and introduces a race between the check and
the real call.

`https://blocked.yusef.ren/v1/canary` is a public integration example. The
`VerifyCanary` method generates a random nonce and verifies the echoed protocol,
host, and timestamp:

```kotlin
val resultJSON = Jissrbypass.newClient()
    .verifyCanary("https://blocked.yusef.ren/v1/canary")
```

For production monitoring, create a canary under a domain you control and keep
it independent of authentication, databases, and application state. Do not make
another organization's canary a prerequisite for your API.

## 9. Preserve security properties

- Never disable certificate verification.
- Never fall back to plaintext HTTP.
- Do not silently retry the same non-idempotent request after an ambiguous
  transport failure.
- Keep access tokens in headers, not URLs.
- Redact tokens and response bodies from logs.
- Validate redirected hostnames against an application allowlist.
- Keep edge-to-origin TLS in strict verification mode.
- Apply normal server-side authentication, authorization, abuse controls, and
  rate limits. ECH is privacy for connection setup, not an access-control
  mechanism.

HTTP headers, including the default `Jissr-Bypass/0.1` User-Agent, are inside
the encrypted TLS application data. A network observer can still see resolver
and edge IP addresses, timing, sizes, and the public ECH cover name. See the
[threat model](threat-model.md) for the complete boundary.

## 10. Test the complete application

Before release, test:

1. ordinary Wi-Fi and mobile networks;
2. a network where ordinary access to the sample hostname is disrupted;
3. authenticated GET and JSON POST requests;
4. expired credentials and all expected non-2xx statuses;
5. offline, captive-portal, and high-latency behavior;
6. a `3xx` response and your redirect allowlist;
7. Cloudflare and Google resolver availability;
8. provider ECH key rotation; and
9. rapid repeated UI actions while a request is in progress.

Keep network work off the UI thread. Disable or coalesce duplicate UI actions,
and tie results to the screen or task that initiated them. This prevents a
correct transport request from becoming an application lifecycle crash.

## 11. Production checklist

- [ ] The API hostname is proxied through a verified ECH-capable provider.
- [ ] The public HTTPS DNS record contains a valid `ech` parameter.
- [ ] The provider presents a trusted certificate for the API hostname.
- [ ] Provider-to-origin TLS uses strict verification.
- [ ] Every protected request uses Jissr Bypass.
- [ ] Authentication headers and JSON content types are set explicitly.
- [ ] Non-2xx responses and redirects are handled by the application.
- [ ] Requests run outside the mobile UI thread.
- [ ] Logs redact secrets and response data.
- [ ] The SDK artifact and version are pinned and checksummed.
- [ ] A controlled ECH canary is monitored from an independent network.
- [ ] The full test matrix passes on Android, Apple, and desktop targets in use.
