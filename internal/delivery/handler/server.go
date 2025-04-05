package handler

import (
	"github.com/Rafiur/go-url-shortener/internal/config"
	"github.com/Rafiur/go-url-shortener/internal/usecase"
)

type Handler struct {
	Config *config.Config
	//Logger logger.Logger
	URLPostgresService *usecase.URLPostgresService
	URLRedisService    *usecase.URLRedisService
}

func NewHandler(cfg *config.Config, urlPostgresService *usecase.URLPostgresService, urlRedisService *usecase.URLRedisService) *Handler {
	return &Handler{
		Config:             cfg,
		URLPostgresService: urlPostgresService,
		URLRedisService:    urlRedisService,
	}
}
