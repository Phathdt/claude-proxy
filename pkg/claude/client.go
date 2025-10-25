package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles communication with Claude API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// MessageRequest represents a Claude API message request
type MessageRequest struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	TopK        *int            `json:"top_k,omitempty"`
	System      string          `json:"system,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessageResponse represents a Claude API message response
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        *Usage         `json:"usage,omitempty"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewClient creates a new Claude API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
		},
	}
}

// SendMessage sends a message to Claude API
func (c *Client) SendMessage(ctx context.Context, accessToken string, req *MessageRequest) (*MessageResponse, error) {
	if req.Stream {
		return nil, fmt.Errorf("use SendMessageStream for streaming requests")
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/messages", c.baseURL), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode response
	var msgResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &msgResp, nil
}

// SendMessageStream sends a streaming message to Claude API
func (c *Client) SendMessageStream(ctx context.Context, accessToken string, req *MessageRequest) (<-chan StreamEvent, <-chan error) {
	eventChan := make(chan StreamEvent, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(eventChan)
		defer close(errChan)

		// Force stream to true
		req.Stream = true

		// Marshal request
		reqBody, err := json.Marshal(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/messages", c.baseURL), bytes.NewReader(reqBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		// Send request
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
			return
		}

		// Parse SSE stream
		if err := c.parseSSEStream(resp.Body, eventChan); err != nil {
			errChan <- err
			return
		}
	}()

	return eventChan, errChan
}

// parseSSEStream parses Server-Sent Events stream
func (c *Client) parseSSEStream(reader io.Reader, eventChan chan<- StreamEvent) error {
	scanner := bufio.NewScanner(reader)
	
	var eventType string
	var eventData strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if eventType != "" && eventData.Len() > 0 {
				event := StreamEvent{
					Event: eventType,
					Data:  eventData.String(),
				}
				eventChan <- event
				
				// Reset for next event
				eventType = ""
				eventData.Reset()
			}
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(strings.TrimSpace(line[5:]))
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// StreamEvent represents a single Server-Sent Event
type StreamEvent struct {
	Event string
	Data  string
}
