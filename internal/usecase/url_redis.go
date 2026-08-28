package usecase

import (
	"context"
	"errors"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/Rafiur/go-url-shortener/utils"
	"github.com/asaskevich/govalidator"
)

type URLRedisService struct {
	URLRedisRepo repository.URLRedisRepo
}

func NewURLRedisService(repo repository.URLRedisRepo) *URLRedisService {
	return &URLRedisService{URLRedisRepo: repo}
}

// Available reports whether the cache is usable. Callers use it to decide
// between skipping the cache and refusing a cache-only request.
func (s *URLRedisService) Available() bool {
	return s.URLRedisRepo.Available()
}

func (s *URLRedisService) Create(ctx context.Context, req *entity.URL) error {
	if !govalidator.IsURL(req.OriginalURL) {
		return errors.New("Invalid URL")
	}

	if req.ShortCode == "" {
		req.ShortCode = utils.GenerateShortCode(shortCodeLength)
	}

	err := s.URLRedisRepo.Create(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *URLRedisService) Get(ctx context.Context, shortcode string) (*entity.URL, error) {
	return s.URLRedisRepo.Get(ctx, shortcode)
}
