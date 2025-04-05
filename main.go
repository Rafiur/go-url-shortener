package main

import (
	"github.com/Rafiur/go-url-shortener/internal/config"
	"github.com/Rafiur/go-url-shortener/internal/router"

	"github.com/labstack/echo/v4"
)

func main() {
	config.NewConfig()

	e := echo.New()

	router.InitPostgresRoutes(e.Group("/pg"))
	router.InitRedisRoutes(e.Group("/redis"))

	e.Logger.Fatal(e.Start(":8080"))
}
