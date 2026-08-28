package usecase

import (
	"context"
	"errors"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/Rafiur/go-url-shortener/utils"

	"github.com/asaskevich/govalidator"
)

const (
	shortCodeLength = 7
	// createAttempts bounds how many fresh codes we try before giving up. A
	// collision at 7 base64 characters is vanishingly rare, but it is a unique
	// column, so an unretried clash would surface to the user as an error.
	createAttempts = 4
)

// ErrShortCodeTaken means a caller-supplied alias is already in use.
var ErrShortCodeTaken = errors.New("that short code is already taken")

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

	// A custom alias is the caller's choice, so a clash is a real error rather
	// than something to retry behind their back.
	if req.ShortCode != "" {
		err := s.URLPostgresRepo.Create(ctx, req)
		if errors.Is(err, repository.ErrDuplicateShortCode) {
			return ErrShortCodeTaken
		}
		return err
	}

	var err error
	for attempt := 0; attempt < createAttempts; attempt++ {
		req.ShortCode = utils.GenerateShortCode(shortCodeLength)

		err = s.URLPostgresRepo.Create(ctx, req)
		if !errors.Is(err, repository.ErrDuplicateShortCode) {
			return err
		}
	}

	return err
}

func (s *URLPostgresService) Get(ctx context.Context, shortcode string) (*entity.URL, error) {
	return s.URLPostgresRepo.Get(ctx, shortcode)
}

// RecordClick increments a link's hit counter. Callers treat failure as
// non-fatal: a lost count must never cost the user their redirect.
func (s *URLPostgresService) RecordClick(ctx context.Context, shortcode string) error {
	return s.URLPostgresRepo.IncrementClicks(ctx, shortcode)
}
