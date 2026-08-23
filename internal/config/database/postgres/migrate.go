package postgres

import (
	"context"

	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/schema"
	"github.com/uptrace/bun"
)

func Migrate(ctx context.Context, db *bun.DB) error {
	_, err := db.NewCreateTable().
		Model((*schema.URL)(nil)).
		IfNotExists().
		Exec(ctx)
	return err
}
