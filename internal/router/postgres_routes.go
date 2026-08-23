package router

import (
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/labstack/echo/v4"
)

func InitPostgresRoutes(g *echo.Group, h *handler.Handler) {
	g.POST("", h.CreatePostgresURL)
	g.GET("", h.GetPostgresURL)
	g.GET("/:shortcode", h.RedirectPostgresURL)
}
