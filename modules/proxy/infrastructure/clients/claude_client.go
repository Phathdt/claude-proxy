package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	sctx "github.com/phathdt/service-context"
)

// ClaudeAPIClient handles HTTP communication with Claude API using req
type ClaudeAPIClient struct {
	baseURL string
	client  *req.Client
	logger  sctx.Logger
}

// NewClaudeAPIClient creates a new Claude API client with req
func NewClaudeAPIClient(baseURL string, logger sctx.Logger) *ClaudeAPIClient {
	client := req.C().
		SetBaseURL(baseURL).
		SetTimeout(60*time.Second).
		SetCommonRetryCount(2).
		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetCommonHeaders(map[string]string{
			"Content-Type":      "application/json",
			"anthropic-version": "2023-06-01",
			"anthropic-beta":    "oauth-2025-04-20", // Required for OAuth authentication
		})

	c := &ClaudeAPIClient{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}

	// Add request/response logging middleware
	c.client.OnBeforeRequest(c.logRequest)
	c.client.OnAfterResponse(c.logResponse)

	return c
}

// logRequest logs the outgoing request to Claude API
func (c *ClaudeAPIClient) logRequest(client *req.Client, req *req.Request) error {
	fields := sctx.Fields{
		"method": req.Method,
		"url":    req.RawURL,
	}

	// Log headers for debugging authentication issues
	if len(req.Headers) > 0 {
		headerMap := make(map[string]string)
		for key, values := range req.Headers {
			if len(values) > 0 {
				headerMap[key] = values[0]
			}
		}
		fields["headers"] = headerMap
	}

	// Log body size if present
	if len(req.Body) > 0 {
		fields["body_size"] = len(req.Body)
		// For small bodies (< 500 bytes), log the actual content
		if len(req.Body) < 500 {
			fields["body_preview"] = string(req.Body)
		}
	}

	c.logger.Withs(fields).Debug("Sending request to Claude API")

	return nil
}

// logResponse logs the response from Claude API
func (c *ClaudeAPIClient) logResponse(client *req.Client, resp *req.Response) error {
	statusCode := resp.StatusCode
	contentType := resp.Header.Get("Content-Type")

	fields := sctx.Fields{
		"status_code":           statusCode,
		"content_type":          contentType,
		"request_method":        resp.Request.Method,
		"request_path":          resp.Request.URL.String(),
		"response_headers_size": len(resp.Header),
	}

	// Log response body (be careful with large responses)
	body := resp.String()
	if body != "" && len(body) < 10000 { // Log only if under 10KB to avoid huge logs
		fields["response_body"] = body
	} else if body != "" {
		fields["response_body_size"] = len(body)
	}

	// Log request body if available (from Response.Request)
	if len(resp.Request.Body) > 0 {
		if len(resp.Request.Body) < 10000 {
			fields["request_body"] = string(resp.Request.Body)
		} else {
			fields["request_body_size"] = len(resp.Request.Body)
		}
	}

	// Log based on status code
	switch {
	case statusCode >= 500:
		c.logger.Withs(fields).Error("Claude API request failed with 5xx error")
	case statusCode >= 400:
		c.logger.Withs(fields).Warn("Claude API request failed with 4xx error")
	case statusCode >= 300:
		c.logger.Withs(fields).Info("Claude API request redirected")
	default:
		c.logger.Withs(fields).Debug("Claude API request successful")
	}

	return nil
}

// ProxyRequest proxies an HTTP request to Claude API using req
func (c *ClaudeAPIClient) ProxyRequest(
	ctx context.Context,
	method, path string,
	accessToken string,
	body []byte,
) (*http.Response, error) {
	// Create req request with context
	// Common headers (Content-Type, Anthropic-Version, Anthropic-Beta) are already set
	// Only add the Authorization header which varies per request
	request := c.client.R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + accessToken,
		})

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
