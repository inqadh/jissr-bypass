package org.jissr.bypass.okhttp

import java.io.IOException
import java.net.IDN
import java.util.Locale
import jissrbypass.Client
import jissrbypass.Jissrbypass
import okhttp3.Headers
import okhttp3.Interceptor
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.OkHttpClient
import okhttp3.Protocol
import okhttp3.Request
import okhttp3.Response
import okhttp3.ResponseBody.Companion.toResponseBody
import okio.Buffer
import org.json.JSONArray
import org.json.JSONObject

/**
 * Routes exact configured HTTPS hosts through Jissr Bypass.
 *
 * Install this as the last application interceptor on the central OkHttp
 * client. Retrofit can continue using that client without changing service
 * interfaces. Protected request and response bodies are buffered in memory.
 */
class JissrBypassInterceptor private constructor(
    private val executor: BypassExecutor,
    protectedHosts: Collection<String>,
    private val maxRequestBytes: Long
) : Interceptor {
    private val hosts: Set<String> = protectedHosts.map(::canonicalHost).toSet()

    init {
        require(hosts.isNotEmpty()) { "At least one protected host is required" }
        require(maxRequestBytes > 0) { "maxRequestBytes must be positive" }
    }

    @JvmOverloads
    constructor(
        protectedHosts: Collection<String>,
        client: Client = Jissrbypass.newClient(),
        maxRequestBytes: Long = DEFAULT_MAX_REQUEST_BYTES
    ) : this(NativeBypassExecutor(client), protectedHosts, maxRequestBytes)

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        if (canonicalHost(request.url.host) !in hosts) {
            return chain.proceed(request)
        }
        if (!request.url.isHttps || request.url.port != 443) {
            throw IOException("Jissr Bypass only supports HTTPS on port 443")
        }
        if (chain.call().isCanceled()) {
            throw IOException("Canceled")
        }

        val requestBody = request.body?.let { body ->
            if (body.isDuplex()) {
                throw IOException("Duplex request bodies are not supported")
            }
            if (body.isOneShot()) {
                throw IOException("One-shot request bodies are not supported")
            }
            Buffer().use { buffer ->
                body.writeTo(buffer)
                if (buffer.size > maxRequestBytes) {
                    throw IOException("Request body exceeds $maxRequestBytes bytes")
                }
                buffer.readByteArray()
            }
        } ?: ByteArray(0)

        val startedAt = System.currentTimeMillis()
        val result = executor.execute(
            method = request.method,
            url = request.url.toString(),
            body = requestBody,
            headersJSON = encodeRequestHeaders(request)
        )
        if (chain.call().isCanceled()) {
            throw IOException("Canceled")
        }

        val headers = decodeResponseHeaders(result.headersJSON)
        val mediaType = headers["Content-Type"]?.toMediaTypeOrNull()
        return Response.Builder()
            .request(request)
            .protocol(protocol(result.protocol))
            .code(result.statusCode)
            .message("")
            .headers(headers)
            .body(result.body.toResponseBody(mediaType))
            .sentRequestAtMillis(startedAt)
            .receivedResponseAtMillis(System.currentTimeMillis())
            .build()
    }

    companion object {
        const val DEFAULT_MAX_REQUEST_BYTES: Long = 4L * 1024L * 1024L

        /** Installs host routing without mutating the original client. */
        @JvmStatic
        @JvmOverloads
        fun patch(
            base: OkHttpClient,
            protectedHosts: Collection<String>,
            client: Client = Jissrbypass.newClient(),
            maxRequestBytes: Long = DEFAULT_MAX_REQUEST_BYTES
        ): OkHttpClient = patchWithExecutor(
            base = base,
            protectedHosts = protectedHosts,
            executor = NativeBypassExecutor(client),
            maxRequestBytes = maxRequestBytes
        )

        internal fun patchWithExecutor(
            base: OkHttpClient,
            protectedHosts: Collection<String>,
            executor: BypassExecutor,
            maxRequestBytes: Long = DEFAULT_MAX_REQUEST_BYTES
        ): OkHttpClient = base.newBuilder()
            .addInterceptor(
                JissrBypassInterceptor(
                    executor = executor,
                    protectedHosts = protectedHosts,
                    maxRequestBytes = maxRequestBytes
                )
            )
            .build()

        private fun canonicalHost(value: String): String {
            val trimmed = value.trim().trimEnd('.')
            require(trimmed.isNotEmpty()) { "Protected host is required" }
            val ascii = try {
                IDN.toASCII(trimmed, IDN.USE_STD3_ASCII_RULES)
                    .lowercase(Locale.ROOT)
            } catch (error: IllegalArgumentException) {
                throw IllegalArgumentException("Invalid protected host", error)
            }
            require(ascii.length <= 253 && !isIpLiteral(ascii)) {
                "Protected host must be a DNS name"
            }
            return ascii
        }

        private fun isIpLiteral(host: String): Boolean =
            host.contains(':') ||
                host.split('.').let { labels ->
                    labels.size == 4 && labels.all { label ->
                        label.toIntOrNull()?.let { it in 0..255 } == true
                    }
                }

        private fun encodeRequestHeaders(request: Request): String {
            val result = JSONObject()
            for (name in request.headers.names()) {
                if (name.lowercase(Locale.ROOT) in MANAGED_REQUEST_HEADERS) {
                    continue
                }
                val values = request.headers.values(name)
                if (values.size == 1) {
                    result.put(name, values[0])
                } else {
                    result.put(name, JSONArray(values))
                }
            }
            return result.toString()
        }

        private fun decodeResponseHeaders(encoded: String): Headers {
            val source = JSONObject(encoded)
            val result = Headers.Builder()
            for (name in source.keys()) {
                if (name.lowercase(Locale.ROOT) in MANAGED_RESPONSE_HEADERS) {
                    continue
                }
                when (val value = source.get(name)) {
                    is JSONArray -> {
                        for (index in 0 until value.length()) {
                            result.add(name, value.getString(index))
                        }
                    }
                    is String -> result.add(name, value)
                    else -> throw IOException("Invalid response header value for $name")
                }
            }
            return result.build()
        }

        private fun protocol(value: String): Protocol = when (value.uppercase(Locale.ROOT)) {
            "HTTP/2", "HTTP/2.0" -> Protocol.HTTP_2
            "HTTP/1.0" -> Protocol.HTTP_1_0
            else -> Protocol.HTTP_1_1
        }

        private val MANAGED_REQUEST_HEADERS = setOf(
            "connection",
            "content-length",
            "host",
            "transfer-encoding"
        )
        private val MANAGED_RESPONSE_HEADERS = setOf(
            "connection",
            "content-length",
            "transfer-encoding"
        )
    }

    internal constructor(
        executor: BypassExecutor,
        protectedHosts: Collection<String>
    ) : this(executor, protectedHosts, DEFAULT_MAX_REQUEST_BYTES)
}

internal data class BypassResult(
    val statusCode: Int,
    val protocol: String,
    val headersJSON: String,
    val body: ByteArray
)

internal fun interface BypassExecutor {
    @Throws(Exception::class)
    fun execute(
        method: String,
        url: String,
        body: ByteArray,
        headersJSON: String
    ): BypassResult
}

private class NativeBypassExecutor(
    private val client: Client
) : BypassExecutor {
    override fun execute(
        method: String,
        url: String,
        body: ByteArray,
        headersJSON: String
    ): BypassResult {
        val response = client.execute(method, url, body, headersJSON)
        return BypassResult(
            statusCode = response.statusCode.toInt(),
            protocol = response.protocol,
            headersJSON = response.headersJSON,
            body = response.body
        )
    }
}
