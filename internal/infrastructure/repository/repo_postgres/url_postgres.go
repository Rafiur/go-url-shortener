package repo_postgres

import (
	"context"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/schema"
	"github.com/uptrace/bun"
)

type URLPostgresRepo struct {
	db *bun.DB
}

func NewURLPostgresRepo(db *bun.DB) *URLPostgresRepo {
	return &URLPostgresRepo{db: db}
}

func (repo *URLPostgresRepo) Create(ctx context.Context, req *entity.URL) error {
	data := schema.URL{
		ShortCode:   req.ShortCode,
		OriginalURL: req.OriginalURL,
		CreatedAt:   req.CreatedAt,
	}
	_, err := repo.db.NewInsert().
		Model(&data).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (repo *URLPostgresRepo) Get(ctx context.Context, shortCode string) (*entity.URL, error) {
	var data schema.URL
	err := repo.db.NewSelect().
		Model(&data).
		Where("short_code = ?", shortCode).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return data.SchemaToEntity(), nil
}
