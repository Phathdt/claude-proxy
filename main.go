package main

import (
	"embed"
	"log"
	"os"

	"claude-proxy/cmd/api"
	mycli "claude-proxy/cli"

	"github.com/urfave/cli/v2"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

func main() {
	// Set the frontend FS for the API server
	api.FrontendFS = frontendFS
	app := &cli.App{
		Name:  "claude-proxy",
		Usage: "Wallet risk assessment API service",
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Start the wallet checker API server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yaml",
						Usage:   "Configuration file path",
					},
				},
				Action: mycli.RunServer,
			},
			{
				Name:    "api",
				Aliases: []string{"a"},
				Usage:   "Start the API server (same as server)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yaml",
						Usage:   "Configuration file path",
					},
				},
				Action: mycli.RunAPI,
			},
		},
		Action: func(c *cli.Context) error {
			// Default action - run server with default config
			return mycli.RunServerWithConfig("config.yaml")
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
