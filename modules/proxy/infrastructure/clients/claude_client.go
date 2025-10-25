package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
)

// ClaudeAPIClient handles HTTP communication with Claude API using req
type ClaudeAPIClient struct {
	baseURL string
	client  *req.Client
}

// NewClaudeAPIClient creates a new Claude API client with req
func NewClaudeAPIClient(baseURL string) *ClaudeAPIClient {
	client := req.C().
		SetBaseURL(baseURL).
		SetTimeout(60 * time.Second).
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetCommonHeaders(map[string]string{
			"Content-Type":      "application/json",
			"anthropic-version": "2023-06-01",
			"anthropic-beta":    "oauth-2025-04-20", // Required for OAuth authentication
		})

	return &ClaudeAPIClient{
		baseURL: baseURL,
		client:  client,
	}
}

// ProxyRequest proxies an HTTP request to Claude API using req
func (c *ClaudeAPIClient) ProxyRequest(ctx context.Context, method, path string, headers map[string]string, body []byte) (*http.Response, error) {
	// Create req request with context
	request := c.client.R().
		SetContext(ctx).
		SetHeaders(headers)

	// Set body if present
	if len(body) > 0 {
		request.SetBodyBytes(body)
	}

	// Execute request based on method
	var resp *req.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = request.Get(path)
	case http.MethodPost:
		resp, err = request.Post(path)
	case http.MethodPut:
		resp, err = request.Put(path)
	case http.MethodPatch:
		resp, err = request.Patch(path)
	case http.MethodDelete:
		resp, err = request.Delete(path)
	case http.MethodHead:
		resp, err = request.Head(path)
	case http.MethodOptions:
		resp, err = request.Options(path)
	default:
		resp, err = request.Send(method, path)
	}

	if err != nil {
		return nil, err
	}

	// Return the underlying *http.Response
	// req automatically handles the response body properly for proxying
	return resp.Response, nil
}
