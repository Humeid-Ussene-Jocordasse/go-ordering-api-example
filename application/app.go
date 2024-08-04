package application

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
	config Config
}

func New(config Config) *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
	}
	app.loadRoutes()

	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.ServerPort),
		Handler: app.router,
	}

	fmt.Println("Connecting to redis-db")
	// Ping a redis client to make sure the connection is working
	err := app.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("could connect to redis instance %w", err)
	}

	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Println("failed to close redis", err)
		}
	}()

	fmt.Println("Starting application")

	ch := make(chan error, 1)

	// Initiating a Go routine
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		// Closing the channel so that every listeners stop listing from it
		close(ch)

	}()
	// retuning the channel value, and checking if it is opened
	ctx.Done()

	// It's a Switch case functionality, but for channels
	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
