package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-api/internal/app"
	"go-api/internal/config"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	handler, _, cleanup, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
