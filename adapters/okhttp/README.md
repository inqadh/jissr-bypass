# Jissr Bypass for OkHttp and Retrofit

The optional adapter patches one central OkHttp client while leaving existing
request call sites and Retrofit service interfaces unchanged.

```kotlin
val patched = JissrBypassInterceptor.patch(
    base = existingOkHttpClient,
    protectedHosts = listOf("api.example.com")
)

val retrofit = existingRetrofit.newBuilder()
    .client(patched)
    .build()
```

Only exact configured hostnames are routed through Jissr Bypass. All other
requests continue through the base OkHttp configuration.

The initial adapter buffers request and response bodies. It does not support
WebSockets, duplex bodies, one-shot bodies, or streaming responses. Install it
as the last application interceptor so existing authentication and request
metadata interceptors run first.
