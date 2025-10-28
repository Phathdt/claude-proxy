package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"   mapstructure:"server"`
	Logger   LoggerConfig   `yaml:"logger"   mapstructure:"logger"`
	Auth     AuthConfig     `yaml:"auth"     mapstructure:"auth"`
	OAuth    OAuthConfig    `yaml:"oauth"    mapstructure:"oauth"`
	Claude   ClaudeConfig   `yaml:"claude"   mapstructure:"claude"`
	Storage  StorageConfig  `yaml:"storage"  mapstructure:"storage"`
	Retry    RetryConfig    `yaml:"retry"    mapstructure:"retry"`
	Session  SessionConfig  `yaml:"session"  mapstructure:"session"`
	Telegram TelegramConfig `yaml:"telegram" mapstructure:"telegram"`
}

type TelegramConfig struct {
	Enabled  bool          `yaml:"enabled"   mapstructure:"enabled"`
	BotToken string        `yaml:"bot_token" mapstructure:"bot_token"`
	ChatID   string        `yaml:"chat_id"   mapstructure:"chat_id"`
	Timeout  time.Duration `yaml:"timeout"   mapstructure:"timeout"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"  mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
}

type ServerConfig struct {
	Host           string        `yaml:"host"            mapstructure:"host"`
	Port           int           `yaml:"port"            mapstructure:"port"`
	RequestTimeout time.Duration `yaml:"request_timeout" mapstructure:"request_timeout"`
}

// AuthConfig holds API key authentication configuration
type AuthConfig struct {
	APIKey string `yaml:"api_key" mapstructure:"api_key"`
}

// OAuthConfig holds OAuth 2.0 configuration for Claude authentication
type OAuthConfig struct {
	ClientID     string `yaml:"client_id"     mapstructure:"client_id"`
	AuthorizeURL string `yaml:"authorize_url" mapstructure:"authorize_url"`
	TokenURL     string `yaml:"token_url"     mapstructure:"token_url"`
	RedirectURI  string `yaml:"redirect_uri"  mapstructure:"redirect_uri"`
	Scope        string `yaml:"scope"         mapstructure:"scope"`
}

// ClaudeConfig holds Claude API configuration
type ClaudeConfig struct {
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
}

// StorageConfig holds data storage configuration
type StorageConfig struct {
	DataFolder   string        `yaml:"data_folder"   mapstructure:"data_folder"`
	SyncInterval time.Duration `yaml:"sync_interval" mapstructure:"sync_interval"`
}

// RetryConfig holds retry logic configuration
type RetryConfig struct {
	MaxRetries int           `yaml:"max_retries" mapstructure:"max_retries"`
	RetryDelay time.Duration `yaml:"retry_delay" mapstructure:"retry_delay"`
}

// SessionConfig holds session limiting configuration (in-memory storage)
type SessionConfig struct {
	Enabled         bool          `yaml:"enabled"          mapstructure:"enabled"`
	MaxConcurrent   int           `yaml:"max_concurrent"   mapstructure:"max_concurrent"`
	SessionTTL      time.Duration `yaml:"session_ttl"      mapstructure:"session_ttl"`
	CleanupEnabled  bool          `yaml:"cleanup_enabled"  mapstructure:"cleanup_enabled"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" mapstructure:"cleanup_interval"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Load .env file if exists (optional)
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: failed to load .env file: %v\n", err)
	}

	// Configure viper for YAML file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Read YAML configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Configure environment variable support
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	var config Config

	// Unmarshal config with automatic env override
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default logger config if not specified
	if config.Logger.Level == "" {
		config.Logger.Level = "info"
	}
	if config.Logger.Format == "" {
		config.Logger.Format = "text"
	}

	// Set default OAuth config if not specified
	if config.OAuth.AuthorizeURL == "" {
		config.OAuth.AuthorizeURL = "https://claude.ai/oauth/authorize"
	}
	if config.OAuth.TokenURL == "" {
		config.OAuth.TokenURL = "https://api.claude.ai/oauth/token"
	}
	if config.OAuth.RedirectURI == "" {
		config.OAuth.RedirectURI = fmt.Sprintf("http://%s:%d/oauth/callback", config.Server.Host, config.Server.Port)
	}
	if config.OAuth.Scope == "" {
		config.OAuth.Scope = "user:profile user:inference"
	}

	// Set default Claude config if not specified
	if config.Claude.BaseURL == "" {
		config.Claude.BaseURL = "https://api.claude.ai"
	}

	// Set default storage config if not specified
	if config.Storage.DataFolder == "" {
		config.Storage.DataFolder = "~/.claude-proxy/data"
	}

	// Set default retry config if not specified
	if config.Retry.MaxRetries == 0 {
		config.Retry.MaxRetries = 3
	}
	if config.Retry.RetryDelay == 0 {
		config.Retry.RetryDelay = 1 * time.Second
	}

	// Set default server config if not specified
	if config.Server.RequestTimeout == 0 {
		config.Server.RequestTimeout = 5 * time.Minute // 5 minutes for LLM API requests
	}

	// Set default session config if not specified
	if config.Session.MaxConcurrent == 0 {
		config.Session.MaxConcurrent = 3
	}
	if config.Session.SessionTTL == 0 {
		config.Session.SessionTTL = 5 * time.Minute
	}
	if config.Session.CleanupInterval == 0 {
		config.Session.CleanupInterval = 1 * time.Minute
	}

	return &config, nil
}
