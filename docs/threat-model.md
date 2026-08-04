# Threat model

## Goals

Jissr Bypass is designed to:

- avoid dependence on the operating system's DNS resolver;
- keep the origin hostname out of the plaintext TLS ClientHello when the
  origin publishes a valid ECH configuration;
- reject malformed, mixed-provider, stale, or incomplete resolver answers;
- preserve normal Web PKI hostname and certificate verification; and
- give applications structured failure stages without logging response data.

## Trust

The client trusts:

- the configured DNS-over-HTTPS providers for DNS answers;
- the public certificate authorities accepted by the Go runtime;
- the destination origin and its ECH configuration; and
- the host application to protect data before and after the request.

Resolver diversity improves availability but does not provide consensus. The
first provider returning a mutually consistent address and ECH record wins.

## Defended conditions

- Local DNS interference or unusable local DNS.
- Observation or filtering of a plaintext TLS ClientHello hostname.
- Captive portals and middleboxes returning HTML in place of DoH.
- Provider differences between raw and presentation-form JSON DNS responses.
- Cross-host redirects that could accidentally leave the protected transport.
- Stale or cached canary responses.

## Out of scope

- Destination-IP blocking.
- Traffic-analysis resistance.
- Malware or a compromised host application.
- Compromised public certificate authorities.
- Compromise of every configured resolver.
- General-purpose tunnelling or arbitrary non-HTTP protocols.

## Privacy

The library does not include analytics, identifiers, telemetry, or persistent
storage. Applications decide what they send to destination origins. Resolver
operators can observe queried hostnames, so integrators should select providers
whose privacy terms match their requirements.
