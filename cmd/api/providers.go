package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"claude-proxy/cmd/api/handlers"
	"claude-proxy/config"
	authservices "claude-proxy/modules/auth/application/services"
	authinterfaces "claude-proxy/modules/auth/domain/interfaces"
	authclients "claude-proxy/modules/auth/infrastructure/clients"
	authjobs "claude-proxy/modules/auth/infrastructure/jobs"
	authrepos "claude-proxy/modules/auth/infrastructure/repositories"
	proxyservices "claude-proxy/modules/proxy/application/services"
	proxyinterfaces "claude-proxy/modules/proxy/domain/interfaces"
	proxyclients "claude-proxy/modules/proxy/infrastructure/clients"
	proxyjobs "claude-proxy/modules/proxy/infrastructure/jobs"
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

// CloveProviders provides Clove-specific domain providers
var CloveProviders = fx.Options(
	fx.Provide(
		// OAuth client
		NewOAuthClient,
		// Infrastructure - Memory Repositories (cache layer)
		fx.Annotate(
			NewMemoryAccountRepository,
			fx.ResultTags(`name:"cacheAccountRepo"`),
		),
		fx.Annotate(
			NewMemoryTokenRepository,
			fx.ResultTags(`name:"cacheTokenRepo"`),
		),
		fx.Annotate(
			NewMemorySessionRepository,
			fx.ResultTags(`name:"cacheSessionRepo"`),
		),
		// Infrastructure - JSON Repositories (persistence layer)
		fx.Annotate(
			NewJSONAccountRepository,
			fx.ResultTags(`name:"persistenceAccountRepo"`),
		),
		fx.Annotate(
			NewJSONTokenRepository,
			fx.ResultTags(`name:"persistenceTokenRepo"`),
		),
		fx.Annotate(
			NewJSONSessionRepository,
			fx.ResultTags(`name:"persistenceSessionRepo"`),
		),
		// Infrastructure - Clients
		NewClaudeAPIClient,
		// Application - Services (hybrid storage)
		fx.Annotate(
			NewTokenService,
			fx.ParamTags(`name:"cacheTokenRepo"`, `name:"persistenceTokenRepo"`, ``),
		),
		fx.Annotate(
			NewAccountService,
			fx.ParamTags(`name:"cacheAccountRepo"`, `name:"persistenceAccountRepo"`, ``, ``),
		),
		fx.Annotate(
			NewSessionService,
			fx.ParamTags(`name:"cacheSessionRepo"`, `name:"persistenceSessionRepo"`, ``, ``),
		),
		NewProxyService,
		// Infrastructure - Jobs
		NewSyncScheduler,
		NewTokenRefreshScheduler,
		NewSessionCleanupScheduler,
		// Handlers
		NewTokenHandler,
		NewProxyHandler,
		NewAuthHandler,
		NewAccountHandler,
		NewOAuthHandler,
		NewStatisticsHandler,
		NewSessionHandler,
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
	fx.Invoke(
		StartSyncScheduler,
		StartTokenRefreshScheduler,
		StartSessionCleanupScheduler,
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
func NewGinEngine(cfg *config.Config) *gin.Engine {
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

	// Timeout middleware - use configurable timeout for LLM API requests
	engine.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Server.RequestTimeout)
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

// NewOAuthClient creates a new OAuth client for Claude authentication
func NewOAuthClient(cfg *config.Config, appLogger sctx.Logger) authinterfaces.OAuthClient {
	logger := appLogger.Withs(sctx.Fields{"component": "oauth-client"})
	return authclients.NewOAuthClient(
		cfg.OAuth.ClientID,
		cfg.OAuth.AuthorizeURL,
		cfg.OAuth.TokenURL,
		cfg.OAuth.RedirectURI,
		cfg.OAuth.Scope,
		logger,
	)
}

// ============================================================================
// Memory Repository Providers (Fast in-memory operations)
// ============================================================================

// NewMemoryAccountRepository creates a new in-memory account repository (cache)
func NewMemoryAccountRepository(appLogger sctx.Logger) authinterfaces.CacheRepository {
	return authrepos.NewMemoryAccountRepository(appLogger)
}

// NewMemoryTokenRepository creates a new in-memory token repository (cache)
func NewMemoryTokenRepository(appLogger sctx.Logger) authinterfaces.TokenCacheRepository {
	return authrepos.NewMemoryTokenRepository(appLogger)
}

// NewMemorySessionRepository creates a new in-memory session repository (cache)
func NewMemorySessionRepository(appLogger sctx.Logger) authinterfaces.SessionCacheRepository {
	return authrepos.NewMemorySessionRepository(appLogger)
}

// ============================================================================
// JSON Repository Providers (Persistent storage)
// ============================================================================

// NewJSONAccountRepository creates a new JSON account persistence repository
func NewJSONAccountRepository(cfg *config.Config, appLogger sctx.Logger) (authinterfaces.PersistenceRepository, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "json-account-persistence-repository"})

	repo, err := authrepos.NewJSONAccountPersistenceRepository(cfg.Storage.DataFolder)
	if err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to create JSON account persistence repository")
		return nil, fmt.Errorf("failed to create JSON account persistence repository: %w", err)
	}

	logger.Info("JSON account persistence repository initialized successfully")
	return repo, nil
}

// NewJSONTokenRepository creates a new JSON token repository
func NewJSONTokenRepository(
	cfg *config.Config,
	appLogger sctx.Logger,
) (authinterfaces.TokenPersistenceRepository, error) {
	logger := appLogger.Withs(sctx.Fields{"component": "json-token-repository"})

	repo, err := authrepos.NewJSONTokenRepository(cfg.Storage.DataFolder)
	if err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to create JSON token repository")
		return nil, fmt.Errorf("failed to create JSON token repository: %w", err)
	}

	logger.Info("JSON token repository initialized successfully")
	return repo, nil
}

// NewJSONSessionRepository creates a new JSON session repository
func NewJSONSessionRepository(
	cfg *config.Config,
	appLogger sctx.Logger,
) (authinterfaces.SessionPersistenceRepository, error) {
	if !cfg.Session.Enabled {
		appLogger.Info("Session limiting disabled, skipping JSON session repository")
		return nil, nil
	}

	logger := appLogger.Withs(sctx.Fields{"component": "json-session-repository"})

	repo, err := authrepos.NewJSONSessionRepository(cfg.Storage.DataFolder)
	if err != nil {
		logger.Withs(sctx.Fields{"error": err}).Error("Failed to create JSON session repository")
		return nil, fmt.Errorf("failed to create JSON session repository: %w", err)
	}

	logger.Info("JSON session repository initialized successfully")
	return repo, nil
}

// ============================================================================
// Service Providers (Hybrid storage - inject both memory and JSON repos)
// ============================================================================

// NewTokenService creates a new token service with cache and persistence layers
func NewTokenService(
	cacheRepo authinterfaces.TokenCacheRepository,
	persistenceRepo authinterfaces.TokenPersistenceRepository,
	appLogger sctx.Logger,
) authinterfaces.TokenService {
	return authservices.NewTokenService(cacheRepo, persistenceRepo, appLogger)
}

// NewAccountService creates a new account service with cache and persistence layers
func NewAccountService(
	cacheRepo authinterfaces.CacheRepository,
	persistenceRepo authinterfaces.PersistenceRepository,
	oauthClient authinterfaces.OAuthClient,
	appLogger sctx.Logger,
) authinterfaces.AccountService {
	return authservices.NewAccountService(cacheRepo, persistenceRepo, oauthClient, appLogger)
}

// NewSessionService creates a new session service with cache and persistence layers
func NewSessionService(
	cacheRepo authinterfaces.SessionCacheRepository,
	persistenceRepo authinterfaces.SessionPersistenceRepository,
	cfg *config.Config,
	appLogger sctx.Logger,
) authinterfaces.SessionService {
	return authservices.NewSessionService(cacheRepo, persistenceRepo, cfg, appLogger)
}

// NewProxyService creates a new proxy service (only injects auth services)
func NewProxyService(
	accountSvc authinterfaces.AccountService,
	claudeClient *proxyclients.ClaudeAPIClient,
	sessionSvc authinterfaces.SessionService,
	appLogger sctx.Logger,
) proxyinterfaces.ProxyService {
	logger := appLogger.Withs(sctx.Fields{"component": "proxy-service"})
	return proxyservices.NewProxyService(accountSvc, claudeClient, sessionSvc, logger)
}

// ============================================================================
// Infrastructure - Clients
// ============================================================================

// NewClaudeAPIClient creates a new Claude API client
func NewClaudeAPIClient(cfg *config.Config, appLogger sctx.Logger) *proxyclients.ClaudeAPIClient {
	logger := appLogger.Withs(sctx.Fields{"component": "claude-api-client"})
	return proxyclients.NewClaudeAPIClient(cfg.Claude.BaseURL, cfg.Server.RequestTimeout, logger)
}

// ============================================================================
// Job Providers
// ============================================================================

// NewSyncScheduler creates a new sync scheduler for hybrid storage
func NewSyncScheduler(
	accountService authinterfaces.AccountService,
	tokenService authinterfaces.TokenService,
	sessionService authinterfaces.SessionService,
	cfg *config.Config,
	appLogger sctx.Logger,
) *authjobs.SyncScheduler {
	// Default sync interval: 1 minute
	syncInterval := 1 * time.Minute
	if cfg.Storage.SyncInterval > 0 {
		syncInterval = cfg.Storage.SyncInterval
	}

	return authjobs.NewSyncScheduler(
		accountService,
		tokenService,
		sessionService,
		syncInterval,
		appLogger,
	)
}

// StartSyncScheduler starts the sync scheduler with lifecycle management
func StartSyncScheduler(
	lc fx.Lifecycle,
	scheduler *authjobs.SyncScheduler,
	logger sctx.Logger,
) error {
	if err := scheduler.Start(); err != nil {
		return err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Performing final sync before shutdown")
			scheduler.Stop()
			if err := scheduler.FinalSync(); err != nil {
				logger.Withs(sctx.Fields{"error": err}).Error("Final sync failed")
				return err
			}
			return nil
		},
	})

	return nil
}

// NewTokenRefreshScheduler creates a new token refresh scheduler
func NewTokenRefreshScheduler(
	accountSvc authinterfaces.AccountService,
	logger sctx.Logger,
) *proxyjobs.Scheduler {
	return proxyjobs.NewScheduler(accountSvc, logger)
}

// StartTokenRefreshScheduler starts the token refresh scheduler with lifecycle management
func StartTokenRefreshScheduler(
	lc fx.Lifecycle,
	scheduler *proxyjobs.Scheduler,
	logger sctx.Logger,
) error {
	if err := scheduler.Start(); err != nil {
		return err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping token refresh scheduler")
			scheduler.Stop()
			return nil
		},
	})

	return nil
}

// NewSessionCleanupScheduler creates a session cleanup scheduler
func NewSessionCleanupScheduler(
	sessionService authinterfaces.SessionService,
	cfg *config.Config,
	logger sctx.Logger,
) *authjobs.SessionCleanupScheduler {
	if !cfg.Session.Enabled || !cfg.Session.CleanupEnabled {
		logger.Info("Session cleanup disabled")
		return nil
	}

	return authjobs.NewSessionCleanupScheduler(sessionService, cfg, logger)
}

// StartSessionCleanupScheduler starts the session cleanup scheduler with lifecycle management
func StartSessionCleanupScheduler(
	lc fx.Lifecycle,
	scheduler *authjobs.SessionCleanupScheduler,
	cfg *config.Config,
	logger sctx.Logger,
) error {
	if !cfg.Session.Enabled || !cfg.Session.CleanupEnabled || scheduler == nil {
		logger.Info("Session cleanup scheduler not started (disabled or scheduler is nil)")
		return nil
	}

	if err := scheduler.Start(); err != nil {
		return err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping session cleanup scheduler")
			scheduler.Stop()
			return nil
		},
	})

	return nil
}

// ============================================================================
// Handler Providers
// ============================================================================

// NewTokenHandler creates a new token handler
func NewTokenHandler(tokenService authinterfaces.TokenService) *handlers.TokenHandler {
	return handlers.NewTokenHandler(tokenService)
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(proxyService proxyinterfaces.ProxyService) *handlers.ProxyHandler {
	return handlers.NewProxyHandler(proxyService)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(tokenService authinterfaces.TokenService, cfg *config.Config) *handlers.AuthHandler {
	return handlers.NewAuthHandler(tokenService, cfg)
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	accountService authinterfaces.AccountService,
) *handlers.AccountHandler {
	return handlers.NewAccountHandler(accountService)
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthClient authinterfaces.OAuthClient,
	accountSvc authinterfaces.AccountService,
	cfg *config.Config,
) *handlers.OAuthHandler {
	return handlers.NewOAuthHandler(oauthClient, accountSvc, cfg.Claude.BaseURL)
}

// NewStatisticsHandler creates a new statistics handler
func NewStatisticsHandler(
	accountService authinterfaces.AccountService,
	appLogger sctx.Logger,
) *handlers.StatisticsHandler {
	logger := appLogger.Withs(sctx.Fields{"component": "statistics-handler"})
	return handlers.NewStatisticsHandler(accountService, logger)
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	sessionService authinterfaces.SessionService,
	appLogger sctx.Logger,
) *handlers.SessionHandler {
	return handlers.NewSessionHandler(sessionService, appLogger)
}
