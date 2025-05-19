package main

import (
	"data_storage/client"
	"data_storage/client/cli"
	"data_storage/config"
	"log"
)

func main() {
	// 1) Load .env / environment config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// 2) Parse flags into CLIArgs
	args, err := cli.ParseArgs(cfg.DefaultTTL)
	if err != nil {
		log.Fatalf("argument error: %v", err)
	}

	// 3) Create the HTTP SDK client
	sdk, err := client.NewClient(cfg.StoreServerURL, cfg.APIToken)
	if err != nil {
		log.Fatalf("client init error: %v", err)
	}

	// 4) Compose and run
	app := cli.NewCLI(sdk, cfg.DefaultTTL)
	if err := app.Run(args); err != nil {
		log.Fatalf("execution error: %v", err)
	}
}
