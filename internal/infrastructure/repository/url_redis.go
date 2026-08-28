package repository

import (
	"context"
	"errors"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
)

// ErrRedisUnavailable is returned when the cache was not reachable at startup
// and the application is running without it.
var ErrRedisUnavailable = errors.New("redis cache is unavailable")

type URLRedisRepo interface {
	Create(ctx context.Context, url *entity.URL) error
	Get(ctx context.Context, shortCode string) (*entity.URL, error)
	// Available reports whether the cache can be used at all. The redirect path
	// degrades to Postgres when it cannot.
	Available() bool
}
