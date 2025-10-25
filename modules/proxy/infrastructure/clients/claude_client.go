package clients

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// ClaudeAPIClient handles HTTP communication with Claude API
type ClaudeAPIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClaudeAPIClient creates a new Claude API client
func NewClaudeAPIClient(baseURL string) *ClaudeAPIClient {
	return &ClaudeAPIClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// ProxyRequest proxies an HTTP request to Claude API
func (c *ClaudeAPIClient) ProxyRequest(ctx context.Context, method, path string, headers map[string]string, body []byte) (*http.Response, error) {
	// Build target URL
	targetURL := c.baseURL + path

	// Create request
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default headers if not present
	if req.Header.Get("Content-Type") == "" && len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	if req.Header.Get("anthropic-version") == "" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	// Send request
	return c.httpClient.Do(req)
}
