package handler

import (
	"fmt"
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/Rafiur/go-url-shortener/utils"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreatePostgresURL(c echo.Context) error {
	var (
		url entity.URL
		ctx = c.Request().Context()
	)
	if err := c.Bind(&url); err != nil {
		return c.JSON(400, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "Request Bind Error CreatePostgresURL",
		})
	}

	err := h.URLPostgresService.Create(ctx, &url)
	if err != nil {
		return c.JSON(400, utils.APIResponse{
			Success: false,
			Error:   fmt.Sprint(err),
			Message: "Failed to Create PostgresURL",
		})
	}

	return c.JSON(201, utils.APIResponse{
		Success: true,
		Message: "Successfully Created PostgresURL",
	})
}

func (h *Handler) GetPostgresURL(c echo.Context) error {
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

	url, err := h.URLPostgresService.Get(ctx, urlID)
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
