package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// ClaudeAPIClient handles HTTP communication with Claude API using Resty
type ClaudeAPIClient struct {
	baseURL string
	client  *resty.Client
}

// NewClaudeAPIClient creates a new Claude API client with Resty
func NewClaudeAPIClient(baseURL string) *ClaudeAPIClient {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(60 * time.Second)
	client.SetRetryCount(2)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	// Set default headers
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("anthropic-version", "2023-06-01")
	client.SetHeader("anthropic-beta", "oauth-2025-04-20") // Required for OAuth authentication

	return &ClaudeAPIClient{
		baseURL: baseURL,
		client:  client,
	}
}

// ProxyRequest proxies an HTTP request to Claude API using Resty
func (c *ClaudeAPIClient) ProxyRequest(ctx context.Context, method, path string, headers map[string]string, body []byte) (*http.Response, error) {
	// Create Resty request
	// IMPORTANT: Don't read response body automatically
	req := c.client.R().
		SetContext(ctx).
		SetDoNotParseResponse(true) // Keep raw response body

	// Set custom headers (these will override default headers)
	for key, value := range headers {
		req.SetHeader(key, value)
	}

	// Set body if present
	if len(body) > 0 {
		req.SetBody(body)
	}

	// Execute request based on method
	var resp *resty.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = req.Get(path)
	case http.MethodPost:
		resp, err = req.Post(path)
	case http.MethodPut:
		resp, err = req.Put(path)
	case http.MethodPatch:
		resp, err = req.Patch(path)
	case http.MethodDelete:
		resp, err = req.Delete(path)
	case http.MethodHead:
		resp, err = req.Head(path)
	case http.MethodOptions:
		resp, err = req.Options(path)
	default:
		resp, err = req.Execute(method, path)
	}

	if err != nil {
		return nil, err
	}

	// Return the underlying *http.Response with unconsumed body
	return resp.RawResponse, nil
}
