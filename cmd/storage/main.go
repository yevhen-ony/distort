package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"dos/internal/common/config"
	"dos/internal/common/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app, err := NewApp(&cfg)
	if err != nil {
		return err
	}
	defer app.Close()

	if err := app.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}
