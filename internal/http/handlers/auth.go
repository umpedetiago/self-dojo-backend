package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-api/internal/auth"
	"go-api/internal/email"
	"go-api/internal/repository"
)

type AuthHandler struct {
	users          *repository.UserRepository
	passwordResets *repository.PasswordResetRepository
	emailSender    email.Sender // nil = sem Resend; forgot-password retorna o token na resposta
	jwtSecret      string
	tokenTTL       time.Duration
	resetTokenTTL  time.Duration
}

func NewAuthHandler(users *repository.UserRepository, passwordResets *repository.PasswordResetRepository, emailSender email.Sender, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		users:          users,
		passwordResets: passwordResets,
		emailSender:    emailSender,
		jwtSecret:      jwtSecret,
		tokenTTL:       24 * time.Hour,
		resetTokenTTL:  1 * time.Hour,
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	username := strings.TrimSpace(strings.ToLower(req.Username))
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required (3–50 caracteres)"})
		return
	}
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	u, err := h.users.Create(c.Request.Context(), username, req.Email, hash)
	if err != nil {
		if err == repository.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		if err == repository.ErrUsernameAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
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
			"username":  u.Username,
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
			"username":  u.Username,
			"email":     u.Email,
			"createdAt": u.CreatedAt,
		},
		"token": token,
	})
}

// Logout responde 200. Com JWT stateless, o cliente deve descartar o token localmente.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	u, err := h.users.GetByEmail(c.Request.Context(), email)
	if err != nil {
		// Por segurança, sempre retornamos sucesso mesmo se o email não existir
		c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link will be sent"})
		return
	}

	token, err := auth.GenerateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate reset token"})
		return
	}

	expiresAt := time.Now().Add(h.resetTokenTTL)
	if err := h.passwordResets.Create(c.Request.Context(), token, u.ID, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reset token"})
		return
	}

	if h.emailSender != nil {
		if err := h.emailSender.SendPasswordReset(c.Request.Context(), u.Email, token, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send reset email"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link will be sent"})
		return
	}

	// Sem RESEND_API_KEY: retorna o token na resposta para testar o fluxo
	c.JSON(http.StatusOK, gin.H{
		"message":    "if the email exists, a reset link will be sent",
		"token":      token,
		"expires_at": expiresAt,
	})
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	tokenData, err := h.passwordResets.Get(c.Request.Context(), req.Token)
	if err != nil {
		if err == repository.ErrTokenNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
			return
		}
		if err == repository.ErrTokenExpired {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
			return
		}
		if err == repository.ErrTokenAlreadyUsed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token already used"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate token"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.users.UpdatePassword(c.Request.Context(), tokenData.UserID, hash); err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	if err := h.passwordResets.MarkAsUsed(c.Request.Context(), req.Token); err != nil {
		// Log error but don't fail the request - password was updated
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}
