package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Rafiur/go-url-shortener/internal/config"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/Rafiur/go-url-shortener/internal/usecase"

	"github.com/labstack/echo/v4"
)

// --- fakes ------------------------------------------------------------------

// fakePGRepo is guarded by a mutex because clicks are recorded from a detached
// goroutine, so the test and the handler touch it concurrently.
type fakePGRepo struct {
	mu        sync.Mutex
	url       *entity.URL
	getErr    error
	createErr error

	getCalls int
	clicks   chan string
}

func newFakePGRepo() *fakePGRepo {
	return &fakePGRepo{clicks: make(chan string, 8)}
}

func (f *fakePGRepo) Create(_ context.Context, req *entity.URL) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.createErr != nil {
		return f.createErr
	}
	f.url = req
	return nil
}

func (f *fakePGRepo) Get(_ context.Context, _ string) (*entity.URL, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.getCalls++
	return f.url, f.getErr
}

func (f *fakePGRepo) IncrementClicks(_ context.Context, shortCode string) error {
	select {
	case f.clicks <- shortCode:
	default:
	}
	return nil
}

func (f *fakePGRepo) postgresReads() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.getCalls
}

// awaitClick waits for the detached click-recording goroutine to land.
func (f *fakePGRepo) awaitClick(t *testing.T) string {
	t.Helper()

	select {
	case code := <-f.clicks:
		return code
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for a click to be recorded")
		return ""
	}
}

type fakeRedisRepo struct {
	mu        sync.Mutex
	available bool
	stored    map[string]string
}

func newFakeRedisRepo(available bool) *fakeRedisRepo {
	return &fakeRedisRepo{available: available, stored: map[string]string{}}
}

func (f *fakeRedisRepo) Available() bool { return f.available }

func (f *fakeRedisRepo) Create(_ context.Context, url *entity.URL) error {
	if !f.available {
		return repository.ErrRedisUnavailable
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.stored[url.ShortCode] = url.OriginalURL
	return nil
}

func (f *fakeRedisRepo) Get(_ context.Context, shortCode string) (*entity.URL, error) {
	if !f.available {
		return nil, repository.ErrRedisUnavailable
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	original, ok := f.stored[shortCode]
	if !ok {
		return nil, errors.New("short code not found")
	}
	return &entity.URL{ShortCode: shortCode, OriginalURL: original}, nil
}

func (f *fakeRedisRepo) cached(shortCode string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.stored[shortCode]
	return v, ok
}

// --- helpers ----------------------------------------------------------------

func newTestHandler(pg *fakePGRepo, rd *fakeRedisRepo) *Handler {
	return NewHandler(
		&config.Config{},
		usecase.NewURLPostgresService(pg),
		usecase.NewURLRedisService(rd),
	)
}

func newContext(e *echo.Echo, method, target, body string) (echo.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// withShortCode binds the :shortcode path parameter the handlers read.
func withShortCode(c echo.Context, path, code string) echo.Context {
	c.SetPath(path)
	c.SetParamNames("shortcode")
	c.SetParamValues(code)
	return c
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body was not JSON: %v (%s)", err, rec.Body.String())
	}
	return body
}

// --- tests ------------------------------------------------------------------

func TestIndexServesEmbeddedPage(t *testing.T) {
	h := newTestHandler(newFakePGRepo(), newFakeRedisRepo(true))
	c, rec := newContext(echo.New(), http.MethodGet, "/", "")

	if err := h.Index(c); err != nil {
		t.Fatalf("Index returned %v, want nil", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "<title>go-url-shortener</title>") {
		t.Error("Index did not serve the embedded frontend")
	}
}

func TestHealthReportsCacheState(t *testing.T) {
	tests := []struct {
		name      string
		available bool
		want      string
	}{
		{"cache up", true, "enabled"},
		{"cache down", false, "disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(newFakePGRepo(), newFakeRedisRepo(tt.available))
			c, rec := newContext(echo.New(), http.MethodGet, "/healthz", "")

			if err := h.Health(c); err != nil {
				t.Fatalf("Health returned %v, want nil", err)
			}

			// The service is healthy either way: the cache is an optimisation.
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			data, _ := decodeBody(t, rec)["data"].(map[string]any)
			if got := data["cache"]; got != tt.want {
				t.Errorf("cache = %v, want %q", got, tt.want)
			}
		})
	}
}

func TestRedirectServesFromCache(t *testing.T) {
	pg := newFakePGRepo()
	rd := newFakeRedisRepo(true)
	rd.stored["cached1"] = "https://example.com/from-cache"

	h := newTestHandler(pg, rd)
	c, rec := newContext(echo.New(), http.MethodGet, "/cached1", "")

	if err := h.Redirect(withShortCode(c, "/:shortcode", "cached1")); err != nil {
		t.Fatalf("Redirect returned %v, want nil", err)
	}

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "https://example.com/from-cache" {
		t.Errorf("Location = %q, want the cached target", got)
	}
	if reads := pg.postgresReads(); reads != 0 {
		t.Errorf("a cache hit made %d Postgres reads, want 0", reads)
	}
	if code := pg.awaitClick(t); code != "cached1" {
		t.Errorf("recorded click for %q, want %q", code, "cached1")
	}
}

func TestRedirectFallsBackToPostgresAndBackfillsCache(t *testing.T) {
	pg := newFakePGRepo()
	pg.url = &entity.URL{ShortCode: "miss123", OriginalURL: "https://example.com/from-db"}
	rd := newFakeRedisRepo(true)

	h := newTestHandler(pg, rd)
	c, rec := newContext(echo.New(), http.MethodGet, "/miss123", "")

	if err := h.Redirect(withShortCode(c, "/:shortcode", "miss123")); err != nil {
		t.Fatalf("Redirect returned %v, want nil", err)
	}

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if reads := pg.postgresReads(); reads != 1 {
		t.Errorf("a cache miss made %d Postgres reads, want 1", reads)
	}

	// The next request for this code must not reach Postgres again.
	if got, ok := rd.cached("miss123"); !ok || got != "https://example.com/from-db" {
		t.Errorf("cache holds (%q, %v) after a miss, want the resolved URL backfilled", got, ok)
	}
}

func TestRedirectWorksWithoutCache(t *testing.T) {
	pg := newFakePGRepo()
	pg.url = &entity.URL{ShortCode: "nocache", OriginalURL: "https://example.com/no-cache"}
	rd := newFakeRedisRepo(false)

	h := newTestHandler(pg, rd)
	c, rec := newContext(echo.New(), http.MethodGet, "/nocache", "")

	if err := h.Redirect(withShortCode(c, "/:shortcode", "nocache")); err != nil {
		t.Fatalf("Redirect returned %v, want nil", err)
	}

	if rec.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "https://example.com/no-cache" {
		t.Errorf("Location = %q, want the Postgres target", got)
	}
}

func TestRedirectUnknownCodeIs404(t *testing.T) {
	pg := newFakePGRepo()
	pg.getErr = errors.New("sql: no rows in result set")

	h := newTestHandler(pg, newFakeRedisRepo(true))
	c, rec := newContext(echo.New(), http.MethodGet, "/missing", "")

	if err := h.Redirect(withShortCode(c, "/:shortcode", "missing")); err != nil {
		t.Fatalf("Redirect returned %v, want nil", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if success := decodeBody(t, rec)["success"]; success != false {
		t.Errorf("success = %v, want false", success)
	}
}

func TestStatsReturnsClickCount(t *testing.T) {
	pg := newFakePGRepo()
	pg.url = &entity.URL{ShortCode: "abc1234", OriginalURL: "https://example.com", Clicks: 42}

	h := newTestHandler(pg, newFakeRedisRepo(true))
	c, rec := newContext(echo.New(), http.MethodGet, "/stats/abc1234", "")

	if err := h.Stats(withShortCode(c, "/stats/:shortcode", "abc1234")); err != nil {
		t.Fatalf("Stats returned %v, want nil", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	data, _ := decodeBody(t, rec)["data"].(map[string]any)
	if got := data["clicks"]; got != float64(42) {
		t.Errorf("clicks = %v, want 42", got)
	}
}

func TestCreatePostgresURLReturnsShortCode(t *testing.T) {
	h := newTestHandler(newFakePGRepo(), newFakeRedisRepo(true))
	c, rec := newContext(echo.New(), http.MethodPost, "/pg", `{"url":"https://example.com","short":""}`)

	if err := h.CreatePostgresURL(c); err != nil {
		t.Fatalf("CreatePostgresURL returned %v, want nil", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusCreated, rec.Body.String())
	}

	data, _ := decodeBody(t, rec)["data"].(map[string]any)
	if code, _ := data["short_code"].(string); len(code) == 0 {
		t.Error("response carried no short_code")
	}
}

// A taken alias is a conflict, not a malformed request.
func TestCreatePostgresURLTakenAliasIs409(t *testing.T) {
	pg := newFakePGRepo()
	pg.createErr = repository.ErrDuplicateShortCode

	h := newTestHandler(pg, newFakeRedisRepo(true))
	c, rec := newContext(echo.New(), http.MethodPost, "/pg", `{"url":"https://example.com","short":"taken"}`)

	if err := h.CreatePostgresURL(c); err != nil {
		t.Fatalf("CreatePostgresURL returned %v, want nil", err)
	}

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestCreatePostgresURLRejectsInvalidURL(t *testing.T) {
	h := newTestHandler(newFakePGRepo(), newFakeRedisRepo(true))
	c, rec := newContext(echo.New(), http.MethodPost, "/pg", `{"url":"not a url","short":""}`)

	if err := h.CreatePostgresURL(c); err != nil {
		t.Fatalf("CreatePostgresURL returned %v, want nil", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// The cache-only endpoints must say so rather than pretending to work.
func TestRedisEndpointsRefuseWhenCacheDown(t *testing.T) {
	h := newTestHandler(newFakePGRepo(), newFakeRedisRepo(false))
	c, rec := newContext(echo.New(), http.MethodPost, "/redis", `{"url":"https://example.com","short":""}`)

	if err := h.CreateRedisURL(c); err != nil {
		t.Fatalf("CreateRedisURL returned %v, want nil", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
