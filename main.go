package main

import (
	"fmt"
	"github.com/Rafiur/go-url-shortener/internal/config"
	"github.com/Rafiur/go-url-shortener/internal/config/database/postgres"
	"github.com/Rafiur/go-url-shortener/internal/config/database/redis"
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/repo_postgres"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/repo_redis"
	"github.com/Rafiur/go-url-shortener/internal/router"
	"github.com/Rafiur/go-url-shortener/internal/usecase"

	"github.com/labstack/echo/v4"
)

func main() {
	conf := config.NewConfig("config.env")
	dbPostgres := postgres.NewDB(conf)
	redisClient, err := redis.SetupRedis(conf)
	if err != nil {
		fmt.Println(err)
	}

	repoP := repo_postgres.NewURLPostgresRepo(dbPostgres)
	usecaseP := usecase.NewURLPostgresService(repoP)

	repoR := repo_redis.NewURLRedisRepo(redisClient)
	usecaseR := usecase.NewURLRedisService(repoR)

	container := handler.NewHandler(conf, usecaseP, usecaseR)

	e := echo.New()

	router.InitPostgresRoutes(e.Group("/pg"), container)
	router.InitRedisRoutes(e.Group("/redis"), container)

	e.Logger.Fatal(e.Start(":8080"))
}
