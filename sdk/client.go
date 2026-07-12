package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type VerificationResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Path    string `json:"path"`
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

// Verify sends a request to the rate-limiter gateway.
// Returns (allowed = true) if request passes rate limiting, otherwise returns false.
func (c *Client) Verify(ctx context.Context, routePath string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/gateway%s", c.baseURL, routePath)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request to rate limiter failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return false, nil
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		return false, fmt.Errorf("rate limiter returned error (%d): %s", resp.StatusCode, errResp.Error)
	}

	return false, fmt.Errorf("rate limiter returned status code %d", resp.StatusCode)
}
