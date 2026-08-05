# Releasing

## Before the first public tag

1. Verify the `org.jissr` namespace in Maven Central using the owned domain.
2. Enable GitHub private vulnerability reporting.
3. Create code-signing, Maven Central, and artifact-signing credentials in the
   public repository's CI secret store.

The canonical Go module path is `github.com/inqadh/jissr-bypass`. Published Go
module versions are immutable, so it must not be changed after the first tag.

## Candidate

```sh
go mod tidy
go test ./...
```

Build Android on a host with a JDK and Android SDK:

```powershell
.\scripts\build-android.ps1
```

Build Apple artifacts on macOS with Xcode:

```sh
./scripts/build-apple.sh
```

Build the desktop matrix:

```powershell
.\scripts\build-desktop.ps1 -Version 0.1.0
```

On Windows, the candidate orchestrator runs tests, vet, the Android binding,
desktop cross-builds, and writes `dist/SHA256SUMS`:

```powershell
.\scripts\release.ps1 -Version 0.1.0
```

Validate the ordinary health endpoint and the ECH canary from at least two
independent networks. A successful health request alone does not validate ECH;
the CLI must verify the versioned nonce challenge.

## Publish

1. Commit the exact source tree used for all artifacts.
2. Create a signed semantic-version tag.
3. Build artifacts from the tag.
4. Generate SHA-256 checksums and an SPDX or CycloneDX SBOM.
5. Publish the Go tag and request it through `proxy.golang.org`.
6. Publish the signed Android artifact to Maven Central.
7. Upload the XCFramework ZIP and update its Swift Package checksum.
8. Upload signed desktop archives, checksums, and SBOMs.
9. Re-run the public canary examples from a clean consumer project.

## Maven Central

Android is published as:

```text
org.jissr:jissr-bypass-android:<version>
```

Maven Central components are immutable. Run the `Maven Central` workflow
manually with a tag and the same version without the `v` prefix. Keep `upload`
disabled to build and inspect the signed bundle only.

When `upload` is enabled, the workflow submits a `USER_MANAGED` deployment. It
waits for Central validation but does not make the component public. Inspect
the validated deployment in the Central Portal before selecting **Publish**.

The protected `maven-central` GitHub environment requires:

- `MAVEN_SIGNING_KEY_B64`: base64-encoded ASCII-armored PGP private key;
- `MAVEN_SIGNING_PASSWORD`: private-key passphrase, if one is set;
- `CENTRAL_TOKEN_USERNAME`: Central Portal user-token username; and
- `CENTRAL_TOKEN_PASSWORD`: Central Portal user-token password.

The `org.jissr` namespace must be verified against `jissr.org` before upload.
