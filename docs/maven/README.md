# Jissr Bypass Android

Jissr Bypass is a native HTTPS transport that resolves address and HTTPS
records over DNS-over-HTTPS and establishes TLS 1.3 connections with Encrypted
Client Hello (ECH). It is an application HTTP client, not a VPN.

## Requirements

- Android API 26 or newer.
- Internet permission in the application manifest.
- A destination served through an ECH-capable CDN or origin.
- Network calls made away from the Android main thread.

The AAR contains native libraries for `arm64-v8a`, `armeabi-v7a`, `x86`, and
`x86_64`. Gradle and Android App Bundles deliver only the ABI needed by each
device. The package has no additional Maven runtime dependencies.

## Install

Gradle Kotlin DSL:

```kotlin
dependencies {
    implementation("org.jissr:jissr-bypass-android:<version>")
}
```

Gradle Groovy DSL:

```groovy
dependencies {
    implementation "org.jissr:jissr-bypass-android:<version>"
}
```

Use a fixed release version. Do not use dynamic selectors such as `+`.

## Kotlin example

```kotlin
import jissrbypass.Jissrbypass
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

val response = withContext(Dispatchers.IO) {
    Jissrbypass.newClient().execute(
        "GET",
        "https://api.example.com/v1/profile",
        ByteArray(0),
        """{"Accept":"application/json","Authorization":"Bearer <token>"}"""
    )
}

check(response.statusCode in 200L..299L)
val responseBody = response.body.toString(Charsets.UTF_8)
val responseHeaders = response.headersJSON
```

`execute` accepts request headers as a JSON object. Each value may be a string
or an array of strings. It returns the status code, response bytes, and response
headers encoded as a JSON object of string arrays.

## Public API

- `Jissrbypass.newClient()` creates a client with the maintained resolver
  fallback set.
- `Client.execute(...)` performs an HTTPS request with request headers.
- `Client.fetch(...)` performs a request and returns only the response body.
- `Client.probe(host)` diagnoses DNS, HTTPS-record, ECH, and connection stages.
- `Client.verifyCanary(url)` verifies the versioned nonce canary protocol.

The generated HTML reference is included in this Javadoc JAR. Open
`index.html` after extracting it.

## Operational guidance

- Reuse one client instead of creating a client for every request.
- Apply application-level timeouts and cancellation around blocking calls.
- Treat transport errors separately from non-2xx HTTP responses.
- Continue using ordinary HTTPS authentication, authorization, and certificate
  validation.
- Ensure every request to a protected hostname uses the same transport.

The default User-Agent and all HTTP headers are carried inside the encrypted
HTTPS request.

## Shrinking

The AAR includes consumer ProGuard rules for its generated bindings. Do not
remove the `jissrbypass.*` or `go.*` binding classes through custom R8 rules.

## License and source

Apache License 2.0.

- Source: https://github.com/inqadh/jissr-bypass
- Android example: https://github.com/inqadh/jissr-bypass/tree/main/examples/android
- Integration guide: https://github.com/inqadh/jissr-bypass/blob/main/docs/integration-guide.md
- Issues: https://github.com/inqadh/jissr-bypass/issues
