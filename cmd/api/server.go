package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"claude-proxy/config"

	"github.com/gin-gonic/gin"
	sctx "github.com/phathdt/service-context"
	"go.uber.org/fx"
)

// StartAPIServer starts the API server component
func StartAPIServer(
	lc fx.Lifecycle,
	engine *gin.Engine,
	cfg *config.Config,
	appLogger sctx.Logger,
) {
	// Global health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	port := cfg.Server.Port
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			appLogger.Withs(sctx.Fields{"port": port}).Info("Starting API Server")
			appLogger.Info("API Endpoints:")
			appLogger.Info("  GET  /health - General health check")

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
