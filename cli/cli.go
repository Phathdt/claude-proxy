package cli

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"claude-proxy/cmd/api"
)

// RunServer starts the API server (now the only service)
func RunServer(c *cli.Context) error {
	configPath := c.String("config")
	return RunServerWithConfig(configPath)
}

// RunServerWithConfig starts the API server with the specified configuration
func RunServerWithConfig(configPath string) error {
	app := fx.New(
		fx.Supply(configPath),
		api.APIProviders,
		fx.Invoke(api.StartAPIServer),
	)

	app.Run()
	return nil
}

// RunAPI starts the API service (same as RunServer now)
func RunAPI(c *cli.Context) error {
	return RunServer(c)
}

// RunAPIWithConfig starts the API service with the specified configuration
func RunAPIWithConfig(configPath string) error {
	return RunServerWithConfig(configPath)
}
