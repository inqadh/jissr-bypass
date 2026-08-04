# Contributing

Run the complete Go test suite before submitting a change:

```sh
go test ./...
```

Changes to DNS parsing must include fixtures for both wire-format data and the
supported JSON fallback representations. Changes to transport behavior must
preserve TLS certificate validation, TLS 1.3 minimums, ECH configuration
validation, bounded response sizes, and redirect isolation.

Generated AARs, XCFrameworks, executables, credentials, and local Worker files
must not be committed.
