package postgres

import (
	"context"
	"fmt"

	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/schema"
	"github.com/uptrace/bun"
)

func Migrate(ctx context.Context, db *bun.DB) error {
	if _, err := db.NewCreateTable().
		Model((*schema.URL)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		return err
	}

	// CREATE TABLE IF NOT EXISTS leaves an existing table untouched, so columns
	// added after the first deploy need their own idempotent statement.
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS clicks BIGINT NOT NULL DEFAULT 0`,
	); err != nil {
		return fmt.Errorf("add clicks column: %w", err)
	}

	// The redirect path looks links up by short_code on every request.
	if _, err := db.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls (short_code)`,
	); err != nil {
		return fmt.Errorf("create short_code index: %w", err)
	}

	return nil
}
