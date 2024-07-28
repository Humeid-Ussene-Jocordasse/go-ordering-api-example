package application

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
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
		Addr:    ":3000",
		Handler: a.router,
	}

	fmt.Println("Connecting to redis-db")
	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("could connect to redis instance %w", err)
	}

	fmt.Println("Starting application")

	ch := make(chan error, 1)
	//go func() {
	//	err = server.ListenAndServe()
	//	if err != nil {
	//		return fmt.Errorf("failed to start server: %w", err)
	//	}
	//}()
	err = <-ch
	go startServer(server.ListenAndServe(), ch)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return err
}

func startServer(err error, ch chan error) {
	ch <- fmt.Errorf("failed to start server: %w", err)
}
