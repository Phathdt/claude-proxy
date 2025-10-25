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

	// Protected API routes with API key authentication
	v1 := engine.Group("/v1")
	v1.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
	{
		v1.POST("/messages", messagesHandler.CreateMessage)
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
			appLogger.Info("  Claude API (requires X-API-Key header):")
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
