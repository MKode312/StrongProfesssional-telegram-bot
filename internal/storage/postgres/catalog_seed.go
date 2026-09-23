package postgres

import (
	"context"
	"fmt"
)

func (s *Storage) EnsureProductCatalog(ctx context.Context) error {
	var ready bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns c
			WHERE c.table_schema = 'public'
			  AND c.table_name = 'products'
			  AND c.column_name = 'category'
		)
		AND EXISTS (
			SELECT 1 FROM products
			WHERE active = true AND category IN ('degreaser', 'solvent')
		)`).Scan(&ready)
	if err != nil {
		return fmt.Errorf("check product catalog: %w", err)
	}
	if !ready {
		return fmt.Errorf("apply migrations/003_product_categories.sql using the database owner account")
	}
	return nil
}
