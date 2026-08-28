package handler

import (
	"fmt"
	"net/http"

	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository"
	"github.com/Rafiur/go-url-shortener/utils"
	"github.com/labstack/echo/v4"
)

// cacheUnavailable short-circuits the Redis-only endpoints when the app started
// without a reachable cache. The Postgres endpoints and the root redirect keep
// working regardless.
func (h *Handler) cacheUnavailable(c echo.Context) error {
	return c.JSON(http.StatusServiceUnavailable, utils.APIResponse{
		Success: false,
		Error:   repository.ErrRedisUnavailable.Error(),
		Message: "Redis is not configured or was unreachable at startup",
	})
}

func (h *Handler) CreateRedisURL(c echo.Context) error {
	var (
		req entity.Request
		ctx = c.Request().Context()
	)
	if !h.URLRedisService.Available() {
		return h.cacheUnavailable(c)
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(400, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "Request Bind Error CreateRedisURL",
		})
	}

	url := entity.URL{
		OriginalURL: req.URL,
		ShortCode:   req.CustomShort,
	}

	err := h.URLRedisService.Create(ctx, &url)
	if err != nil {
		return c.JSON(400, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "Failed to Create Redis URL",
		})
	}
	return c.JSON(201, utils.APIResponse{
		Success: true,
		Data:    url,
		Message: "Redis URL Created Successfully",
	})
}

func (h *Handler) GetRedisURL(c echo.Context) error {
	var (
		urlID = c.QueryParam("shortcode")
		ctx   = c.Request().Context()
	)
	if !h.URLRedisService.Available() {
		return h.cacheUnavailable(c)
	}

	if urlID == "" {
		return c.JSON(400, utils.APIResponse{
			Success: false,
			Message: "shortcode is required",
		})
	}

	url, err := h.URLRedisService.Get(ctx, urlID)
	if err != nil {
		return c.JSON(404, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "URL not found",
		})
	}

	return c.JSON(200, utils.APIResponse{
		Success: true,
		Data:    url,
		Message: "URL retrieved successfully",
	})
}

func (h *Handler) RedirectRedisURL(c echo.Context) error {
	var (
		shortCode = c.Param("shortcode")
		ctx       = c.Request().Context()
	)
	if !h.URLRedisService.Available() {
		return h.cacheUnavailable(c)
	}

	url, err := h.URLRedisService.Get(ctx, shortCode)
	if err != nil {
		return c.JSON(404, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "URL not found",
		})
	}

	return c.Redirect(302, url.OriginalURL)
}
