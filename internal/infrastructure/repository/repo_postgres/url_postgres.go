package repo_postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/schema"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

// pgUniqueViolation is the SQLSTATE code Postgres returns for a unique
// constraint breach.
const pgUniqueViolation = "23505"

type URLPostgresRepo struct {
	db *bun.DB
}

func NewURLPostgresRepo(db *bun.DB) *URLPostgresRepo {
	return &URLPostgresRepo{db: db}
}

// isDuplicate reports whether err is a unique-constraint violation, so the
// driver-specific error never leaks past this package.
func isDuplicate(err error) bool {
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) && pgErr.Field('C') == pgUniqueViolation {
		return true
	}
	return strings.Contains(err.Error(), pgUniqueViolation)
}

func (repo *URLPostgresRepo) Create(ctx context.Context, req *entity.URL) error {
	data := schema.URL{
		ShortCode:   req.ShortCode,
		OriginalURL: req.OriginalURL,
		CreatedAt:   req.CreatedAt,
	}
	_, err := repo.db.NewInsert().
		Model(&data).
		Returning("*").
		Exec(ctx)
	if err != nil {
		if isDuplicate(err) {
			return repository.ErrDuplicateShortCode
		}
		return err
	}

	req.ID = data.ID
	req.Clicks = data.Clicks
	req.CreatedAt = data.CreatedAt
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

// IncrementClicks bumps the hit counter for a short code. The update happens in
// the database so concurrent redirects cannot lose a count.
func (repo *URLPostgresRepo) IncrementClicks(ctx context.Context, shortCode string) error {
	_, err := repo.db.NewUpdate().
		Model((*schema.URL)(nil)).
		Set("clicks = clicks + 1").
		Where("short_code = ?", shortCode).
		Exec(ctx)
	return err
}
