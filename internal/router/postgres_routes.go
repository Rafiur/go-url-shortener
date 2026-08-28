package router

import (
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/labstack/echo/v4"
)

// InitPostgresRoutes wires the Postgres-backed endpoints. createLimit guards
// only link creation; reads and redirects are left unthrottled.
func InitPostgresRoutes(g *echo.Group, h *handler.Handler, createLimit echo.MiddlewareFunc) {
	g.POST("", h.CreatePostgresURL, createLimit)
	g.GET("", h.GetPostgresURL)
	g.GET("/:shortcode", h.RedirectPostgresURL)
}
