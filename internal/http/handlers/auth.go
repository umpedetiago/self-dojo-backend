package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-api/internal/auth"
	"go-api/internal/repository"
)

type AuthHandler struct {
	users     *repository.UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthHandler(users *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		users:     users,
		jwtSecret: jwtSecret,
		tokenTTL:  24 * time.Hour,
	}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	u, err := h.users.Create(c.Request.Context(), req.Email, hash)
	if err != nil {
		if err == repository.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	token, err := auth.GenerateToken(u.ID, h.jwtSecret, h.tokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":        u.ID.String(),
			"email":     u.Email,
			"createdAt": u.CreatedAt,
		},
		"token": token,
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	u, err := h.users.GetByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !auth.CheckPassword(req.Password, u.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(u.ID, h.jwtSecret, h.tokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        u.ID.String(),
			"email":     u.Email,
			"createdAt": u.CreatedAt,
		},
		"token": token,
	})
}
