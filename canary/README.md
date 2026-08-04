# Jissr Bypass canary

This Worker is the controlled origin for `blocked.yusef.ren`.

It deliberately has no database, cookies, analytics, account system, or
dependency on the Jisr control plane. Keeping the canary isolated makes a
successful response a narrow test of DNS-over-HTTPS, ECH, TLS, and HTTP.

## Endpoints

- `GET /` — explanatory page for people.
- `GET /v1/health` — ordinary monitoring.
- `GET /v1/canary?nonce=<base64url>` — versioned SDK challenge.

The SDK accepts a canary response only when its protocol, hostname, freshness,
and unpredictable echoed nonce all match.

## Deploy

Authenticate Wrangler with a Cloudflare token that can edit Workers and the
`yusef.ren` zone, then run:

```sh
npm ci
npm run check
npm run deploy
```

The Custom Domain route creates the proxied DNS record and certificate. Do not
create a separate CNAME at the same hostname.
