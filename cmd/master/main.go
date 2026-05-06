package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"dos/internal/common/config"
	"dos/internal/common/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	cfg := Config{}
	err := config.LoadConfig(*configPath, &cfg)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger.Init(&cfg.Logger)
	
	app, err := NewApp(&cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("run app: %w", err)
	}
	return nil
}

