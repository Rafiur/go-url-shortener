package router

import (
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/labstack/echo/v4"
)

func InitRedisRoutes(g *echo.Group, h *handler.Handler) {
	g.POST("", h.CreateRedisURL)
	g.GET("", h.GetRedisURL)
}
