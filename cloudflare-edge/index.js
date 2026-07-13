// index.js
// Cloudflare Worker Edge proxy for Limiter.io

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const apiKey = request.headers.get("X-API-Key");

    if (!apiKey) {
      return new Response("Missing X-API-Key header", { status: 401 });
    }

    const key = `rate_limit:${apiKey}:${url.pathname}`;
    const globalRedisUrl = env.GLOBAL_REDIS_REST_URL;
    const redisToken = env.REDIS_TOKEN;

    // 1. Edge cache evaluation (Sub-2ms check)
    if (globalRedisUrl && redisToken) {
      try {
        const redisCheck = await fetch(`${globalRedisUrl}/get/${key}`, {
          headers: { Authorization: `Bearer ${redisToken}` }
        });
        
        const result = await redisCheck.json();
        if (result && result.result) {
          const quota = JSON.parse(result.result);
          if (quota.remaining <= 0 && Date.now() < quota.resetAt) {
            return new Response("Rate Limit Exceeded (Blocked at Edge)", {
              status: 429,
              headers: {
                "Retry-After": Math.ceil((quota.resetAt - Date.now()) / 1000).toString(),
                "X-RateLimit-Edge": "true"
              }
            });
          }
        }
      } catch (err) {
        console.warn("Failed to check edge cache:", err.message);
      }
    }

    // 2. Gateway Sync / Verification
    const gatewayUrl = env.API_GATEWAY_URL;
    if (!gatewayUrl) {
      return new Response("Edge Gateway URL configuration missing", { status: 500 });
    }

    try {
      const response = await fetch(`${gatewayUrl}/gateway${url.pathname}`, {
        method: request.method,
        headers: request.headers,
        body: request.body
      });

      const limit = response.headers.get("X-RateLimit-Limit");
      const remaining = response.headers.get("X-RateLimit-Remaining");
      const resetSec = response.headers.get("X-RateLimit-Reset");

      if (remaining && resetSec && globalRedisUrl && redisToken) {
        const quotaInfo = {
          remaining: parseInt(remaining),
          resetAt: Date.now() + (parseInt(resetSec) * 1000)
        };
        // Async background cache write
        ctx.waitUntil(
          fetch(`${globalRedisUrl}/set/${key}/${JSON.stringify(quotaInfo)}/EX/${resetSec}`, {
            headers: { Authorization: `Bearer ${redisToken}` }
          })
        );
      }

      return response;
    } catch (err) {
      // Fail-open: allow request if rate limiter gateway is down
      console.error("Central limiter unreachable, failing open:", err);
      return fetch(request);
    }
  }
};
