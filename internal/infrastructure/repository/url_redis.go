package repository

import (
	"context"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
)

type URLRedisRepo interface {
	Create(ctx context.Context, url *entity.URL) error
	Get(ctx context.Context, shortCode string) (*entity.URL, error)
}
