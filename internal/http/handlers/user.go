package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-api/internal/http/middleware"
	"go-api/internal/repository"
	"go-api/internal/storage"
)

type updateRequest struct {
	Username           *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	DisplayName        *string `json:"display_name,omitempty"`
	PhotoURL           *string `json:"photo_url,omitempty"`
	Role               *string `json:"role,omitempty"`
	MartialArtType     *string `json:"martial_art_type,omitempty"`
	LegacyBeltID       *string `json:"legacy_belt_id,omitempty"`
	LegacyDegree       *int    `json:"legacy_degree,omitempty"`
	LegacyTotalClasses *int    `json:"legacy_total_classes,omitempty"`
	LegacyHasAparador  *bool   `json:"legacy_has_aparadores,omitempty"`
}

func mapUserResponse(u *repository.User) gin.H {
	resp := gin.H{
		"id":                   u.ID.String(),
		"username":             u.Username,
		"email":                u.Email,
		"createdAt":            u.CreatedAt,
		"display_name":         u.DisplayName,
		"photo_url":            u.PhotoURL,
		"role":                 u.Role,
		"martial_art_type":     u.MartialArt,
		"legacy_belt_id":       u.LegacyBeltID,
		"legacy_degree":        u.LegacyDegree,
		"legacy_total_classes": u.LegacyTotal,
	}
	if u.HasAparador != nil {
		resp["legacy_has_aparadores"] = *u.HasAparador
	}
	return resp
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	u, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, mapUserResponse(u))
}

func (h *AuthHandler) Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	if req.Username == nil &&
		req.DisplayName == nil &&
		req.PhotoURL == nil &&
		req.Role == nil &&
		req.MartialArtType == nil &&
		req.LegacyBeltID == nil &&
		req.LegacyDegree == nil &&
		req.LegacyTotalClasses == nil &&
		req.LegacyHasAparador == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
		return
	}

	u, err := h.users.UpdateProfile(c.Request.Context(), userID, repository.UpdateUserProfileInput{
		Username:           req.Username,
		DisplayName:        req.DisplayName,
		PhotoURL:           req.PhotoURL,
		Role:               req.Role,
		MartialArtType:     req.MartialArtType,
		LegacyBeltID:       req.LegacyBeltID,
		LegacyDegree:       req.LegacyDegree,
		LegacyTotalClasses: req.LegacyTotalClasses,
		LegacyHasAparador:  req.LegacyHasAparador,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": mapUserResponse(u),
	})
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	if h.avatarStorage == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "avatar storage not configured"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if fileHeader.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}
	if fileHeader.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5MB)"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file must be an image"})
		return
	}

	userID := middleware.MustGetUserID(c)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	uploadPath := fmt.Sprintf("profiles/%s/avatar_%d%s", userID.String(), time.Now().UnixMilli(), ext)
	publicURL, err := h.avatarStorage.Upload(c.Request.Context(), storage.UploadInput{
		Path:        uploadPath,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload avatar", "details": err.Error()})
		return
	}

	urlCopy := publicURL
	u, err := h.users.UpdateProfile(c.Request.Context(), userID, repository.UpdateUserProfileInput{
		PhotoURL: &urlCopy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile photo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"photo_url": publicURL,
		"user":      mapUserResponse(u),
	})
}
