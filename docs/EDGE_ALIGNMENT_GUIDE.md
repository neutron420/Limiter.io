# Edge Alignment Guide (Making Limiter.io behave like Cloudflare)

This guide shows you how to integrate **Limiter.io** with **Cloudflare Workers** at the network edge to achieve global Anycast proxy speed, sub-5ms decision latency, and shield your origin servers from heavy traffic.

---

## 🌩️ 1. Architecture Overview

By wrapping Limiter.io's API with an Edge Reverse Proxy, requests are checked at the network boundary before hitting your origin server:

```
[ User Request ]
       │
       ▼
[ Cloudflare Edge Node ] (Anycast DNS)
       │
       ├─► [ Cloudflare Worker ] (Edge Throttler)
       │         │
       │         ├──► [ Global Upstash Redis ] (Decides in <2ms)
       │         │
       │         └─── (Blocked? Returns HTTP 429 immediately)
       │
       ▼ (Allowed)
[ Your Go Origin Server ] (Executes Core Business Logic)
```

---

## 💻 2. Cloudflare Worker Script (Edge Interceptor)

Deploy this script in your Cloudflare dashboard to intercept routes and check rate limits globally.

Create a file `edge-worker.js`:

```javascript
// edge-worker.js
// Deployed on Cloudflare Workers edge nodes globally

const API_GATEWAY_URL = "https://api.yourdomain.com/api/v1";
const GLOBAL_REDIS_REST_URL = "https://your-upstash-redis-rest-url.upstash.io";
const REDIS_TOKEN = "your_upstash_redis_rest_token";

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const apiKey = request.headers.get("X-API-Key");

    if (!apiKey) {
      return new Response("Missing X-API-Key header", { status: 401 });
    }

    // 1. Edge rate-limit cache check (Sub-2ms check)
    // Query Upstash Redis REST API directly from the Edge
    const key = `rate_limit:${apiKey}:${url.pathname}`;
    const redisCheck = await fetch(`${GLOBAL_REDIS_REST_URL}/get/${key}`, {
      headers: { Authorization: `Bearer ${REDIS_TOKEN}` }
    });
    
    const result = await redisCheck.json();
    if (result && result.result) {
      const quota = JSON.parse(result.result);
      // If locally recorded as blocked, return 429 immediately at the edge
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

    // 2. Fallback / Sync with Central Limiter.io Gateway
    // If not cached, or cache has expired, fetch origin status
    try {
      const response = await fetch(`${API_GATEWAY_URL}/gateway${url.pathname}`, {
        method: request.method,
        headers: request.headers,
        body: request.body
      });

      // Parse and sync rate-limit headers to Edge cache in the background
      const limit = response.headers.get("X-RateLimit-Limit");
      const remaining = response.headers.get("X-RateLimit-Remaining");
      const resetSec = response.headers.get("X-RateLimit-Reset");

      if (remaining && resetSec) {
        const quotaInfo = {
          remaining: parseInt(remaining),
          resetAt: Date.now() + (parseInt(resetSec) * 1000)
        };
        // Async update in edge cache (doesn't block client response)
        ctx.waitUntil(
          fetch(`${GLOBAL_REDIS_REST_URL}/set/${key}/${JSON.stringify(quotaInfo)}/EX/${resetSec}`, {
            headers: { Authorization: `Bearer ${REDIS_TOKEN}` }
          })
        );
      }

      return response;
    } catch (err) {
      // Fail-Open Policy: if gateway fails, allow edge request to origin
      console.error("Central limiter unreachable, failing open:", err);
      return fetch(request);
    }
  }
};
```

---

## 🗃️ 3. Global Geodistributed Redis Setup

To sync the rate counters across Cloudflare's 300+ global edge locations:

1. **Deploy Upstash Redis**:
   Create a database in Upstash Console and choose the **Global Database** setting (automatically replicates read/writes across primary edge nodes in US, Europe, and Asia).
2. **Configure Go Backend**:
   Map your backend Redis client to the Global Upstash replica URL. Now, when your Go server increments request logs via Lua, the counts are updated globally at all edge nodes within milliseconds.
