package repo_postgres

import (
	"context"
	"errors"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/redis/go-redis/v9"
	"time"
)

type URLRedisRepo struct {
	client *redis.Client
}

func NewURLRedisRepo(client *redis.Client) *URLRedisRepo {
	return &URLRedisRepo{client: client}
}

func (repo *URLRedisRepo) Create(ctx context.Context, url *entity.URL) error {
	err := repo.client.Set(ctx, url.ShortCode, url.OriginalURL, 24*time.Hour).Err() //24 hours
	if err != nil {
		return err
	}
	return nil
}

func (repo *URLRedisRepo) Get(ctx context.Context, shortCode string) (*entity.URL, error) {
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
