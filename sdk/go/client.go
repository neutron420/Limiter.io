package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RateLimitResult contains the full rate-limit evaluation response.
type RateLimitResult struct {
	Allowed   bool          // Whether the request passed rate limiting
	Remaining int           // Number of requests remaining in the current window
	Limit     int           // Total requests allowed per window
	Reset     time.Duration // Time until the rate limit window resets
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewClient(baseURL string, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Check sends a request to the rate-limiter gateway and returns the full result
// including remaining quota and reset time.
func (c *Client) Check(ctx context.Context, routePath string) (*RateLimitResult, error) {
	url := fmt.Sprintf("%s/api/v1/gateway%s", c.baseURL, routePath)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to rate limiter failed: %w", err)
	}
	defer resp.Body.Close()

	result := &RateLimitResult{
		Allowed: resp.StatusCode == http.StatusOK,
	}

	// Parse standard rate-limit headers
	if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
		result.Remaining, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
		result.Limit, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
		sec, _ := strconv.Atoi(v)
		result.Reset = time.Duration(sec) * time.Second
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusTooManyRequests {
		return result, nil
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		return result, fmt.Errorf("rate limiter returned error (%d): %s", resp.StatusCode, errResp.Error)
	}

	return result, fmt.Errorf("rate limiter returned status code %d", resp.StatusCode)
}

// Verify is a convenience method that returns a simple bool.
// Use Check() if you need remaining quota and reset time.
func (c *Client) Verify(ctx context.Context, routePath string) (bool, error) {
	res, err := c.Check(ctx, routePath)
	if err != nil {
		return false, err
	}
	return res.Allowed, nil
}
