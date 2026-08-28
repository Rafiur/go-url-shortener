package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/Rafiur/go-url-shortener/internal/config"
	"github.com/Rafiur/go-url-shortener/internal/config/database/postgres"
	"github.com/Rafiur/go-url-shortener/internal/config/database/redis"
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/repo_postgres"
	"github.com/Rafiur/go-url-shortener/internal/infrastructure/repository/repo_redis"
	"github.com/Rafiur/go-url-shortener/internal/router"
	"github.com/Rafiur/go-url-shortener/internal/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	conf := config.NewConfig("config.env")

	dbPostgres := postgres.NewDB(conf)
	if err := postgres.Migrate(context.Background(), dbPostgres); err != nil {
		log.Fatalf("failed to migrate postgres schema: %v", err)
	}

	// The cache is an optimisation, not a dependency: if Redis is unreachable
	// the app still serves every link from Postgres, just without cache hits.
	redisClient, err := redis.SetupRedis(conf)
	if err != nil {
		log.Printf("WARNING: running without cache, all reads will hit postgres: %v", err)
		redisClient = nil
	}

	repoP := repo_postgres.NewURLPostgresRepo(dbPostgres)
	usecaseP := usecase.NewURLPostgresService(repoP)

	repoR := repo_redis.NewURLRedisRepo(redisClient)
	usecaseR := usecase.NewURLRedisService(repoR)

	container := handler.NewHandler(conf, usecaseP, usecaseR)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// Only needed when the frontend is hosted apart from this binary (GitHub
	// Pages); same-origin requests from the embedded UI never hit this.
	allowedOrigins := conf.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
	}))

	// Creation is the only expensive, write-heavy path, so it is the only one
	// rate limited — redirects stay unthrottled.
	createLimit := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(conf.CreateRateLimit),
				Burst:     conf.CreateRateBurst,
				ExpiresIn: 3 * time.Minute,
			},
		),
	})

	e.GET("/", container.Index)
	e.GET("/healthz", container.Health)
	e.GET("/stats/:shortcode", container.Stats)

	router.InitPostgresRoutes(e.Group("/pg"), container, createLimit)
	router.InitRedisRoutes(e.Group("/redis"), container, createLimit)

	// Registered last: the catch-all short-code resolver at the root.
	e.GET("/:shortcode", container.Redirect)

	e.Logger.Fatal(e.Start(":" + conf.Port))
}
