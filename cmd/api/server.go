package api

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"

	"claude-proxy/config"
	"claude-proxy/pkg/handlers"
	"claude-proxy/pkg/middleware"
	"claude-proxy/pkg/token"

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
	oauthHandler *handlers.OAuthHandler,
	messagesHandler *handlers.MessagesHandler,
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	appAccountHandler *handlers.AppAccountHandler,
	tokensHandler *handlers.TokensHandler,
	proxyHandler *handlers.ProxyHandler,
	tokenManager *token.Manager,
) {
	// OAuth routes (public, no auth required)
	oauth := engine.Group("/oauth")
	{
		oauth.GET("/authorize", oauthHandler.GetAuthorizeURL)
		oauth.POST("/exchange", oauthHandler.ExchangeCode)
		oauth.GET("/callback", oauthHandler.HandleCallback)
	}

	// Health check (public)
	engine.GET("/health", healthHandler.Check)

	// Protected Claude API proxy routes (user token authentication via Bearer)
	v1 := engine.Group("/v1")
	v1.Use(middleware.BearerTokenAuth(tokenManager, appLogger))
	{
		v1.GET("/models", proxyHandler.GetModels)
		v1.POST("/messages", proxyHandler.CreateMessage)
	}

	// Legacy /api routes for compatibility
	api := engine.Group("/api")
	{
		api.GET("/health", healthHandler.Check)

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/validate", authHandler.Validate)
		}

		// App account routes (protected with API key)
		appAccounts := api.Group("/app-accounts")
		appAccounts.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
		{
			appAccounts.POST("", appAccountHandler.CreateAppAccount)
			appAccounts.POST("/complete", appAccountHandler.CompleteAppAccount)
			appAccounts.GET("", appAccountHandler.ListAppAccounts)
			appAccounts.GET("/:id", appAccountHandler.GetAppAccount)
			appAccounts.PUT("/:id", appAccountHandler.UpdateAppAccount)
			appAccounts.DELETE("/:id", appAccountHandler.DeleteAppAccount)
		}

		// Token routes (protected with API key)
		tokens := api.Group("/tokens")
		tokens.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
		{
			tokens.GET("", tokensHandler.ListTokens)
			tokens.POST("", tokensHandler.CreateToken)
			tokens.POST("/generate-key", tokensHandler.GenerateKey)
			tokens.GET("/:id", tokensHandler.GetToken)
			tokens.PUT("/:id", tokensHandler.UpdateToken)
			tokens.DELETE("/:id", tokensHandler.DeleteToken)
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
			appLogger.Withs(sctx.Fields{"port": port}).Info("Starting Clove API Server")
			appLogger.Info("API Endpoints:")
			appLogger.Info("  OAuth:")
			appLogger.Info("    GET  /oauth/authorize - Generate OAuth URL (returns state + code_verifier)")
			appLogger.Info("    POST /oauth/exchange  - Exchange code for tokens (manual flow)")
			appLogger.Info("    GET  /oauth/callback  - OAuth callback page (shows code to copy)")
			appLogger.Info("  Claude API Proxy (requires Bearer token):")
			appLogger.Info("    GET  /v1/models       - Get available Claude models")
			appLogger.Info("    POST /v1/messages     - Send message to Claude")
			appLogger.Info("  Health:")
			appLogger.Info("    GET  /health          - Health check with account status")
			appLogger.Info("    GET  /api/health      - Health check (legacy)")
			appLogger.Info("  Auth:")
			appLogger.Info("    POST /api/auth/login     - Admin login with API key")
			appLogger.Info("    POST /api/auth/validate  - Validate API key")

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
