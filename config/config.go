package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database      DatabaseConfig      `yaml:"database"       mapstructure:"database"`
	Server        ServerConfig        `yaml:"server"         mapstructure:"server"`
	Logger        LoggerConfig        `yaml:"logger"         mapstructure:"logger"`
	Telegram      TelegramConfig      `yaml:"telegram"       mapstructure:"telegram"`
	WalletChecker WalletCheckerConfig `yaml:"wallet_checker" mapstructure:"wallet_checker"`
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

type DatabaseConfig struct {
	URI string `yaml:"uri" mapstructure:"uri"`
}

type ServerConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
}

// WalletCheckerConfig holds the configuration for wallet checker module
type WalletCheckerConfig struct {
	Enabled     bool                     `yaml:"enabled"     mapstructure:"enabled"`
	Chainalysis ChainalysisConfig        `yaml:"chainalysis" mapstructure:"chainalysis"`
	Cache       WalletCheckerCacheConfig `yaml:"cache"       mapstructure:"cache"`
}

// ChainalysisConfig holds the configuration for Chainalysis API integration
type ChainalysisConfig struct {
	Enabled   bool            `yaml:"enabled"    mapstructure:"enabled"`
	APIKey    string          `yaml:"api_key"    mapstructure:"api_key"`
	BaseURL   string          `yaml:"base_url"   mapstructure:"base_url"`
	Timeout   time.Duration   `yaml:"timeout"    mapstructure:"timeout"`
	RateLimit RateLimitConfig `yaml:"rate_limit" mapstructure:"rate_limit"`
}

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	DailyLimit       int           `yaml:"daily_limit"        mapstructure:"daily_limit"`
	SyncInterval     time.Duration `yaml:"sync_interval"      mapstructure:"sync_interval"`
	Timezone         string        `yaml:"timezone"           mapstructure:"timezone"`
	BackupOnShutdown bool          `yaml:"backup_on_shutdown" mapstructure:"backup_on_shutdown"`
}

// WalletCheckerCacheConfig holds cache configuration for wallet checker
type WalletCheckerCacheConfig struct {
	Enabled              bool          `yaml:"enabled"                mapstructure:"enabled"`
	NormalTTLHours       int           `yaml:"normal_ttl_hours"       mapstructure:"normal_ttl_hours"`
	CleanupIntervalHours int           `yaml:"cleanup_interval_hours" mapstructure:"cleanup_interval_hours"`
	MaxSize              int64         `yaml:"max_size"               mapstructure:"max_size"`
	NumCounters          int64         `yaml:"num_counters"           mapstructure:"num_counters"`
	BufferItems          int64         `yaml:"buffer_items"           mapstructure:"buffer_items"`
	DefaultTTL           time.Duration `yaml:"default_ttl"            mapstructure:"default_ttl"`
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

	return &config, nil
}
