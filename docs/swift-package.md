# Swift Package Manager distribution

Apple distribution uses two organization-owned repositories:

- `inqadh/jissr-bypass` builds and publishes the immutable XCFramework ZIP.
- `inqadh/jissr-bypass-swift` contains the checksum-pinned Swift package
  manifest and URLSession adapter.

Separating the lightweight wrapper avoids committing a large generated binary
to source control and allows its manifest to pin the exact bytes already
published by the core release.

## Release sequence

1. Tag and publish the core release with
   `JissrBypass.xcframework.zip`.
2. Download that exact release asset.
3. Generate the wrapper source:

   ```powershell
   .\scripts\prepare-swift-package.ps1 `
       -Version 0.1.0-rc.3 `
       -XCFrameworkZip .\JissrBypass.xcframework.zip `
       -Output .\dist\swift-package
   ```

4. Commit the generated source to `inqadh/jissr-bypass-swift`.
5. Run `swift test` on macOS.
6. Tag the wrapper with the same version.

Developers then add:

```text
https://github.com/inqadh/jissr-bypass-swift
```

The package exposes both the raw `JissrBypass` binary module and the
`JissrBypassURLSession` adapter.

Never calculate the checksum from a locally rebuilt archive when publishing
the wrapper. Download the exact GitHub release asset so the SwiftPM checksum
matches the public bytes.
