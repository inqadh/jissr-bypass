const PROTOCOL = "jissr-bypass/1";
const HOST = "blocked.yusef.ren";

function securityHeaders(contentType: string): Headers {
	const headers = new Headers({
		"Content-Type": contentType,
		"Cache-Control": "no-store, max-age=0",
		"Pragma": "no-cache",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy": "no-referrer",
		"Content-Security-Policy": "default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'",
	});
	return headers;
}

function json(value: unknown, status = 200): Response {
	return new Response(JSON.stringify(value), {
		status,
		headers: securityHeaders("application/json; charset=utf-8"),
	});
}

function canary(request: Request, url: URL): Response {
	if (request.method !== "GET") {
		return json({ error: "method_not_allowed" }, 405);
	}
	const nonce = url.searchParams.get("nonce") ?? "";
	if (!/^[A-Za-z0-9_-]{16,128}$/.test(nonce)) {
		return json({ error: "invalid_nonce" }, 400);
	}
	return json({
		ok: true,
		protocol: PROTOCOL,
		host: HOST,
		nonce,
		issued_at: new Date().toISOString().replace(/\.\d{3}Z$/, "Z"),
	});
}

function page(): Response {
	const html = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Jissr Bypass connectivity canary</title>
<style>
:root{color-scheme:dark;font-family:system-ui,sans-serif;background:#081018;color:#e8f1f8}
main{max-width:720px;margin:12vh auto;padding:32px}
.status{display:inline-block;padding:6px 12px;border:1px solid #36d399;border-radius:999px;color:#36d399}
code{color:#8bd5ff}p{line-height:1.65;color:#b9cad8}a{color:#8bd5ff}
</style>
</head>
<body><main>
<div class="status">Canary online</div>
<h1>Jissr Bypass</h1>
<p>This hostname is a controlled connectivity canary. Native applications can use it to distinguish an ordinary HTTPS failure from a successful TLS 1.3 Encrypted Client Hello connection.</p>
<p>Protocol: <code>${PROTOCOL}</code></p>
<p>No account, cookie, fingerprint, or request body is required by this service.</p>
</main></body></html>`;
	return new Response(html, {
		headers: securityHeaders("text/html; charset=utf-8"),
	});
}

export default {
	fetch(request: Request): Response {
		const url = new URL(request.url);
		if (url.hostname !== HOST) {
			return json({ error: "unknown_host" }, 421);
		}
		if (url.pathname === "/v1/canary") return canary(request, url);
		if (url.pathname === "/v1/health" && (request.method === "GET" || request.method === "HEAD")) {
			return new Response(request.method === "HEAD" ? null : "ok\n", {
				headers: securityHeaders("text/plain; charset=utf-8"),
			});
		}
		if (url.pathname === "/" && request.method === "GET") return page();
		return json({ error: "not_found" }, 404);
	},
};
