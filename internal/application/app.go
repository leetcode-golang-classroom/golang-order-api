package application

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8001",
		Handler: a.router,
	}
	err := a.rdb.Ping().Err()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}
	log.Println("Starting server")
	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
