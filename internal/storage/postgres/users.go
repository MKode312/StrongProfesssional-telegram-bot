package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"str-prof-bot/internal/domain"
)

func (s *Storage) User(ctx context.Context, telegramID int64) (domain.User, bool, error) {
	var user domain.User
	err := s.pool.QueryRow(ctx, `SELECT telegram_id, email, phone, last_name, first_name FROM users WHERE telegram_id = $1`, telegramID).Scan(&user.TelegramID, &user.Email, &user.Phone, &user.LastName, &user.FirstName)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, fmt.Errorf("select user: %w", err)
	}
	return user, true, nil
}

func (s *Storage) SaveUser(ctx context.Context, user domain.User) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (telegram_id, email, phone, last_name, first_name) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id) DO UPDATE SET email = EXCLUDED.email, phone = EXCLUDED.phone, last_name = EXCLUDED.last_name, first_name = EXCLUDED.first_name, updated_at = now()`,
		user.TelegramID, user.Email, user.Phone, user.LastName, user.FirstName)
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}
