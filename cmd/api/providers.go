package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"claude-proxy/cmd/api/handlers"
	"claude-proxy/config"
	"claude-proxy/modules/proxy/application/services"
	"claude-proxy/modules/proxy/domain/interfaces"
	"claude-proxy/modules/proxy/infrastructure/clients"
	"claude-proxy/modules/proxy/infrastructure/repositories"
	"claude-proxy/pkg/errors"
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
		// Infrastructure - Repositories
		NewTokenRepository,
		NewAccountRepository,
		// Infrastructure - Clients
		NewClaudeAPIClient,
		// Application - Services
		NewTokenService,
		NewAccountService,
		NewProxyService,
		// Handlers
		NewTokenHandler,
		NewProxyHandler,
		NewAuthHandler,
		NewAccountHandler,
		NewOAuthHandler,
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
		cfg.OAuth.Scope,
	)
}

// oauthRefreshAdapter adapts OAuth service to interfaces.TokenRefresher interface
type oauthRefreshAdapter struct {
	oauthService *oauth.Service
}

func (a *oauthRefreshAdapter) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (string, string, int, error) {
	tokenResp, err := a.oauthService.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return "", "", 0, err
	}
	return tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn, nil
}

// NewTokenRepository creates a new JSON token repository
func NewTokenRepository(cfg *config.Config, appLogger sctx.Logger) (interfaces.TokenRepository, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "token-repository"})

	repo, err := repositories.NewJSONTokenRepository(cfg.Storage.DataFolder)
	if err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to create token repository")
		return nil, fmt.Errorf("failed to create token repository: %w", err)
	}

	logger.Info("Token repository initialized successfully")
	return repo, nil
}

// NewAccountRepository creates a new JSON account repository
func NewAccountRepository(cfg *config.Config, appLogger sctx.Logger) (interfaces.AccountRepository, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "account-repository"})

	repo, err := repositories.NewJSONAccountRepository(cfg.Storage.DataFolder)
	if err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to create account repository")
		return nil, fmt.Errorf("failed to create account repository: %w", err)
	}

	logger.Info("Account repository initialized successfully")
	return repo, nil
}

// NewClaudeAPIClient creates a new Claude API client
func NewClaudeAPIClient(cfg *config.Config) *clients.ClaudeAPIClient {
	return clients.NewClaudeAPIClient(cfg.Claude.BaseURL)
}

// NewTokenService creates a new token service
func NewTokenService(tokenRepo interfaces.TokenRepository) interfaces.TokenService {
	return services.NewTokenService(tokenRepo)
}

// NewAccountService creates a new account service
func NewAccountService(
	accountRepo interfaces.AccountRepository,
	oauthService *oauth.Service,
	appLogger sctx.Logger,
) interfaces.AccountService {
	logger := appLogger.Withs(sctx.Fields{"component": "account-service"})

	// Create adapter for token refresh
	refresher := &oauthRefreshAdapter{oauthService: oauthService}

	return services.NewAccountService(accountRepo, refresher, logger)
}

// NewProxyService creates a new proxy service
func NewProxyService(
	accountRepo interfaces.AccountRepository,
	accountSvc interfaces.AccountService,
	claudeClient *clients.ClaudeAPIClient,
	appLogger sctx.Logger,
) interfaces.ProxyService {
	logger := appLogger.Withs(sctx.Fields{"component": "proxy-service"})
	return services.NewProxyService(accountRepo, accountSvc, claudeClient, logger)
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(tokenService interfaces.TokenService) *handlers.TokenHandler {
	return handlers.NewTokenHandler(tokenService)
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(proxyService interfaces.ProxyService) *handlers.ProxyHandler {
	return handlers.NewProxyHandler(proxyService)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config) *handlers.AuthHandler {
	return handlers.NewAuthHandler(cfg)
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	accountService interfaces.AccountService,
	accountRepo interfaces.AccountRepository,
) *handlers.AccountHandler {
	return handlers.NewAccountHandler(accountService, accountRepo)
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthService *oauth.Service,
	accountRepo interfaces.AccountRepository,
	accountSvc interfaces.AccountService,
	cfg *config.Config,
) *handlers.OAuthHandler {
	return handlers.NewOAuthHandler(oauthService, accountRepo, accountSvc, cfg.Claude.BaseURL)
}
