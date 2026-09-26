package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"str-prof-bot/internal/config"
)

type Storage struct{ pool *pgxpool.Pool }

func New(ctx context.Context, cfg config.PostgresConfig) (*Storage, error) {
	pool, err := pgxpool.New(ctx, cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() { s.pool.Close() }
