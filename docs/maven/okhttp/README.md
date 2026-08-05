# Jissr Bypass for OkHttp and Retrofit

This optional adapter patches one central OkHttp client while preserving
existing OkHttp calls and Retrofit service interfaces.

```kotlin
val patched = JissrBypassInterceptor.patch(
    base = existingOkHttpClient,
    protectedHosts = listOf("api.example.com")
)
```

Add both artifacts with the same immutable version:

```kotlin
dependencies {
    implementation("org.jissr:jissr-bypass-android:<version>")
    implementation("org.jissr:jissr-bypass-okhttp:<version>")
}
```

Only exact configured hosts use Jissr Bypass. Other hosts continue through the
base OkHttp transport.

The adapter supports bounded HTTPS requests used by ordinary JSON APIs. It
does not support WebSockets, duplex or one-shot bodies, streaming responses,
or immediate cancellation of in-progress native work.
