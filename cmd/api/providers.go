package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"claude-proxy/config"
	"claude-proxy/pkg/errors"
	"claude-proxy/pkg/telegram"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
	"go.uber.org/fx"
)

// CoreProviders provides all dependencies needed for the API service
var CoreProviders = fx.Options(
	fx.Provide(
		LoadConfig,
		func(cfg *config.Config) (sctx.ServiceContext, sctx.Logger, error) {
			return InitServiceContext(cfg)
		},
	),
)

// WalletCheckerProviders provides wallet checker domain providers
var WalletCheckerProviders = fx.Options(
	fx.Provide(
		// Telegram client
		NewTelegramClient,
	),
)

// APIProviders provides all dependencies needed for the API service
var APIProviders = fx.Options(
	CoreProviders,
	WalletCheckerProviders,
	fx.Provide(
		NewGinEngine,
	),
)

// LoadConfig loads configuration from the specified path
func LoadConfig(configPath string) (*config.Config, error) {
	return config.LoadConfig(configPath)
}

// InitServiceContext creates and loads the service context with database component and sets up global logger
func InitServiceContext(cfg *config.Config) (sctx.ServiceContext, sctx.Logger, error) {
	// Set up global logger first
	loggerConfig := &sctx.Config{
		DefaultLevel: cfg.Logger.Level,
		BasePrefix:   "claude-proxy",
		Format:       cfg.Logger.Format,
	}
	customLogger := sctx.NewAppLogger(loggerConfig)
	sctx.SetGlobalLogger(customLogger)

	// Create service context
	sc := sctx.NewServiceContext(
		sctx.WithName("claude-proxy"),
	)

	// Load all components
	if err := sc.Load(); err != nil {
		return nil, nil, fmt.Errorf("failed to load service context: %w", err)
	}

	return sc, sctx.GlobalLogger().GetLogger("main"), nil
}

// NewGinEngine creates a new Gin engine with middleware
func NewGinEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	engine.Use(ginLoggerMiddleware())

	engine.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger := sctx.GlobalLogger().GetLogger("gin")
		logger.Withs(sctx.Fields{"panic": recovered}).Error("PANIC RECOVERED")

		// Check if it's an AppError panic (our custom error handling pattern)
		if appErr, ok := recovered.(errors.AppError); ok {
			logger.Withs(sctx.Fields{
				"error_code":   appErr.ErrorCode(),
				"status_code":  appErr.StatusCode(),
				"error_detail": appErr.Details(),
			}).Debug("Handling custom app error panic")

			c.JSON(appErr.StatusCode(), gin.H{
				"code":    appErr.ErrorCode(),
				"message": appErr.Message(),
				"details": appErr.Details(),
			})
			c.Abort()
			return
		}

		// Handle other error types
		if err, ok := recovered.(error); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "An unexpected error occurred",
				"code":    "INTERNAL_SERVER_ERROR",
				"details": err.Error(),
			})
		} else if panicMsg, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "Application panic occurred",
				"code":    "PANIC_ERROR",
				"details": panicMsg,
			})
		} else {
			logger.Withs(sctx.Fields{"type": fmt.Sprintf("%T", recovered)}).Error("Unknown panic type")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "An unexpected error occurred",
				"code":    "UNKNOWN_ERROR",
			})
		}
		c.Abort()
	}))

	// CORS middleware - Allow all domains
	engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().
			Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Timeout middleware
	engine.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	return engine
}

// ginLoggerMiddleware creates a Gin middleware for structured logging
func ginLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		logger := sctx.GlobalLogger().GetLogger("gin")

		fields := sctx.Fields{
			"method":      method,
			"path":        path,
			"client_ip":   clientIP,
			"status_code": statusCode,
			"latency":     latency.String(),
			"user_agent":  c.Request.UserAgent(),
		}

		if errorMessage != "" {
			fields["error"] = errorMessage
		}

		switch {
		case statusCode >= 500:
			logger.Withs(fields).Error("HTTP Request")
		case statusCode >= 400:
			logger.Withs(fields).Warn("HTTP Request")
		default:
			logger.Withs(fields).Info("HTTP Request")
		}
	}
}

// NewTelegramClient creates a new Telegram client
func NewTelegramClient(cfg *config.Config, appLogger sctx.Logger) *telegram.Client {
	telegramConfig := telegram.Config{
		Enabled:  cfg.Telegram.Enabled,
		BotToken: cfg.Telegram.BotToken,
		ChatID:   cfg.Telegram.ChatID,
		Timeout:  cfg.Telegram.Timeout,
	}

	logger := appLogger.Withs(sctx.Fields{"component": "telegram-client"})
	return telegram.NewClient(telegramConfig, logger)
}
