package repository

import (
	"context"
	"errors"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
)

// ErrDuplicateShortCode reports that a short code is already taken. The
// Postgres implementation translates the driver's unique-violation into this so
// the usecase layer can retry without knowing which database it is talking to.
var ErrDuplicateShortCode = errors.New("short code already exists")

type URLPostgresRepo interface {
	Create(ctx context.Context, req *entity.URL) error
	Get(ctx context.Context, shortCode string) (*entity.URL, error)
	IncrementClicks(ctx context.Context, shortCode string) error
}
