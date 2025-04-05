package usecase

import (
	"context"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
)

type URLRedisService struct {
	URLRedisRepo repository.URLRedisRepo
}

func NewURLRedisService(repo repository.URLRedisRepo) *URLRedisService {
	return &URLRedisService{URLRedisRepo: repo}
}

func (s *URLRedisService) Create(ctx context.Context, req *entity.URL) error {
	err := s.URLRedisRepo.Create(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *URLRedisService) Get(ctx context.Context, shortcode string) (*entity.URL, error) {
	return s.URLRedisRepo.Get(ctx, shortcode)
}
