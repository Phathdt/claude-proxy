package api

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"

	"claude-proxy/cmd/api/handlers"
	"claude-proxy/config"
	"claude-proxy/modules/auth/domain/interfaces"
	"claude-proxy/pkg/middleware"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
	"go.uber.org/fx"
)

// FrontendFS is set from main package with the embedded frontend files
var FrontendFS embed.FS

// StartAPIServer starts the API server component
func StartAPIServer(
	lc fx.Lifecycle,
	engine *gin.Engine,
	cfg *config.Config,
	appLogger sctx.Logger,
	tokenHandler *handlers.TokenHandler,
	proxyHandler *handlers.ProxyHandler,
	authHandler *handlers.AuthHandler,
	accountHandler *handlers.AccountHandler,
	oauthHandler *handlers.OAuthHandler,
	statisticsHandler *handlers.StatisticsHandler,
	sessionHandler *handlers.SessionHandler,
	tokenService interfaces.TokenService,
) {
	// Health check (public)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": fmt.Sprint(engine),
		})
	})

	// Protected Claude API proxy routes (user token authentication via Bearer)
	v1 := engine.Group("/v1")
	// v1.Use(middleware.OpenAICompatibility())
	v1.Use(middleware.BearerTokenAuth(tokenService, appLogger))
	{
		v1.Any("/*path", proxyHandler.ProxyRequest)
	}

	// OAuth routes (public - for account creation)
	oauth := engine.Group("/oauth")
	{
		oauth.GET("/authorize", oauthHandler.GetAuthorizeURL)
		oauth.POST("/exchange", oauthHandler.ExchangeCode)
	}

	// API routes for admin
	api := engine.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
			})
		})

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/validate", authHandler.Validate)
		}

		// Token routes (protected with API key)
		tokens := api.Group("/tokens")
		tokens.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
		{
			tokens.GET("", tokenHandler.ListTokens)
			tokens.POST("", tokenHandler.CreateToken)
			tokens.GET("/:id", tokenHandler.GetToken)
			tokens.PUT("/:id", tokenHandler.UpdateToken)
			tokens.DELETE("/:id", tokenHandler.DeleteToken)
		}

		// Account routes (protected with API key)
		accounts := api.Group("/accounts")
		accounts.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
		{
			accounts.GET("", accountHandler.ListAccounts)
			accounts.GET("/:id", accountHandler.GetAccount)
			accounts.PUT("/:id", accountHandler.UpdateAccount)
			accounts.DELETE("/:id", accountHandler.DeleteAccount)
		}

		// Admin routes (protected with API key)
		admin := api.Group("/admin")
		admin.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
		{
			admin.GET("/statistics", statisticsHandler.GetStatistics)
			admin.GET("/sessions", sessionHandler.ListAllSessions)
		}

		// Session routes (protected with API key)
		sessions := api.Group("/sessions")
		sessions.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
		{
			sessions.DELETE("/:id", sessionHandler.RevokeSession)
		}
	}

	// Serve static frontend files
	staticFS, err := fs.Sub(FrontendFS, "frontend/dist")
	if err == nil {
		engine.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if path == "/" {
				path = "/index.html"
			}

			// Try to serve the file
			filePath := path[1:] // Remove leading slash
			file, err := staticFS.Open(filePath)
			if err != nil {
				// If file not found, serve index.html for SPA routing
				indexFile, err := staticFS.Open("index.html")
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
					return
				}
				defer indexFile.Close()
				c.DataFromReader(http.StatusOK, -1, "text/html; charset=utf-8", indexFile, nil)
				return
			}
			defer file.Close()

			// Detect MIME type from file extension
			ext := filepath.Ext(filePath)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			// Serve the file with proper MIME type
			stat, _ := file.Stat()
			c.DataFromReader(http.StatusOK, stat.Size(), contentType, file, nil)
		})
	}

	port := cfg.Server.Port
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			appLogger.Withs(sctx.Fields{"port": port}).Info("Starting Claude Proxy Server")
			appLogger.Info("API Endpoints:")
			appLogger.Info("  Claude API Proxy (requires Bearer token):")
			appLogger.Info("    ANY  /v1/*path        - Proxy all Claude API requests")
			appLogger.Info("  Health:")
			appLogger.Info("    GET  /health          - Health check")
			appLogger.Info("    GET  /api/health      - Health check (legacy)")
			appLogger.Info("  OAuth (public):")
			appLogger.Info("    GET  /oauth/authorize - Get OAuth authorization URL")
			appLogger.Info("    POST /oauth/exchange  - Exchange OAuth code for account")
			appLogger.Info("    GET  /oauth/callback  - OAuth callback handler")
			appLogger.Info("  Auth (public):")
			appLogger.Info("    POST /api/auth/login    - Admin login")
			appLogger.Info("    POST /api/auth/validate - Validate API key")
			appLogger.Info("  Token Management (requires API key):")
			appLogger.Info("    GET    /api/tokens    - List all tokens")
			appLogger.Info("    POST   /api/tokens    - Create new token")
			appLogger.Info("    GET    /api/tokens/:id - Get token by ID")
			appLogger.Info("    PUT    /api/tokens/:id - Update token")
			appLogger.Info("    DELETE /api/tokens/:id - Delete token")
			appLogger.Info("  Account Management (requires API key):")
			appLogger.Info("    GET    /api/accounts         - List all accounts")
			appLogger.Info("    GET    /api/accounts/:id     - Get account by ID")
			appLogger.Info("    PUT    /api/accounts/:id     - Update account")
			appLogger.Info("    DELETE /api/accounts/:id     - Delete account")
			appLogger.Info("    GET    /api/accounts/:id/sessions - List account sessions")
			appLogger.Info("    DELETE /api/accounts/:id/sessions - Revoke all account sessions")
			appLogger.Info("  Session Management (requires API key):")
			appLogger.Info("    GET    /api/admin/sessions  - List all sessions")
			appLogger.Info("    DELETE /api/sessions/:id    - Revoke session by ID")

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					appLogger.Withs(sctx.Fields{"error": err}).Fatal("API server failed to start")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			appLogger.Info("Stopping API server...")
			return server.Shutdown(ctx)
		},
	})
}
