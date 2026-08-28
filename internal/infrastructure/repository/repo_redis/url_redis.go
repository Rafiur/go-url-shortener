package repo_redis

import (
	"context"
	"errors"
	"time"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/redis/go-redis/v9"
)

type URLRedisRepo struct {
	client *redis.Client
}

// NewURLRedisRepo accepts a nil client, in which case the cache is treated as
// unavailable and every operation reports ErrRedisUnavailable.
func NewURLRedisRepo(client *redis.Client) *URLRedisRepo {
	return &URLRedisRepo{client: client}
}

func (repo *URLRedisRepo) Available() bool {
	return repo.client != nil
}

func (repo *URLRedisRepo) Create(ctx context.Context, url *entity.URL) error {
	if !repo.Available() {
		return repository.ErrRedisUnavailable
	}

	err := repo.client.Set(ctx, url.ShortCode, url.OriginalURL, 24*time.Hour).Err() //24 hours
	if err != nil {
		return err
	}
	return nil
}

func (repo *URLRedisRepo) Get(ctx context.Context, shortCode string) (*entity.URL, error) {
	if !repo.Available() {
		return nil, repository.ErrRedisUnavailable
	}

	val, err := repo.client.Get(ctx, shortCode).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("short code not found")
		}
		return nil, err
	}
	return &entity.URL{
		ShortCode:   shortCode,
		OriginalURL: val,
	}, nil
}
