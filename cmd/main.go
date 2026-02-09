package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-api/internal/config"
	"go-api/internal/db"
	"go-api/internal/email"
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
	server.Use(middleware.CORS())

	server.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	server.GET("/swagger", handlers.SwaggerUI)
	server.StaticFile("/openapi.yaml", "./docs/openapi.yaml")

	userRepo := repository.NewUserRepository(pool)
	passwordResetRepo := repository.NewPasswordResetRepository(pool)

	var emailSender email.Sender
	if cfg.ResendAPIKey != "" {
		var err error
		emailSender, err = email.NewResendSender(email.ResendConfig{
			APIKey:           cfg.ResendAPIKey,
			From:             cfg.EmailFrom,
			FrontendResetURL: cfg.FrontendResetPasswordURL,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	authHandler := handlers.NewAuthHandler(userRepo, passwordResetRepo, emailSender, cfg.JWTSecret)

	auth := server.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", middleware.AuthRequired(cfg.JWTSecret), authHandler.Logout)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	server.GET("/me", middleware.AuthRequired(cfg.JWTSecret), func(c *gin.Context) {
		userID := middleware.MustGetUserID(c)
		u, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":       u.ID.String(),
			"username": u.Username,
			"email":    u.Email,
		})
	})

	server.PUT("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Update)

	if err := server.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}

}
