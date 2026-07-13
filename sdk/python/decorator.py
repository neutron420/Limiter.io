# Flask decorator wrapper for Python SDK client

from functools import wraps
import logging

def flask_rate_limit(client):
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            from flask import request, jsonify
            try:
                allowed = client.verify(request.path)
                if not allowed:
                    return jsonify({
                        "error": "Too Many Requests. Rate limit exceeded."
                    }), 429
            except Exception as e:
                # Fail-open strategy: allow requests if limiter gateway goes offline
                logging.warning(f"[RateLimiter SDK] Gateway check failed, failing open: {e}")
                
            return f(*args, **kwargs)
        return decorated_function
    return decorator
