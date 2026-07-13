import requests
from dataclasses import dataclass


@dataclass
class RateLimitResult:
    """Full rate-limit evaluation result."""
    allowed: bool       # Whether the request passed rate limiting
    remaining: int      # Requests remaining in the current window
    limit: int          # Total requests allowed per window
    reset_in: int       # Seconds until the window resets


class LimiterClient:
    """Official Limiter.io Python SDK client."""

    def __init__(self, base_url: str, api_key: str, timeout: float = 5.0):
        self.base_url = base_url
        self.api_key = api_key
        self.timeout = timeout

    def check(self, route_path: str) -> RateLimitResult:
        """Check rate limits and return the full result including remaining quota."""
        url = f"{self.base_url}/api/v1/gateway{route_path}"
        try:
            res = requests.get(
                url,
                headers={
                    "X-API-Key": self.api_key,
                    "Accept": "application/json"
                },
                timeout=self.timeout
            )

            return RateLimitResult(
                allowed=res.status_code == 200,
                remaining=int(res.headers.get("X-RateLimit-Remaining", 0)),
                limit=int(res.headers.get("X-RateLimit-Limit", 0)),
                reset_in=int(res.headers.get("X-RateLimit-Reset", 0)),
            )
        except requests.RequestException as e:
            raise RuntimeError(f"Rate Limiter connection failed: {e}")

    def verify(self, route_path: str) -> bool:
        """Simple boolean check. Use check() for full quota details."""
        result = self.check(route_path)
        return result.allowed
