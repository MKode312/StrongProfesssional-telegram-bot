package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"str-prof-bot/internal/config"
	"str-prof-bot/internal/lib/logger/sl"
	"str-prof-bot/internal/services/mailer"
	"str-prof-bot/internal/services/orders"
	"str-prof-bot/internal/storage/postgres"
	"str-prof-bot/internal/telegram-bot/handlers"
)

type App struct {
	bot     *bot.Bot
	storage *postgres.Storage
	log     *slog.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	storage, err := postgres.New(context.Background(), cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("initialize postgres storage: %w", err)
	}
	if err := storage.EnsureProductCatalog(context.Background()); err != nil {
		storage.Close()
		return nil, fmt.Errorf("ensure product catalog: %w", err)
	}
	sender, err := mailer.New(log, cfg.SMTP)
	if err != nil {
		storage.Close()
		return nil, fmt.Errorf("initialize mailer: %w", err)
	}
	handler := handlers.New(log, orders.New(storage, sender))
	telegramBot, err := bot.New(cfg.Telegram.Token, bot.WithDefaultHandler(handler.Handle), bot.WithErrorsHandler(func(err error) {
		log.Error("telegram bot error", sl.Err(err))
	}))
	if err != nil {
		storage.Close()
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &App{bot: telegramBot, storage: storage, log: log, ctx: ctx, cancel: cancel, done: make(chan struct{})}, nil
}

func (a *App) Run() {
	a.log.Info("application started")
	a.bot.Start(a.ctx)
	close(a.done)
	a.log.Info("application stopped")
}

func (a *App) GracefulShutdown(ctx context.Context) error {
	a.cancel()
	a.log.Info("application shutdown started")
	select {
	case <-a.done:
		a.storage.Close()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
