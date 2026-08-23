package handler

import (
	"fmt"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/utils"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateRedisURL(c echo.Context) error {
	var (
		req entity.Request
		ctx = c.Request().Context()
	)
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
