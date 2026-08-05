# Android integration

This guide covers the Android package distributed through Maven Central as:

```text
org.jissr:jissr-bypass-android:<version>
```

## Platform requirements

| Requirement | Value |
|---|---|
| Minimum Android API | 26 |
| Package format | Android AAR |
| Native ABIs | `arm64-v8a`, `armeabi-v7a`, `x86`, `x86_64` |
| Maven runtime dependencies | None |
| License | Apache-2.0 |

The universal AAR contains all four native libraries. When an application is
distributed as an Android App Bundle, Google Play produces a device-specific
split containing only the appropriate ABI.

## Add the dependency

New Android projects normally include Maven Central already. Confirm that
`mavenCentral()` is present:

```kotlin
// settings.gradle.kts
dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
    }
}
```

Add a fixed version to the application module:

```kotlin
// app/build.gradle.kts
dependencies {
    implementation("org.jissr:jissr-bypass-android:0.1.0-rc.2")
}
```

Groovy projects use:

```groovy
dependencies {
    implementation "org.jissr:jissr-bypass-android:0.1.0-rc.2"
}
```

Avoid dynamic versions such as `+`, `latest.release`, or version ranges in
production builds.

## Manifest

The consuming application must allow network access:

```xml
<uses-permission android:name="android.permission.INTERNET" />
```

No VPN service, local proxy, accessibility service, or certificate installation
is required.

## Make a request

The API is synchronous. Call it on `Dispatchers.IO`, a worker, or another
background executor.

```kotlin
import jissrbypass.Client
import jissrbypass.Jissrbypass
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class ProtectedApi {
    private val client: Client = Jissrbypass.newClient()

    suspend fun profile(accessToken: String): String =
        withContext(Dispatchers.IO) {
            val response = client.execute(
                "GET",
                "https://api.example.com/v1/profile",
                ByteArray(0),
                """
                {
                  "Accept": "application/json",
                  "Authorization": "Bearer $accessToken"
                }
                """.trimIndent()
            )

            if (response.statusCode !in 200L..299L) {
                val errorBody = response.body.toString(Charsets.UTF_8)
                error("HTTP ${response.statusCode}: $errorBody")
            }
            response.body.toString(Charsets.UTF_8)
        }
}
```

Reuse a client across requests. The client maintains the resolver configuration
and HTTP transport needed for consistent behavior.

## Request and response model

`Client.execute(method, url, body, headersJSON)` accepts:

- an HTTP method such as `GET`, `POST`, `PUT`, or `DELETE`;
- an absolute `https://` URL;
- raw request-body bytes; and
- a JSON object of request headers.

A header value may be a string:

```json
{"Accept":"application/json"}
```

or an array:

```json
{"Accept":["application/json"],"X-Trace":["one","two"]}
```

The returned `Response` exposes:

- `statusCode`: HTTP status as a `Long`;
- `body`: response bytes; and
- `headersJSON`: a JSON object whose values are arrays of strings.

An HTTP error status is still a valid response. DNS, TLS, ECH, timeout, and
other transport failures throw an exception.

## Cancellation and timeouts

The generated mobile binding is blocking and does not currently expose a
per-request cancellation method. Apply an application-level deadline and avoid
starting work that is no longer needed. A coroutine cancellation stops awaiting
the result but may not immediately interrupt native work.

Never call the API from the Android main thread.

## Diagnostics

Use `probe` to distinguish resolution, HTTPS-record, ECH, and connection
failures:

```kotlin
val probeJSON = withContext(Dispatchers.IO) {
    client.probe("api.example.com")
}
```

Use the maintained nonce endpoint to verify the complete client path:

```kotlin
val canaryJSON = withContext(Dispatchers.IO) {
    client.verifyCanary(
        "https://blocked.yusef.ren/v1/canary"
    )
}
```

Do not send authentication tokens, account data, or application payloads to a
diagnostic canary.

## R8 and ProGuard

The published AAR contains consumer ProGuard rules for generated bindings. If
the application adds broad custom shrinking rules, retain:

```proguard
-keep class jissrbypass.** { *; }
-keep class go.** { *; }
```

Exercise a minified release build on a physical device before publishing.

## Server requirements

The target hostname must support TLS 1.3 ECH and publish a usable HTTPS DNS
record. The application continues to use ordinary HTTPS semantics:
authentication, authorization, cookies, request bodies, status codes, and
certificate validation are unchanged.

Every request to the protected hostname must use Jissr Bypass. Falling back to
the platform HTTP client for the same hostname can reproduce the original
network failure.

For the server-side configuration and rollout checklist, see the
[A-to-Z integration guide](integration-guide.md).

## Release verification

Before shipping:

1. Pin an immutable package version.
2. Build both debug and minified release variants.
3. Exercise a real request and the nonce canary on at least two networks.
4. Confirm authentication headers, request bodies, non-2xx responses, and
   application timeouts.
5. Confirm the packaged APK/AAB contains the expected device ABI only.
