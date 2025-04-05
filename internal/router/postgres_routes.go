package router

import (
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/repo_postgres"
	"github.com/Rafiur/go-url-shortener/internal/postgres/handler"
	"github.com/labstack/echo/v4"
)

func InitPostgresRoutes(g *echo.Group) {
	r := repo_postgres.NewURLPostgresRepo()
	u := pgUsecase.NewURLUsecase(repo)
	h := handler.NewURLHandler(usecase)

	g.POST("/shorten", h.Shorten)
	g.GET("/:code", h.Resolve)
}
