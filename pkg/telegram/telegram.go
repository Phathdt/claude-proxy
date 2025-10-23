package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sctx "github.com/phathdt/service-context"
)

// Config holds Telegram bot configuration
type Config struct {
	Enabled  bool          `mapstructure:"enabled"`
	BotToken string        `mapstructure:"bot_token"`
	ChatID   string        `mapstructure:"chat_id"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// Client represents a Telegram bot client
type Client struct {
	config     Config
	httpClient *http.Client
	logger     sctx.Logger
}

// telegramMessage represents the Telegram sendMessage API payload
type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// telegramResponse represents the Telegram API response
type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// NewClient creates a new Telegram client
func NewClient(config Config, logger sctx.Logger) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger.Withs(sctx.Fields{"component": "telegram-client"}),
	}
}

// SendMessage sends a text message to the configured chat
func (c *Client) SendMessage(ctx context.Context, message string) error {
	if !c.config.Enabled {
		c.logger.Debug("Telegram notifications disabled, skipping message")
		return nil
	}

	if c.config.BotToken == "" || c.config.ChatID == "" {
		return fmt.Errorf("telegram bot_token or chat_id not configured")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.config.BotToken)

	payload := telegramMessage{
		ChatID:    c.config.ChatID,
		Text:      message,
		ParseMode: "Markdown",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Withs(sctx.Fields{
			"error": err.Error(),
		}).Error("Failed to send telegram message")
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	var telegramResp telegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return fmt.Errorf("failed to decode telegram response: %w", err)
	}

	if !telegramResp.OK {
		c.logger.Withs(sctx.Fields{
			"status_code": resp.StatusCode,
			"description": telegramResp.Description,
		}).Error("Telegram API returned error")
		return fmt.Errorf("telegram API error: %s", telegramResp.Description)
	}

	c.logger.Withs(sctx.Fields{
		"chat_id": c.config.ChatID,
	}).Debug("Telegram message sent successfully")

	return nil
}

// SendMarkdownMessage sends a formatted markdown message
func (c *Client) SendMarkdownMessage(ctx context.Context, title, message string) error {
	formattedMessage := fmt.Sprintf("*%s*\n\n%s", title, message)
	return c.SendMessage(ctx, formattedMessage)
}

// IsEnabled returns whether Telegram notifications are enabled
func (c *Client) IsEnabled() bool {
	return c.config.Enabled
}
