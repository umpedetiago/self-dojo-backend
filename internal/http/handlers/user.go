package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
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
		"id":                  h.ID.String(),
		"student_modality_id": h.StudentModalityID.String(),
		"belt_id":             h.BeltID,
		"degree":              h.Degree,
		"promoted_at":         h.PromotedAt,
		"notes":               h.Notes,
	}
	if h.PromotedBy != nil {
		resp["promoted_by"] = h.PromotedBy.String()
	}
	return resp
}

func mapStudentModalityResponse(m repository.StudentModality) gin.H {
	resp := gin.H{
		"id":                      m.ID.String(),
		"member_id":               m.MemberID.String(),
		"modality_id":             m.ModalityID.String(),
		"martial_art_type":        m.MartialArtType,
		"belt_id":                 m.BeltID,
		"degree":                  m.Degree,
		"total_classes":           m.TotalClasses,
		"classes_at_current_belt": m.ClassesAtCurrentBelt,
		"enrolled_at":             m.EnrolledAt,
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

// computeStudentModalityStats calcula valores agregados para uma student_modality,
// como aulas restantes até a próxima faixa e tempo de treino em dias.
// Usa configuração de graduação da academia quando disponível.
func (h *AuthHandler) computeStudentModalityStats(ctx context.Context, m repository.StudentModality, now time.Time) (classesUntilNextBelt *int, trainingTimeDays *int) {
	// Tempo de treino aproximado em dias (baseado em enrolled_at)
	days := int(now.Sub(m.EnrolledAt).Hours() / 24)
	if days < 0 {
		days = 0
	}
	trainingTimeDays = &days

	// Para calcular aulas até a próxima faixa, precisamos da configuração de graduação.
	if h.academies == nil {
		return nil, trainingTimeDays
	}

	modality, err := h.academies.GetModalityByID(ctx, m.ModalityID)
	if err != nil || modality == nil {
		return nil, trainingTimeDays
	}

	// Se a modalidade está configurada para usar a graduação padrão e não há
	// belt_configs no banco, deixamos o cálculo para o cliente (fallback).
	if modality.UseDefaultGraduaton && len(modality.BeltConfigs) == 0 {
		return nil, trainingTimeDays
	}

	// Ordena as faixas configuradas por min_classes (ascendente) para inferir a ordem.
	belts := make([]repository.BeltConfig, 0, len(modality.BeltConfigs))
	belts = append(belts, modality.BeltConfigs...)
	if len(belts) == 0 {
		return nil, trainingTimeDays
	}
	sort.Slice(belts, func(i, j int) bool {
		return belts[i].MinClasses < belts[j].MinClasses
	})

	// Encontra a posição da faixa atual e a próxima faixa configurada.
	currentIdx := -1
	for i, b := range belts {
		if b.BeltID == m.BeltID {
			currentIdx = i
			break
		}
	}
	if currentIdx == -1 || currentIdx >= len(belts)-1 {
		// Não há próxima faixa configurada ou não encontramos a atual.
		return nil, trainingTimeDays
	}

	nextBeltCfg := belts[currentIdx+1]
	currentBeltCfg := belts[currentIdx]
	if nextBeltCfg.MinClasses <= 0 {
		return nil, trainingTimeDays
	}

	remainingInCurrentBelt := currentBeltCfg.MinClasses - (m.Degree * *currentBeltCfg.MinClassesPerDegree) - m.ClassesAtCurrentBelt
	if remainingInCurrentBelt < 0 {
		remainingInCurrentBelt = 0
	}

	return &remainingInCurrentBelt, trainingTimeDays
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

				// Retorna as modalidades agregadas (fonte principal de graduação/progresso),
				// já com estatísticas calculadas pelo backend quando possível.
				out := make([]gin.H, 0, len(modalities))
				var primaryClassesUntilNextBelt *int
				var primaryTrainingTimeDays *int
				totalClassesAll := 0
				now := time.Now().UTC()

				for _, m := range modalities {
					item := mapStudentModalityResponse(m)

					classesUntilNextBelt, trainingTimeDays := h.computeStudentModalityStats(c.Request.Context(), m, now)
					if classesUntilNextBelt != nil {
						item["classes_until_next_belt"] = *classesUntilNextBelt
					}
					if trainingTimeDays != nil {
						item["training_time_days"] = *trainingTimeDays
					}

					if chosen != nil && m.ID == chosen.ID {
						primaryClassesUntilNextBelt = classesUntilNextBelt
						primaryTrainingTimeDays = trainingTimeDays
					}

					totalClassesAll += m.TotalClasses
					out = append(out, item)
				}
				resp["student_modalities"] = out

				// Soma de aulas em todas as modalidades aprovadas.
				resp["total_classes_all"] = totalClassesAll

				// Mantém compatibilidade: também preenche campos top-level com a modalidade escolhida.
				if chosen.BeltID != "" {
					resp["martial_art_type"] = chosen.MartialArtType
					resp["belt_id"] = chosen.BeltID
					resp["degree"] = chosen.Degree
					resp["total_classes"] = chosen.TotalClasses
				}

				// Preenche campos agregados globais a partir da modalidade principal.
				if primaryClassesUntilNextBelt != nil {
					resp["classes_until_next_belt"] = *primaryClassesUntilNextBelt
				}
				if primaryTrainingTimeDays != nil {
					resp["training_time_days"] = *primaryTrainingTimeDays
				}

				// Conveniência: retorna dados mínimos da membership/academia do aluno.
				if member, err := h.academies.GetAcademyMemberByID(c.Request.Context(), chosen.MemberID); err == nil {
					resp["academy_id"] = member.AcademyID.String()
					resp["academy_status"] = member.Status
					if member.JoinedAt != nil {
						resp["joined_at"] = *member.JoinedAt
					}
					// Nome da academia principal para exibição.
					if academy, err := h.academies.GetByID(c.Request.Context(), member.AcademyID); err == nil && academy != nil {
						resp["academy_name"] = academy.Name
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
