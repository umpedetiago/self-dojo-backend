package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-api/internal/http/middleware"
)

type updateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

func (h *AuthHandler) Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	u, err := h.users.Update(c.Request.Context(), userID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       u.ID.String(),
			"username": u.Username,
			"email":    u.Email,
		},
	})
}
