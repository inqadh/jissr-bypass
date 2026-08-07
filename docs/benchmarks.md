# Performance benchmarks

Jissr Bypass keeps performance tests beside the transport implementation so
connection reuse, route caching, and fallback behavior can be checked as the
library evolves.

## Deterministic Go benchmarks

Run the unit suite and in-memory benchmarks:

```sh
go test ./...
go test -run '^$' \
  -bench 'Benchmark(ResolutionCacheHit|AutoCachedNativeRoute|AutoCachedBypassRoute)$' \
  -benchmem ./...
```

Reference result on Windows/amd64:

| Benchmark | Time | Allocations |
| --- | ---: | ---: |
| Cached resolution | 218 ns/op | 3 |
| Cached native route | 1.07 µs/op | 27 |
| Cached bypass route | 649 ns/op | 16 |

These benchmarks isolate library bookkeeping. They do not simulate Internet
latency.

An opt-in live benchmark is also available:

```sh
JISSR_BENCHMARK_URL=https://blocked.yusef.ren/v1/health \
  go test -run '^$' -bench BenchmarkLiveTransports -benchtime=10x
```

Run it against both an affected canary and an ordinary reachable API. A Wafai
reference run against the unblocked `https://alneasher.net/api/health` path
measured approximately 343 ms for pooled native, 361 ms for pooled bypass, and
361 ms for auto. The roughly 5% spread is within normal Internet-path
variation; the cached auto route does not repeat the fallback probe.

The connected Android reference run on the same unblocked endpoint completed
10/10 requests for both paths: pooled native measured 642 ms median and pooled
auto measured 599 ms median. This confirms the auto client cached native rather
than paying a fallback probe on every request. Internet latency varied
substantially between runs, so compare route behavior and repeated samples
rather than treating either absolute value as a product guarantee.

## Android instrumentation benchmark

The Android app integration contains
`JissrBypassBenchmarkTest`. It runs the native, bypass-only, and auto clients
against the public canary through the same isolated ECH process used by the
application.

```sh
./gradlew :app:connectedDebugAndroidTest \
  -Pandroid.testInstrumentationRunnerArguments.class=com.jisr.app.JissrBypassBenchmarkTest
```

Disconnect any VPN before running the test. Otherwise the result measures the
VPN route instead of the device's underlying network.

Reference measurements from a Blade A76 5G running Android 15 on an affected
Wi-Fi network, collected on 2026-08-07 with two runs of ten requests:

| Path | Success | Median / range |
| --- | ---: | ---: |
| Native | 0/20 | failures varied from 56 ms to 1.40 s median |
| Bypass, first request | 2/2 | 1.16–1.18 s |
| Bypass, pooled | 20/20 | 386–395 ms median |
| Auto, first request | 2/2 | 1.92–2.28 s |
| Auto, cached route | 20/20 | 109–390 ms median |

The native failure time varies because the network may reject either the TCP
connection or the TLS handshake. The auto policy limits its native route check
to 750 ms in total, then caches the bypass decision. The first auto request
still includes both route detection and a fresh DoH/ECH/TLS connection; later
requests reuse the selected transport.

Treat live values as environment-specific. The regression requirements are:

- bypass and auto requests succeed;
- repeated requests reuse their client and transport;
- route and resolution caches reset on a network change; and
- native failure never causes an unsafe replay after TLS has completed.
