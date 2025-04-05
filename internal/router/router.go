package router

import (
	"github.com/Rafiur/go-url-shortener/internal/delivery/handler"
	"github.com/labstack/echo/v4"
)
func Route(e *echo.Echo,
	gqlHandler *handler.Handler,
	//handler *grpc_handler.Handler,
) {

	// GraphQL API
	e.POST("/graphql", func(c echo.Context) error {
		gqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})