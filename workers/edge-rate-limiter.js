// Cloudflare Edge Worker for distributed rate limiting
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const clientIp = request.headers.get('CF-Connecting-IP') || 'unknown';
    const route = url.pathname;
    const key = `rl:${clientIp}:${route}`;

    // In KV-based rate limiting
    const { value } = await env.KV_NAMESPACE.getWithMetadata(key);
    const now = Date.now();
    const windowMs = 60000;
    const maxReq = 100;

    let current = 0;
    let windowStart = now;

    if (value) {
      current = value.count;
      windowStart = value.windowStart;
      if (now - windowStart > windowMs) {
        current = 0;
        windowStart = now;
      }
    }

    current++;
    await env.KV_NAMESPACE.put(key, JSON.stringify({ count: current, windowStart }), {
      expirationTtl: Math.ceil(windowMs / 1000) * 2,
    });

    if (current > maxReq) {
      return new Response(
        JSON.stringify({ error: 'rate_limit_exceeded', retry_after: Math.ceil((windowMs - (now - windowStart)) / 1000) }),
        { status: 429, headers: { 'Content-Type': 'application/json', 'Retry-After': Math.ceil((windowMs - (now - windowStart)) / 1000).toString() } }
      );
    }

    // Proxy upstream
    const upstreamUrl = `https://api.limiter.io${url.pathname}${url.search}`;
    const upstreamRequest = new Request(upstreamUrl, request);
    upstreamRequest.headers.set('X-Edge-Processed', 'true');
    upstreamRequest.headers.set('X-Real-IP', clientIp);

    const upstreamResponse = await fetch(upstreamRequest);
    const responseHeaders = new Headers(upstreamResponse.headers);
    responseHeaders.set('X-RateLimit-Limit', maxReq.toString());
    responseHeaders.set('X-RateLimit-Remaining', Math.max(0, maxReq - current).toString());
    responseHeaders.set('X-Edge-Region', env.CF_REGION || 'unknown');

    return new Response(upstreamResponse.body, {
      status: upstreamResponse.status,
      headers: responseHeaders,
    });
  },
};

export const config = {
  name: 'edge-rate-limiter',
  compatibility_date: '2024-01-01',
};
