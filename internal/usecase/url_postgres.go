package usecase

import (
	"context"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
)

type URLPostgresService struct {
	URLPostgresRepo repository.URLPostgresRepo
}

func NewURLPostgresService(urlPostgresRepo repository.URLPostgresRepo) *URLPostgresService {
	return &URLPostgresService{URLPostgresRepo: urlPostgresRepo}
}

func (s *URLPostgresService) Create(ctx context.Context, req *entity.URL) error {
	err := s.URLPostgresRepo.Create(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *URLPostgresService) Get(ctx context.Context, shortcode string) (*entity.URL, error) {
	return s.URLPostgresRepo.Get(ctx, shortcode)
}
