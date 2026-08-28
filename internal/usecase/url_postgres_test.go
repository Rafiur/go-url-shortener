package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
)

// fakePostgresRepo stands in for the database so the usecase layer can be
// tested without one. This is the payoff of depending on the interface rather
// than on repo_postgres directly.
type fakePostgresRepo struct {
	// duplicatesBeforeSuccess makes the first N Create calls collide.
	duplicatesBeforeSuccess int
	alwaysDuplicate         bool
	createErr               error

	createdCodes []string
	getResult    *entity.URL
	getErr       error
	incremented  []string
}

func (f *fakePostgresRepo) Create(_ context.Context, req *entity.URL) error {
	f.createdCodes = append(f.createdCodes, req.ShortCode)

	if f.createErr != nil {
		return f.createErr
	}
	if f.alwaysDuplicate || len(f.createdCodes) <= f.duplicatesBeforeSuccess {
		return repository.ErrDuplicateShortCode
	}

	req.ID = int64(len(f.createdCodes))
	return nil
}

func (f *fakePostgresRepo) Get(_ context.Context, _ string) (*entity.URL, error) {
	return f.getResult, f.getErr
}

func (f *fakePostgresRepo) IncrementClicks(_ context.Context, shortCode string) error {
	f.incremented = append(f.incremented, shortCode)
	return nil
}

func TestCreateRejectsInvalidURL(t *testing.T) {
	repo := &fakePostgresRepo{}
	svc := NewURLPostgresService(repo)

	err := svc.Create(context.Background(), &entity.URL{OriginalURL: "not a url"})

	if err == nil {
		t.Fatal("Create accepted an invalid URL, want an error")
	}
	if len(repo.createdCodes) != 0 {
		t.Errorf("Create hit the repository %d times for an invalid URL, want 0", len(repo.createdCodes))
	}
}

func TestCreateGeneratesShortCode(t *testing.T) {
	repo := &fakePostgresRepo{}
	svc := NewURLPostgresService(repo)
	url := &entity.URL{OriginalURL: "https://example.com"}

	if err := svc.Create(context.Background(), url); err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}

	if len(url.ShortCode) != shortCodeLength {
		t.Errorf("generated code %q has length %d, want %d", url.ShortCode, len(url.ShortCode), shortCodeLength)
	}
	if len(repo.createdCodes) != 1 {
		t.Errorf("Create made %d repository calls, want 1", len(repo.createdCodes))
	}
}

func TestCreateKeepsCustomAlias(t *testing.T) {
	repo := &fakePostgresRepo{}
	svc := NewURLPostgresService(repo)
	url := &entity.URL{OriginalURL: "https://example.com", ShortCode: "my-link"}

	if err := svc.Create(context.Background(), url); err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}

	if url.ShortCode != "my-link" {
		t.Errorf("short code = %q, want the caller's alias %q", url.ShortCode, "my-link")
	}
}

// A caller-chosen alias must not be silently swapped for a random one.
func TestCreateReportsTakenAlias(t *testing.T) {
	repo := &fakePostgresRepo{alwaysDuplicate: true}
	svc := NewURLPostgresService(repo)
	url := &entity.URL{OriginalURL: "https://example.com", ShortCode: "taken"}

	err := svc.Create(context.Background(), url)

	if !errors.Is(err, ErrShortCodeTaken) {
		t.Fatalf("Create returned %v, want ErrShortCodeTaken", err)
	}
	if len(repo.createdCodes) != 1 {
		t.Errorf("Create retried a custom alias %d times, want exactly 1 attempt", len(repo.createdCodes))
	}
	if url.ShortCode != "taken" {
		t.Errorf("short code = %q, want the alias left untouched", url.ShortCode)
	}
}

func TestCreateRetriesOnCollision(t *testing.T) {
	repo := &fakePostgresRepo{duplicatesBeforeSuccess: 2}
	svc := NewURLPostgresService(repo)
	url := &entity.URL{OriginalURL: "https://example.com"}

	if err := svc.Create(context.Background(), url); err != nil {
		t.Fatalf("Create returned %v after retryable collisions, want nil", err)
	}

	if len(repo.createdCodes) != 3 {
		t.Fatalf("Create made %d attempts, want 3 (two collisions then success)", len(repo.createdCodes))
	}
	// Retrying with the same colliding code would loop forever.
	for i, a := range repo.createdCodes {
		for _, b := range repo.createdCodes[i+1:] {
			if a == b {
				t.Errorf("retry reused short code %q instead of generating a fresh one", a)
			}
		}
	}
}

func TestCreateGivesUpAfterMaxAttempts(t *testing.T) {
	repo := &fakePostgresRepo{alwaysDuplicate: true}
	svc := NewURLPostgresService(repo)

	err := svc.Create(context.Background(), &entity.URL{OriginalURL: "https://example.com"})

	if !errors.Is(err, repository.ErrDuplicateShortCode) {
		t.Fatalf("Create returned %v, want the duplicate error surfaced", err)
	}
	if len(repo.createdCodes) != createAttempts {
		t.Errorf("Create made %d attempts, want %d", len(repo.createdCodes), createAttempts)
	}
}

// A non-collision failure is not retryable and must surface immediately.
func TestCreateDoesNotRetryOtherErrors(t *testing.T) {
	boom := errors.New("connection refused")
	repo := &fakePostgresRepo{createErr: boom}
	svc := NewURLPostgresService(repo)

	err := svc.Create(context.Background(), &entity.URL{OriginalURL: "https://example.com"})

	if !errors.Is(err, boom) {
		t.Fatalf("Create returned %v, want the underlying error", err)
	}
	if len(repo.createdCodes) != 1 {
		t.Errorf("Create retried a non-collision error %d times, want 1 attempt", len(repo.createdCodes))
	}
}

func TestRecordClickDelegatesToRepo(t *testing.T) {
	repo := &fakePostgresRepo{}
	svc := NewURLPostgresService(repo)

	if err := svc.RecordClick(context.Background(), "abc1234"); err != nil {
		t.Fatalf("RecordClick returned %v, want nil", err)
	}

	if len(repo.incremented) != 1 || repo.incremented[0] != "abc1234" {
		t.Errorf("incremented = %v, want [abc1234]", repo.incremented)
	}
}
