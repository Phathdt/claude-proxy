package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"claude-proxy/config"
	"claude-proxy/pkg/account"
	"claude-proxy/pkg/claude"
	"claude-proxy/pkg/errors"
	"claude-proxy/pkg/handlers"
	"claude-proxy/pkg/oauth"
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

// CloveProviders provides Clove-specific domain providers
var CloveProviders = fx.Options(
	fx.Provide(
		// OAuth service
		NewOAuthService,
		// Account manager with token refresher
		NewAccountManager,
		// Multi-account manager
		NewMultiAccountManager,
		// Claude API client
		NewClaudeClient,
		// Handlers
		NewOAuthHandler,
		NewMessagesHandler,
		NewHealthHandler,
		NewAuthHandler,
		NewAppAccountHandler,
		// Telegram client (optional)
		NewTelegramClient,
	),
)

// APIProviders provides all dependencies needed for the API service
var APIProviders = fx.Options(
	CoreProviders,
	CloveProviders,
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
			Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
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

// NewOAuthService creates a new OAuth service
func NewOAuthService(cfg *config.Config) *oauth.Service {
	return oauth.NewService(
		cfg.OAuth.ClientID,
		cfg.OAuth.AuthorizeURL,
		cfg.OAuth.TokenURL,
		cfg.OAuth.RedirectURI,
	)
}

// oauthRefreshAdapter adapts OAuth service to account.TokenRefresher interface
type oauthRefreshAdapter struct {
	oauthService *oauth.Service
}

func (a *oauthRefreshAdapter) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, int, error) {
	tokenResp, err := a.oauthService.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return "", "", 0, err
	}
	return tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn, nil
}

// NewAccountManager creates a new account manager
func NewAccountManager(cfg *config.Config, oauthService *oauth.Service, appLogger sctx.Logger) (*account.Manager, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "account-manager"})

	// Create adapter for token refresh
	refresher := &oauthRefreshAdapter{oauthService: oauthService}

	// Create account manager
	manager := account.NewManager(cfg.Storage.DataFolder, refresher)

	// Initialize (create data folder and load existing account)
	if err := manager.Initialize(); err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to initialize account manager")
		return nil, fmt.Errorf("failed to initialize account manager: %w", err)
	}

	logger.Info("Account manager initialized successfully")
	return manager, nil
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(cfg *config.Config) *claude.Client {
	return claude.NewClient(cfg.Claude.BaseURL)
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *oauth.Service, accountManager *account.Manager, cfg *config.Config) *handlers.OAuthHandler {
	return handlers.NewOAuthHandler(oauthService, accountManager, cfg.Claude.BaseURL)
}

// NewMessagesHandler creates a new messages handler
func NewMessagesHandler(claudeClient *claude.Client, accountManager *account.Manager, cfg *config.Config) *handlers.MessagesHandler {
	return handlers.NewMessagesHandler(claudeClient, accountManager, cfg)
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(accountManager *account.Manager) *handlers.HealthHandler {
	return handlers.NewHealthHandler(accountManager)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config) *handlers.AuthHandler {
	return handlers.NewAuthHandler(cfg)
}

// NewMultiAccountManager creates a new multi-account manager
func NewMultiAccountManager(cfg *config.Config, oauthService *oauth.Service, appLogger sctx.Logger) (*account.MultiAccountManager, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "multi-account-manager"})

	// Create adapter for token refresh
	refresher := &oauthRefreshAdapter{oauthService: oauthService}

	// Create multi-account manager
	manager := account.NewMultiAccountManager(cfg.Storage.DataFolder, refresher)

	// Initialize (create data folder and load existing accounts)
	if err := manager.Initialize(); err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to initialize multi-account manager")
		return nil, fmt.Errorf("failed to initialize multi-account manager: %w", err)
	}

	logger.Info("Multi-account manager initialized successfully")
	return manager, nil
}

// NewAppAccountHandler creates a new app account handler
func NewAppAccountHandler(oauthService *oauth.Service, multiAcctMgr *account.MultiAccountManager, cfg *config.Config) *handlers.AppAccountHandler {
	return handlers.NewAppAccountHandler(oauthService, multiAcctMgr, cfg.Claude.BaseURL)
}
