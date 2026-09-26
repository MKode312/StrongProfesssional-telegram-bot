package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"str-prof-bot/internal/domain"
)

func (s *Storage) ProductsByCategory(ctx context.Context, category string) ([]domain.Product, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, description, price, category FROM products WHERE active = true AND category = $1 ORDER BY name`, category)
	if err != nil {
		return nil, fmt.Errorf("select products by category: %w", err)
	}
	defer rows.Close()

	products, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Product])
	if err != nil {
		return nil, fmt.Errorf("collect products by category: %w", err)
	}
	return products, nil
}

func (s *Storage) Products(ctx context.Context) ([]domain.Product, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, description, price, category FROM products WHERE active = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("select products: %w", err)
	}
	defer rows.Close()

	products, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Product])
	if err != nil {
		return nil, fmt.Errorf("collect products: %w", err)
	}
	return products, nil
}

func (s *Storage) SearchProducts(ctx context.Context, keyword string) ([]domain.Product, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, description, price, category
		FROM products
		WHERE active = true AND position(lower($1) in lower(name)) > 0
		ORDER BY name`, keyword)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	products, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Product])
	if err != nil {
		return nil, fmt.Errorf("collect searched products: %w", err)
	}
	return products, nil
}

func (s *Storage) AddToCart(ctx context.Context, telegramID, productID int64, capacity string, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("cart quantity must be positive")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO cart_items (telegram_id, product_id, capacity, quantity) VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id, product_id, capacity) DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity`, telegramID, productID, capacity, quantity)
	if err != nil {
		return fmt.Errorf("add cart item: %w", err)
	}
	return nil
}

func (s *Storage) ChangeCartItem(ctx context.Context, telegramID, productID int64, capacity string, delta int) error {
	if delta == 0 {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin cart update: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE cart_items SET quantity = quantity + $4 WHERE telegram_id = $1 AND product_id = $2 AND capacity = $3`, telegramID, productID, capacity, delta); err != nil {
		return fmt.Errorf("update cart item: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE telegram_id = $1 AND product_id = $2 AND capacity = $3 AND quantity <= 0`, telegramID, productID, capacity); err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Storage) Cart(ctx context.Context, telegramID int64) ([]domain.CartItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.name, p.description, p.price, p.category, ci.capacity, ci.quantity
		FROM cart_items ci JOIN products p ON p.id = ci.product_id
		WHERE ci.telegram_id = $1 ORDER BY p.name`, telegramID)
	if err != nil {
		return nil, fmt.Errorf("select cart: %w", err)
	}
	defer rows.Close()
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.CartItem])
	if err != nil {
		return nil, fmt.Errorf("collect cart: %w", err)
	}
	return items, nil
}
