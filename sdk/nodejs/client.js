const axios = require('axios');

/**
 * @typedef {Object} RateLimitResult
 * @property {boolean} allowed    - Whether the request passed rate limiting
 * @property {number}  remaining  - Requests remaining in the current window
 * @property {number}  limit      - Total requests allowed per window
 * @property {number}  resetIn    - Seconds until the window resets
 */

class LimiterClient {
  /**
   * @param {string} baseURL - Limiter.io gateway URL (e.g. "http://localhost:8080")
   * @param {string} apiKey  - Developer API key
   * @param {object} [opts]
   * @param {number} [opts.timeout=5000] - Request timeout in ms
   */
  constructor(baseURL, apiKey, opts = {}) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.client = axios.create({
      baseURL: baseURL,
      timeout: opts.timeout || 5000,
    });
  }

  /**
   * Check rate limits and return the full result including remaining quota.
   * @param {string} routePath - The route path to check (e.g. "/v1/users")
   * @returns {Promise<RateLimitResult>}
   */
  async check(routePath) {
    try {
      const url = `/api/v1/gateway${routePath}`;
      const response = await this.client.get(url, {
        headers: {
          'X-API-Key': this.apiKey,
          'Accept': 'application/json'
        }
      });

      return {
        allowed: true,
        remaining: parseInt(response.headers['x-ratelimit-remaining'] || '0', 10),
        limit: parseInt(response.headers['x-ratelimit-limit'] || '0', 10),
        resetIn: parseInt(response.headers['x-ratelimit-reset'] || '0', 10),
      };
    } catch (error) {
      if (error.response && error.response.status === 429) {
        return {
          allowed: false,
          remaining: 0,
          limit: parseInt(error.response.headers['x-ratelimit-limit'] || '0', 10),
          resetIn: parseInt(error.response.headers['x-ratelimit-reset'] || '0', 10),
        };
      }
      throw error;
    }
  }

  /**
   * Simple boolean check. Use check() for full quota details.
   * @param {string} routePath
   * @returns {Promise<boolean>}
   */
  async verify(routePath) {
    const result = await this.check(routePath);
    return result.allowed;
  }
}

module.exports = LimiterClient;
