package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/zapevnik/pr-review-service/internal/app"
	"github.com/zapevnik/pr-review-service/internal/app/config"
	"github.com/zapevnik/pr-review-service/internal/app/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configPath := config.Getenv("CONFIG_PATH", "config.yaml")

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var logCfg logger.Config

	if cfg.Env == "dev" {
		logCfg.Level = "debug"
		logCfg.Pretty = true
	} else {
		logCfg.Level = "info"
		logCfg.Pretty = false
	}

	logg := logger.New(logCfg)

	application := app.New(logg, &cfg)
	if err := application.Run(ctx); err != nil {
		logg.Error("application stopped", "error", err)
	}
}
