package usecase

import (
	"context"
	"errors"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/asaskevich/govalidator"
)

type URLPostgresService struct {
	URLPostgresRepo repository.URLPostgresRepo
}

func NewURLPostgresService(urlPostgresRepo repository.URLPostgresRepo) *URLPostgresService {
	return &URLPostgresService{URLPostgresRepo: urlPostgresRepo}
}

func (s *URLPostgresService) Create(ctx context.Context, req *entity.URL) error {

	if !govalidator.IsURL(req.OriginalURL) {
		return errors.New("Invalid URL")
	}

	err := s.URLPostgresRepo.Create(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *URLPostgresService) Get(ctx context.Context, shortcode string) (*entity.URL, error) {
	return s.URLPostgresRepo.Get(ctx, shortcode)
}
