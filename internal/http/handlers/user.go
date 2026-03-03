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
	Username       *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	DisplayName    *string `json:"display_name,omitempty"`
	PhotoURL       *string `json:"photo_url,omitempty"`
	Role           *string `json:"role,omitempty"`
	MartialArtType *string `json:"martial_art_type,omitempty"`
}

func mapUserResponse(u *repository.User) gin.H {
	resp := gin.H{
		"id":               u.ID.String(),
		"username":         u.Username,
		"email":            u.Email,
		"createdAt":        u.CreatedAt,
		"display_name":     u.DisplayName,
		"photo_url":        u.PhotoURL,
		"role":             u.Role,
		"martial_art_type": u.MartialArt,
	}
	if u.HasAparador != nil {
		resp["has_aparadores"] = *u.HasAparador
	}
	return resp
}

func mapGraduationHistoryItemResponse(h repository.GraduationHistoryItem) gin.H {
	resp := gin.H{
		"id":                 h.ID.String(),
		"student_modality_id": h.StudentModalityID.String(),
		"belt_id":            h.BeltID,
		"degree":             h.Degree,
		"promoted_at":        h.PromotedAt,
		"notes":              h.Notes,
	}
	if h.PromotedBy != nil {
		resp["promoted_by"] = h.PromotedBy.String()
	}
	return resp
}

func mapStudentModalityResponse(m repository.StudentModality) gin.H {
	resp := gin.H{
		"id":                     m.ID.String(),
		"member_id":              m.MemberID.String(),
		"modality_id":            m.ModalityID.String(),
		"martial_art_type":       m.MartialArtType,
		"belt_id":                m.BeltID,
		"degree":                 m.Degree,
		"total_classes":          m.TotalClasses,
		"classes_at_current_belt": m.ClassesAtCurrentBelt,
		"enrolled_at":            m.EnrolledAt,
	}
	if m.AssignedTeacherID != nil {
		resp["assigned_teacher_id"] = m.AssignedTeacherID.String()
	}
	if m.PromotionDate != nil {
		resp["promotion_date"] = *m.PromotionDate
	}
	if len(m.GraduationHistory) > 0 {
		items := make([]gin.H, 0, len(m.GraduationHistory))
		for _, it := range m.GraduationHistory {
			items = append(items, mapGraduationHistoryItemResponse(it))
		}
		resp["graduation_history"] = items
	} else {
		resp["graduation_history"] = []any{}
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
	resp := mapUserResponse(u)

	if h.academies != nil {
		if modalities, err := h.academies.ListUserStudentModalities(c.Request.Context(), userID); err == nil {
			if len(modalities) > 0 {
				// Tenta priorizar a modalidade que bate com martial_art_type do usuário.
				var chosen *repository.StudentModality
				if u.MartialArt != "" {
					for i := range modalities {
						if modalities[i].MartialArtType == u.MartialArt {
							chosen = &modalities[i]
							break
						}
					}
				}
				// Se não encontrou por tipo, usa a primeira modalidade.
				if chosen == nil {
					chosen = &modalities[0]
				}

				resp["primary_student_modality_id"] = chosen.ID.String()

				// Retorna as modalidades agregadas (fonte principal de graduação/progresso).
				out := make([]gin.H, 0, len(modalities))
				for _, m := range modalities {
					out = append(out, mapStudentModalityResponse(m))
				}
				resp["student_modalities"] = out

				// Mantém compatibilidade: também preenche campos top-level com a modalidade escolhida.
				if chosen.BeltID != "" {
					resp["martial_art_type"] = chosen.MartialArtType
					resp["belt_id"] = chosen.BeltID
					resp["degree"] = chosen.Degree
					resp["total_classes"] = chosen.TotalClasses
				}

				// Conveniência: retorna dados mínimos da membership/academia do aluno.
				if member, err := h.academies.GetAcademyMemberByID(c.Request.Context(), chosen.MemberID); err == nil {
					resp["academy_id"] = member.AcademyID.String()
					resp["academy_status"] = member.Status
					if member.JoinedAt != nil {
						resp["joined_at"] = *member.JoinedAt
					}
				}
			} else {
				resp["student_modalities"] = []any{}
			}
		}
	}

	c.JSON(http.StatusOK, resp)
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
		req.MartialArtType == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
		return
	}

	u, err := h.users.UpdateProfile(c.Request.Context(), userID, repository.UpdateUserProfileInput{
		Username:       req.Username,
		DisplayName:    req.DisplayName,
		PhotoURL:       req.PhotoURL,
		Role:           req.Role,
		MartialArtType: req.MartialArtType,
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
