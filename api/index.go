package handler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"go-api/internal/app"
	"go-api/internal/config"
)

var (
	once        sync.Once
	cached      http.Handler
	cachedErr   error
	cachedClose app.Cleanup
)

func initApp() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	h, _, cleanup, err := app.New(ctx, cfg)
	if err != nil {
		cachedErr = err
		return
	}

	cached = h
	cachedClose = cleanup
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initApp)

	if cachedErr != nil {
		http.Error(w, cachedErr.Error(), http.StatusInternalServerError)
		return
	}

	cached.ServeHTTP(w, r)
}

