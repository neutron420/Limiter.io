// Express middleware wrapper for Node.js SDK client

function expressRateLimit(client) {
  return async (req, res, next) => {
    try {
      const allowed = await client.verify(req.path);
      if (!allowed) {
        return res.status(429).json({
          error: 'Too Many Requests. Rate limit exceeded.'
        });
      }
      next();
    } catch (error) {
      // Fail-open strategy: allow requests on connection timeouts
      console.warn('[RateLimiter SDK] Gateway connection failed, failing open:', error.message);
      next();
    }
  };
}

module.exports = expressRateLimit;
