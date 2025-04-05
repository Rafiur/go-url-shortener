package repository

import (
	"context"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
)

type URLPostgresRepo interface {
	Create(ctx context.Context, req *entity.URL) error
	Get(ctx context.Context, shortCode string) (*entity.URL, error)
}
