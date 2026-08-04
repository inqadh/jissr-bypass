# Security policy

Please do not publish suspected vulnerabilities in a public issue before a fix
is available.

For the initial private-development phase, report findings to the repository
maintainers through the private repository security channel. Before the first
public release, enable GitHub private vulnerability reporting and replace this
paragraph with the organization's dedicated security contact.

## Supported versions

Until `v1.0.0`, only the latest tagged pre-release receives security fixes.
After `v1.0.0`, the project will publish a version-support table here.

## Release requirements

Every public release must:

- pass unit, cross-build, Android binding, and Apple binding checks;
- contain no credentials or service secrets;
- publish SHA-256 checksums and an SBOM;
- use an immutable semantic-version tag; and
- document security-relevant behavior changes.
