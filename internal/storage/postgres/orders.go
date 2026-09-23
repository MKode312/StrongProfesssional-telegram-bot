package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"str-prof-bot/internal/domain"
)

func (s *Storage) CreateOrder(ctx context.Context, user domain.User) (domain.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin order: %w", err)
	}
	defer tx.Rollback(ctx)

	items, err := s.cartTx(ctx, tx, user.TelegramID)
	if err != nil {
		return domain.Order{}, err
	}
	if len(items) == 0 {
		return domain.Order{}, fmt.Errorf("cart is empty")
	}
	total := int64(0)
	for _, item := range items {
		total += item.Price * int64(item.Quantity)
	}

	var order domain.Order
	err = tx.QueryRow(ctx, `INSERT INTO orders (telegram_id, total) VALUES ($1, $2) RETURNING id, created_at`, user.TelegramID, total).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}
	for _, item := range items {
		_, err = tx.Exec(ctx, `INSERT INTO order_items (order_id, product_id, product_name, capacity, unit_price, quantity) VALUES ($1, $2, $3, $4, $5, $6)`, order.ID, item.ID, item.Name, item.Capacity, item.Price, item.Quantity)
		if err != nil {
			return domain.Order{}, fmt.Errorf("insert order item: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE telegram_id = $1`, user.TelegramID); err != nil {
		return domain.Order{}, fmt.Errorf("clear cart: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit order: %w", err)
	}
	order.User, order.Items, order.Total = user, items, total
	return order, nil
}

func (s *Storage) cartTx(ctx context.Context, tx pgx.Tx, telegramID int64) ([]domain.CartItem, error) {
	rows, err := tx.Query(ctx, `
		SELECT p.id, p.name, p.description, p.price, p.category, ci.capacity, ci.quantity
		FROM cart_items ci JOIN products p ON p.id = ci.product_id
		WHERE ci.telegram_id = $1 ORDER BY p.name FOR UPDATE`, telegramID)
	if err != nil {
		return nil, fmt.Errorf("select order cart: %w", err)
	}
	defer rows.Close()
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.CartItem])
	if err != nil {
		return nil, fmt.Errorf("collect order cart: %w", err)
	}
	return items, nil
}
