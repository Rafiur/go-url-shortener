package usecase

import (
	"context"
	"testing"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
)

type fakeRedisRepo struct {
	available bool
	stored    map[string]string
	getResult *entity.URL
	getErr    error
}

func newFakeRedisRepo(available bool) *fakeRedisRepo {
	return &fakeRedisRepo{available: available, stored: map[string]string{}}
}

func (f *fakeRedisRepo) Available() bool { return f.available }

func (f *fakeRedisRepo) Create(_ context.Context, url *entity.URL) error {
	if !f.available {
		return repository.ErrRedisUnavailable
	}
	f.stored[url.ShortCode] = url.OriginalURL
	return nil
}

func (f *fakeRedisRepo) Get(_ context.Context, _ string) (*entity.URL, error) {
	if !f.available {
		return nil, repository.ErrRedisUnavailable
	}
	return f.getResult, f.getErr
}

func TestRedisAvailableReflectsRepo(t *testing.T) {
	for _, available := range []bool{true, false} {
		svc := NewURLRedisService(newFakeRedisRepo(available))

		if got := svc.Available(); got != available {
			t.Errorf("Available() = %v, want %v", got, available)
		}
	}
}

func TestRedisCreateGeneratesShortCode(t *testing.T) {
	repo := newFakeRedisRepo(true)
	svc := NewURLRedisService(repo)
	url := &entity.URL{OriginalURL: "https://example.com"}

	if err := svc.Create(context.Background(), url); err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}

	if len(url.ShortCode) != shortCodeLength {
		t.Errorf("generated code %q has length %d, want %d", url.ShortCode, len(url.ShortCode), shortCodeLength)
	}
	if repo.stored[url.ShortCode] != "https://example.com" {
		t.Errorf("stored[%q] = %q, want the original URL", url.ShortCode, repo.stored[url.ShortCode])
	}
}

func TestRedisCreateRejectsInvalidURL(t *testing.T) {
	repo := newFakeRedisRepo(true)
	svc := NewURLRedisService(repo)

	err := svc.Create(context.Background(), &entity.URL{OriginalURL: "nope"})

	if err == nil {
		t.Fatal("Create accepted an invalid URL, want an error")
	}
	if len(repo.stored) != 0 {
		t.Errorf("stored %d entries for an invalid URL, want 0", len(repo.stored))
	}
}

func TestRedisCreateFailsWhenUnavailable(t *testing.T) {
	svc := NewURLRedisService(newFakeRedisRepo(false))

	err := svc.Create(context.Background(), &entity.URL{OriginalURL: "https://example.com"})

	if err == nil {
		t.Fatal("Create succeeded with the cache unavailable, want an error")
	}
}
