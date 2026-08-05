package org.jissr.bypass.okhttp

import java.util.concurrent.atomic.AtomicInteger
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class JissrBypassInterceptorTest {
    @Test
    fun routesExactProtectedHostAndPreservesRequest() {
        val calls = AtomicInteger()
        val interceptor = JissrBypassInterceptor(
            executor = BypassExecutor { method, url, body, headersJSON ->
                calls.incrementAndGet()
                assertEquals("POST", method)
                assertEquals("https://api.example.com/v1", url)
                assertEquals("""{"hello":"world"}""", body.toString(Charsets.UTF_8))
                assertEquals("Bearer test", JSONObject(headersJSON).getString("Authorization"))
                BypassResult(
                    statusCode = 201,
                    protocol = "HTTP/2.0",
                    headersJSON = """{"Content-Type":["application/json"],"X-Test":["one","two"]}""",
                    body = """{"ok":true}""".toByteArray()
                )
            },
            protectedHosts = listOf("API.Example.com.")
        )
        val client = OkHttpClient.Builder().addInterceptor(interceptor).build()
        val request = Request.Builder()
            .url("https://api.example.com/v1")
            .header("Authorization", "Bearer test")
            .post(
                """{"hello":"world"}"""
                    .toRequestBody("application/json".toMediaType())
            )
            .build()

        client.newCall(request).execute().use { response ->
            assertEquals(201, response.code)
            assertEquals("h2", response.protocol.toString())
            assertEquals(listOf("one", "two"), response.headers.values("X-Test"))
            assertEquals("""{"ok":true}""", response.body.string())
        }
        assertEquals(1, calls.get())
    }

    @Test
    fun leavesUnconfiguredHostsOnOkHttp() {
        MockWebServer().use { server ->
            server.enqueue(MockResponse().setBody("ordinary"))
            val calls = AtomicInteger()
            val interceptor = JissrBypassInterceptor(
                executor = BypassExecutor { _, _, _, _ ->
                    calls.incrementAndGet()
                    error("protected executor must not be called")
                },
                protectedHosts = listOf("example.com")
            )
            val client = OkHttpClient.Builder().addInterceptor(interceptor).build()

            client.newCall(Request.Builder().url(server.url("/")).build())
                .execute()
                .use { response ->
                    assertEquals("ordinary", response.body.string())
                }
            assertEquals(0, calls.get())
        }
    }

    @Test
    fun patchDoesNotMutateBaseClient() {
        val base = OkHttpClient()
        val patched = JissrBypassInterceptor.patchWithExecutor(
            base = base,
            protectedHosts = listOf("api.example.com"),
            executor = BypassExecutor { _, _, _, _ ->
                error("executor should not be called")
            }
        )

        assertTrue(base !== patched)
        assertEquals(0, base.interceptors.size)
        assertEquals(1, patched.interceptors.size)
    }
}
