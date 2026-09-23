package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"str-prof-bot/internal/app"
	"str-prof-bot/internal/config"
	"str-prof-bot/internal/lib/logger/handlers/slogpretty"
	"str-prof-bot/internal/lib/logger/sl"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("application configuration loaded", "environment", cfg.Env)

	application, err := app.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize application", sl.Err(err))
		os.Exit(1)
	}

	go application.Run()

	stopContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopContext.Done()

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := application.GracefulShutdown(shutdownContext); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("graceful shutdown failed", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case "local":
		return slog.New(slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}.NewPrettyHandler(os.Stdout))
	case "prod":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
}
