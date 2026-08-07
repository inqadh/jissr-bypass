# Changelog

All notable changes to Jissr Bypass are documented here.

The project follows Semantic Versioning. Versions containing `-rc` are release
candidates and may change before the stable release.

## Unreleased

## 0.1.1 - 2026-08-07

- Added reusable HTTP/2 transports, per-host DNS/ECH caching, and concurrent
  resolution deduplication.
- Added an opt-in native-first client with a bounded 750 ms route check,
  safe fallback before request transmission, and per-host route caching.
- Added explicit network-state reset and idle-connection cleanup APIs for
  mobile applications.
- Added structured transport events for native attempts, likely interference,
  fallback activation, route selection, resolution, and network lifecycle.
- Added an application-configurable native-attempt timeout to the mobile API.
- Added deterministic Go benchmarks and an Android instrumentation benchmark
  against the public connectivity canary.
- Added exact-host routing for existing Go `http.Client` configurations.
- Added an optional OkHttp/Retrofit adapter with bounded body handling.
- Added an optional URLSession adapter for centrally recreated Apple sessions.
- Added the checksum-pinned Swift package wrapper preparation pipeline.
- Added the negotiated HTTP protocol to mobile response bindings.
- Added complete Android installation, API, diagnostics, shrinking, and
  release-verification documentation.
- Added generated HTML API reference to the Maven Javadoc artifact.
- Added organization and issue-tracker metadata to the Maven POM.

## 0.1.0-rc.2

- Published the Android AAR to Maven Central as
  `org.jissr:jissr-bypass-android:0.1.0-rc.2`.
- Added signed Maven sources, Javadoc, POM, and checksum artifacts.
- Added organization-owned POM and signing identities.

## 0.1.0-rc.1

- First multi-platform release candidate.
