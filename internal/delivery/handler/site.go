package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/utils"
	"github.com/Rafiur/go-url-shortener/web"

	"github.com/labstack/echo/v4"
)

// Index serves the embedded single-page frontend.
func (h *Handler) Index(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, web.Index)
}

// Health is a liveness probe for the container host. It stays healthy without
// the cache, since Postgres alone can serve every request.
func (h *Handler) Health(c echo.Context) error {
	cache := "disabled"
	if h.URLRedisService.Available() {
		cache = "enabled"
	}

	return c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    map[string]string{"cache": cache},
		Message: "ok",
	})
}

// Stats reports how many times a link has been followed.
func (h *Handler) Stats(c echo.Context) error {
	var (
		shortCode = c.Param("shortcode")
		ctx       = c.Request().Context()
	)

	url, err := h.URLPostgresService.Get(ctx, shortCode)
	if err != nil {
		return c.JSON(http.StatusNotFound, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "URL not found",
		})
	}

	return c.JSON(http.StatusOK, utils.APIResponse{
		Success: true,
		Data:    url,
		Message: "Stats retrieved successfully",
	})
}

// recordClick bumps the hit counter without holding up the redirect. It runs
// detached from the request so the write survives the response being sent, and
// a failure only costs a count.
func (h *Handler) recordClick(ctx context.Context, shortCode string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := h.URLPostgresService.RecordClick(ctx, shortCode); err != nil {
			log.Printf("could not record click for %q: %v", shortCode, err)
		}
	}()
}

// Redirect resolves a short code at the root of the domain so links read as
// https://host/abc1234 rather than https://host/pg/abc1234.
//
// Reads are cache-aside: Redis first, falling back to Postgres on a miss and
// backfilling the cache so the next hit is served from memory.
func (h *Handler) Redirect(c echo.Context) error {
	var (
		shortCode = c.Param("shortcode")
		ctx       = c.Request().Context()
	)

	cached := h.URLRedisService.Available()

	if cached {
		if url, err := h.URLRedisService.Get(ctx, shortCode); err == nil {
			h.recordClick(ctx, shortCode)
			return c.Redirect(http.StatusFound, url.OriginalURL)
		}
	}

	url, err := h.URLPostgresService.Get(ctx, shortCode)
	if err != nil {
		return c.JSON(http.StatusNotFound, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "URL not found",
		})
	}

	// Best effort: a failed backfill costs a Postgres read next time, nothing more.
	if cached {
		_ = h.URLRedisService.Create(ctx, &entity.URL{
			ShortCode:   url.ShortCode,
			OriginalURL: url.OriginalURL,
		})
	}

	h.recordClick(ctx, shortCode)
	return c.Redirect(http.StatusFound, url.OriginalURL)
}
