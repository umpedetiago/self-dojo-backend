package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-api/internal/config"
	"go-api/internal/db"
	"go-api/internal/http/handlers"
	"go-api/internal/http/middleware"
	"go-api/internal/repository"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	pool, err := db.ConnectWithRetry(ctx, cfg.DatabaseURL, 10, 3*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.EnsureSchema(ctx, pool); err != nil {
		log.Fatal(err)
	}

	server := gin.Default()

	server.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	server.GET("/swagger", handlers.SwaggerUI)
	server.StaticFile("/openapi.yaml", "./docs/openapi.yaml")

	userRepo := repository.NewUserRepository(pool)
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)

	auth := server.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	server.GET("/me", middleware.AuthRequired(cfg.JWTSecret), func(c *gin.Context) {
		userID := middleware.MustGetUserID(c)
		c.JSON(http.StatusOK, gin.H{"userId": userID.String()})
	})

	if err := server.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
