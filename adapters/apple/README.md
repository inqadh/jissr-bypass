# Jissr Bypass for URLSession

The Apple adapter recreates one central `URLSession` while preserving existing
`URLRequest` construction:

```swift
let session = try JissrBypassURLSession.patch(
    existingConfiguration,
    protectedHosts: ["api.example.com"]
)
```

Only that session is affected. The adapter does not use global
`URLProtocol.registerClass`, and exact host matching avoids unintentionally
routing sibling subdomains.

The first release supports bounded data requests. Background sessions,
WebSockets, streamed uploads, streamed downloads, and `WKWebView` interception
are outside its compatibility contract.

The local package manifest expects the generated XCFramework at
`Binaries/JissrBypass.xcframework`. Release automation supplies that directory
before running Swift tests. Public Swift Package Manager distribution uses a
small organization-owned wrapper repository with a checksum-pinned release
artifact.
