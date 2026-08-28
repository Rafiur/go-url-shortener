package router

import (
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/labstack/echo/v4"
)

// InitRedisRoutes wires the Redis-backed endpoints. createLimit guards only
// link creation; reads and redirects are left unthrottled.
func InitRedisRoutes(g *echo.Group, h *handler.Handler, createLimit echo.MiddlewareFunc) {
	g.POST("", h.CreateRedisURL, createLimit)
	g.GET("", h.GetRedisURL)
	g.GET("/:shortcode", h.RedirectRedisURL)
}
