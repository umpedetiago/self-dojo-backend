package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-api/internal/http/middleware"
	"go-api/internal/repository"
)

type AcademyHandler struct {
	repo *repository.AcademyRepository
}

func NewAcademyHandler(repo *repository.AcademyRepository) *AcademyHandler {
	return &AcademyHandler{repo: repo}
}

type createAcademyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description,omitempty"`
	Address     *string  `json:"address,omitempty"`
	City        *string  `json:"city,omitempty"`
	State       *string  `json:"state,omitempty"`
	Phone       *string  `json:"phone,omitempty"`
	Email       *string  `json:"email,omitempty"`
	Website     *string  `json:"website,omitempty"`
	Modalities  []string `json:"modalities" binding:"required,min=1"`
}

type updateAcademyRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	Address     *string `json:"address,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Website     *string `json:"website,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type createModalityRequest struct {
	MartialArtType       string     `json:"martial_art_type" binding:"required"`
	MasterID             *uuid.UUID `json:"master_id,omitempty"`
	UseDefaultGraduation *bool      `json:"use_default_graduation,omitempty"`
}

type updateModalityRequest struct {
	MasterID             *uuid.UUID `json:"master_id,omitempty"`
	UseDefaultGraduation *bool      `json:"use_default_graduation,omitempty"`
	IsActive             *bool      `json:"is_active,omitempty"`
}

type updateMasterRequest struct {
	MasterID *uuid.UUID `json:"master_id,omitempty"`
}

type setTeachersRequest struct {
	TeacherIDs []uuid.UUID `json:"teacher_ids"`
}

type updateGraduationConfigRequest struct {
	UseDefaultGraduation bool                `json:"use_default_graduation"`
	BeltConfigs          []beltConfigRequest `json:"belt_configs"`
}

type beltConfigRequest struct {
	BeltID              string   `json:"belt_id" binding:"required"`
	BeltName            string   `json:"belt_name"`
	MinClasses          int      `json:"min_classes"`
	MinMonths           *int     `json:"min_months,omitempty"`
	MinClassesPerDegree *int     `json:"min_classes_per_degree,omitempty"`
	RequiresExam        bool     `json:"requires_exam"`
	ExamFee             *float64 `json:"exam_fee,omitempty"`
	Notes               *string  `json:"notes,omitempty"`
}

type listMembershipRequestsQuery struct {
	Status string `form:"status"`
}

type enrollStudentRequest struct {
	AcademyModalityID uuid.UUID `json:"academy_modality_id" binding:"required"`
	InitialBeltID     string    `json:"initial_belt_id" binding:"required"`
	InitialDegree     int       `json:"initial_degree"`
}

type updateStudentModalityRequest struct {
	AssignedTeacherID    *uuid.UUID `json:"assigned_teacher_id,omitempty"`
	TotalClasses         *int       `json:"total_classes,omitempty"`
	ClassesAtCurrentBelt *int       `json:"classes_at_current_belt,omitempty"`
}

type promoteStudentRequest struct {
	NewBeltID string  `json:"new_belt_id" binding:"required"`
	Degree    int     `json:"degree"`
	Notes     *string `json:"notes,omitempty"`
}

type createCheckInRequest struct {
	StudentModalityID uuid.UUID  `json:"student_modality_id" binding:"required"`
	ClassScheduleID   *uuid.UUID `json:"class_schedule_id,omitempty"`
	ClassType         *string    `json:"class_type,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
}

type createMyCheckInRequest struct {
	ClassScheduleID string  `json:"class_schedule_id" binding:"required,uuid4"`
	ClassType       *string `json:"class_type,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

type listCheckInsQuery struct {
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
}

type createClassScheduleRequest struct {
	ModalityID   *uuid.UUID `json:"modality_id,omitempty"`
	InstructorID *uuid.UUID `json:"instructor_id,omitempty"`
	DayOfWeek    int        `json:"day_of_week" binding:"required"`
	StartTime    string     `json:"start_time" binding:"required"`
	EndTime      string     `json:"end_time" binding:"required"`
	ClassType    *string    `json:"class_type,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	MaxStudents  *int       `json:"max_students,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

type updateClassScheduleRequest struct {
	ModalityID   *uuid.UUID `json:"modality_id,omitempty"`
	InstructorID *uuid.UUID `json:"instructor_id,omitempty"`
	DayOfWeek    *int       `json:"day_of_week,omitempty"`
	StartTime    *string    `json:"start_time,omitempty"`
	EndTime      *string    `json:"end_time,omitempty"`
	ClassType    *string    `json:"class_type,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	MaxStudents  *int       `json:"max_students,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

type listClassSchedulesQuery struct {
	ModalityID *uuid.UUID `form:"modalityId"`
	DayOfWeek  *int       `form:"dayOfWeek"`
	IsActive   *bool      `form:"isActive"`
}

type createStudentGroupRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type updateStudentGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type setGroupMembersRequest struct {
	MemberIDs []uuid.UUID `json:"member_ids"`
}

type setClassScheduleGroupsRequest struct {
	GroupIDs []uuid.UUID `json:"group_ids"`
}

func (h *AcademyHandler) ListOwnerAcademies(c *gin.Context) {
	if c.Query("owner") != "me" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner=me is required"})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListByOwner(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list academies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": mapAcademies(items)})
}

func (h *AcademyHandler) CreateAcademy(c *gin.Context) {
	var req createAcademyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	userID := middleware.MustGetUserID(c)
	academy, err := h.repo.Create(c.Request.Context(), userID, repository.CreateAcademyInput{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		Modalities:  req.Modalities,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create academy"})
		return
	}
	c.JSON(http.StatusCreated, academyToMap(*academy))
}

func (h *AcademyHandler) SearchAcademies(c *gin.Context) {
	limit := 20
	if rawLimit := c.Query("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil {
			limit = parsed
		}
	}

	items, err := h.repo.Search(c.Request.Context(), c.Query("q"), c.Query("city"), c.Query("modality"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search academies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": mapAcademies(items)})
}

func (h *AcademyHandler) GetAcademy(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	academy, err := h.repo.GetByID(c.Request.Context(), academyID)
	if err != nil {
		if errors.Is(err, repository.ErrAcademyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get academy"})
		return
	}
	c.JSON(http.StatusOK, academyToMap(*academy))
}

func (h *AcademyHandler) UpdateAcademy(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	var req updateAcademyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)

	academy, err := h.repo.Update(c.Request.Context(), academyID, userID, repository.UpdateAcademyInput{
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		IsActive:    req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update academy"})
		}
		return
	}
	c.JSON(http.StatusOK, academyToMap(*academy))
}

func (h *AcademyHandler) GetAcademyStats(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	stats, err := h.repo.GetStats(c.Request.Context(), academyID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get academy stats"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"approved_members": stats.ApprovedMembers,
		"pending_requests": stats.PendingRequests,
	})
}

func (h *AcademyHandler) GetMyMembershipStatus(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	userID := middleware.MustGetUserID(c)

	member, err := h.repo.GetMembership(c.Request.Context(), academyID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMembershipNotFound) {
			c.JSON(http.StatusOK, gin.H{"status": "none"})
			return
		}
		if errors.Is(err, repository.ErrAcademyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get membership status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": member.Status})
}

func (h *AcademyHandler) CreateMembershipRequest(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	userID := middleware.MustGetUserID(c)

	member, err := h.repo.CreateMembershipRequest(c.Request.Context(), academyID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		case errors.Is(err, repository.ErrMembershipAlreadyPending):
			c.JSON(http.StatusConflict, gin.H{"error": "membership request already pending"})
		case errors.Is(err, repository.ErrMembershipAlreadyApproved):
			c.JSON(http.StatusConflict, gin.H{"error": "already a member of this academy"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create membership request"})
		}
		return
	}
	c.JSON(http.StatusCreated, academyMemberToMap(*member))
}

func (h *AcademyHandler) CancelMyMembershipRequest(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	if err := h.repo.CancelMembershipRequest(c.Request.Context(), academyID, userID); err != nil {
		if errors.Is(err, repository.ErrMembershipNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "pending membership request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel membership request"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AcademyHandler) ListMembershipRequests(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	var query listMembershipRequestsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListMembershipRequests(c.Request.Context(), academyID, userID, query.Status)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list membership requests"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, academyMemberToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) ApproveMembershipRequest(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	member, err := h.repo.ApproveMembershipRequest(c.Request.Context(), academyID, memberID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		case errors.Is(err, repository.ErrMembershipNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "membership request not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve membership request"})
		}
		return
	}
	c.JSON(http.StatusOK, academyMemberToMap(*member))
}

func (h *AcademyHandler) RejectMembershipRequest(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	if err := h.repo.RejectMembershipRequest(c.Request.Context(), academyID, memberID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		case errors.Is(err, repository.ErrMembershipNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "membership request not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject membership request"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AcademyHandler) ListStudents(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListStudents(c.Request.Context(), academyID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list students"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, academyStudentToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) GetStudent(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	student, err := h.repo.GetStudent(c.Request.Context(), academyID, memberID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		case errors.Is(err, repository.ErrStudentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get student"})
		}
		return
	}
	c.JSON(http.StatusOK, academyStudentToMap(*student))
}

func (h *AcademyHandler) EnrollStudentInModality(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	var req enrollStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.EnrollStudentInModality(c.Request.Context(), academyID, memberID, userID, repository.EnrollStudentInput{
		AcademyModalityID: req.AcademyModalityID,
		InitialBeltID:     req.InitialBeltID,
		InitialDegree:     req.InitialDegree,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		case errors.Is(err, repository.ErrStudentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		case errors.Is(err, repository.ErrModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy modality not found"})
		case errors.Is(err, repository.ErrStudentAlreadyEnrolled):
			c.JSON(http.StatusConflict, gin.H{"error": "student already enrolled in modality"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enroll student in modality"})
		}
		return
	}
	c.JSON(http.StatusCreated, studentModalityToMap(*item))
}

func (h *AcademyHandler) UpdateStudentModality(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	studentModalityID, err := uuid.Parse(c.Param("studentModalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student modality id"})
		return
	}
	var req updateStudentModalityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.UpdateStudentModality(c.Request.Context(), academyID, memberID, studentModalityID, userID, repository.UpdateStudentModalityInput{
		AssignedTeacherID:    req.AssignedTeacherID,
		TotalClasses:         req.TotalClasses,
		ClassesAtCurrentBelt: req.ClassesAtCurrentBelt,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update student modality"})
		}
		return
	}
	c.JSON(http.StatusOK, studentModalityToMap(*item))
}

func (h *AcademyHandler) DeleteStudentModality(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	studentModalityID, err := uuid.Parse(c.Param("studentModalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student modality id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	if err := h.repo.DeleteStudentModality(c.Request.Context(), academyID, memberID, studentModalityID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete student modality"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AcademyHandler) PromoteStudent(c *gin.Context) {
	studentModalityID, err := uuid.Parse(c.Param("studentModalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student modality id"})
		return
	}
	var req promoteStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.PromoteStudent(c.Request.Context(), studentModalityID, userID, repository.PromoteStudentInput{
		NewBeltID: req.NewBeltID,
		Degree:    req.Degree,
		Notes:     req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to promote student"})
		}
		return
	}
	c.JSON(http.StatusOK, studentModalityToMap(*item))
}

func (h *AcademyHandler) ListGraduationHistory(c *gin.Context) {
	studentModalityID, err := uuid.Parse(c.Param("studentModalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student modality id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListGraduationHistory(c.Request.Context(), studentModalityID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list graduation history"})
		}
		return
	}
	history := make([]gin.H, 0, len(items))
	for _, item := range items {
		history = append(history, graduationHistoryToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": history})
}

func (h *AcademyHandler) CreateCheckIn(c *gin.Context) {
	var req createCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.CreateCheckIn(c.Request.Context(), userID, repository.CreateCheckInInput{
		StudentModalityID: req.StudentModalityID,
		ClassScheduleID:   req.ClassScheduleID,
		ClassType:         req.ClassType,
		Notes:             req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create check-in"})
		}
		return
	}
	c.JSON(http.StatusCreated, checkInToMap(*item))
}

// CreateMyCheckIn permite que o aluno autenticado faça check-in em um horário,
// resolvendo automaticamente o student_modality_id apropriado.
func (h *AcademyHandler) CreateMyCheckIn(c *gin.Context) {
	var req createMyCheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	scheduleID, err := uuid.Parse(req.ClassScheduleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class_schedule_id"})
		return
	}

	userID := middleware.MustGetUserID(c)

	// Garante que o usuário pode acessar a academia do horário.
	schedule, err := h.repo.GetClassScheduleForActor(c.Request.Context(), scheduleID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load class schedule"})
		}
		return
	}

	// Descobre as modalidades do aluno (student_modalities) aprovadas.
	modalities, err := h.repo.ListUserStudentModalities(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load student modalities"})
		return
	}
	if len(modalities) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student not enrolled in any modality"})
		return
	}

	var studentModalityID uuid.UUID

	if schedule.ModalityID != nil {
		for _, m := range modalities {
			if m.ModalityID == *schedule.ModalityID {
				studentModalityID = m.ID
				break
			}
		}
		if studentModalityID == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "student not enrolled in this modality"})
			return
		}
	} else {
		// Horário geral (sem modality_id) - usa a primeira modalidade do aluno.
		studentModalityID = modalities[0].ID
	}

	input := repository.CreateCheckInInput{
		StudentModalityID: studentModalityID,
		ClassScheduleID:   &schedule.ID,
	}

	if req.ClassType != nil {
		classType := strings.TrimSpace(*req.ClassType)
		if classType != "" {
			input.ClassType = &classType
		}
	}
	if req.Notes != nil {
		notes := strings.TrimSpace(*req.Notes)
		if notes != "" {
			input.Notes = &notes
		}
	}

	item, err := h.repo.CreateCheckIn(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create check-in"})
		}
		return
	}

	c.JSON(http.StatusCreated, checkInToMap(*item))
}

func (h *AcademyHandler) ListCheckIns(c *gin.Context) {
	studentModalityID, err := uuid.Parse(c.Param("studentModalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student modality id"})
		return
	}
	var query listCheckInsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query", "details": err.Error()})
		return
	}
	var startDate *time.Time
	if strings.TrimSpace(query.StartDate) != "" {
		parsed, err := time.Parse(time.RFC3339, query.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid startDate (RFC3339 expected)"})
			return
		}
		startDate = &parsed
	}
	var endDate *time.Time
	if strings.TrimSpace(query.EndDate) != "" {
		parsed, err := time.Parse(time.RFC3339, query.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endDate (RFC3339 expected)"})
			return
		}
		endDate = &parsed
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListCheckIns(c.Request.Context(), studentModalityID, userID, startDate, endDate)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list check-ins"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, checkInToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) CreateClassSchedule(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	var req createClassScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.CreateClassSchedule(c.Request.Context(), academyID, userID, repository.CreateClassScheduleInput{
		ModalityID:   req.ModalityID,
		InstructorID: req.InstructorID,
		DayOfWeek:    req.DayOfWeek,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		ClassType:    req.ClassType,
		IsActive:     req.IsActive,
		MaxStudents:  req.MaxStudents,
		Notes:        req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create class schedule"})
		}
		return
	}
	c.JSON(http.StatusCreated, classScheduleToMap(*item))
}

func (h *AcademyHandler) ListClassSchedules(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	var query listClassSchedulesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListClassSchedules(c.Request.Context(), academyID, userID, query.ModalityID, query.DayOfWeek, query.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list class schedules"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, classScheduleToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) GetClassSchedule(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	scheduleID, err := uuid.Parse(c.Param("scheduleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.GetClassSchedule(c.Request.Context(), academyID, scheduleID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get class schedule"})
		}
		return
	}
	c.JSON(http.StatusOK, classScheduleToMap(*item))
}

func (h *AcademyHandler) UpdateClassSchedule(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	scheduleID, err := uuid.Parse(c.Param("scheduleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}
	var req updateClassScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.UpdateClassSchedule(c.Request.Context(), academyID, scheduleID, userID, repository.UpdateClassScheduleInput{
		ModalityID:   req.ModalityID,
		InstructorID: req.InstructorID,
		DayOfWeek:    req.DayOfWeek,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		ClassType:    req.ClassType,
		IsActive:     req.IsActive,
		MaxStudents:  req.MaxStudents,
		Notes:        req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update class schedule"})
		}
		return
	}
	c.JSON(http.StatusOK, classScheduleToMap(*item))
}

func (h *AcademyHandler) DeleteClassSchedule(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	scheduleID, err := uuid.Parse(c.Param("scheduleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	if err := h.repo.DeleteClassSchedule(c.Request.Context(), academyID, scheduleID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete class schedule"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AcademyHandler) ListStudentGroups(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListStudentGroups(c.Request.Context(), academyID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list student groups"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, studentGroupToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) CreateStudentGroup(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	var req createStudentGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.CreateStudentGroup(c.Request.Context(), academyID, userID, repository.CreateStudentGroupInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create student group"})
		}
		return
	}
	c.JSON(http.StatusCreated, studentGroupToMap(*item))
}

func (h *AcademyHandler) UpdateStudentGroup(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req updateStudentGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.UpdateStudentGroup(c.Request.Context(), academyID, groupID, userID, repository.UpdateStudentGroupInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student group not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update student group"})
		}
		return
	}
	c.JSON(http.StatusOK, studentGroupToMap(*item))
}

func (h *AcademyHandler) DeleteStudentGroup(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	if err := h.repo.DeleteStudentGroup(c.Request.Context(), academyID, groupID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student group not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete student group"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AcademyHandler) ListStudentGroupMembers(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListStudentGroupMembers(c.Request.Context(), academyID, groupID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student group not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list student group members"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, academyMemberToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) SetStudentGroupMembers(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req setGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.SetStudentGroupMembers(c.Request.Context(), academyID, groupID, userID, req.MemberIDs)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrStudentGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student group not found"})
		case errors.Is(err, repository.ErrStudentNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set student group members"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, academyMemberToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) ListClassScheduleGroups(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	scheduleID, err := uuid.Parse(c.Param("scheduleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListClassScheduleGroups(c.Request.Context(), academyID, scheduleID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list class schedule groups"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, studentGroupToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) SetClassScheduleGroups(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	scheduleID, err := uuid.Parse(c.Param("scheduleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}
	var req setClassScheduleGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.SetClassScheduleGroups(c.Request.Context(), academyID, scheduleID, userID, req.GroupIDs)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrClassScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "class schedule not found"})
		case errors.Is(err, repository.ErrStudentGroupNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "student group not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set class schedule groups"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, studentGroupToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) ListAvailableSchedulesForCheckIn(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	// 0=domingo, 6=sabado
	now := time.Now()
	dayOfWeek := int(now.Weekday())
	userID := middleware.MustGetUserID(c)
	items, err := h.repo.ListAvailableSchedulesForCheckIn(c.Request.Context(), academyID, userID, dayOfWeek)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list available schedules"})
		}
		return
	}
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, classScheduleToMap(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *AcademyHandler) ListModalities(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	items, err := h.repo.ListModalities(c.Request.Context(), academyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list modalities"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": mapModalities(items)})
}

func (h *AcademyHandler) CreateModality(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	var req createModalityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.CreateModality(c.Request.Context(), academyID, userID, repository.CreateModalityInput{
		MartialArtType:       req.MartialArtType,
		MasterID:             req.MasterID,
		UseDefaultGraduation: req.UseDefaultGraduation,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrAcademyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "academy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create modality"})
		}
		return
	}
	c.JSON(http.StatusCreated, modalityToMap(*item))
}

func (h *AcademyHandler) UpdateModality(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	modalityID, err := uuid.Parse(c.Param("modalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid modality id"})
		return
	}
	var req updateModalityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.UpdateModality(c.Request.Context(), academyID, modalityID, userID, repository.UpdateModalityInput{
		MasterID:             req.MasterID,
		UseDefaultGraduation: req.UseDefaultGraduation,
		IsActive:             req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update modality"})
		}
		return
	}
	c.JSON(http.StatusOK, modalityToMap(*item))
}

func (h *AcademyHandler) DeleteModality(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	modalityID, err := uuid.Parse(c.Param("modalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid modality id"})
		return
	}
	userID := middleware.MustGetUserID(c)
	if err := h.repo.DeleteModality(c.Request.Context(), academyID, modalityID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete modality"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AcademyHandler) SetMaster(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	modalityID, err := uuid.Parse(c.Param("modalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid modality id"})
		return
	}
	var req updateMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.UpdateModality(c.Request.Context(), academyID, modalityID, userID, repository.UpdateModalityInput{
		MasterID: req.MasterID,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set master"})
		}
		return
	}
	c.JSON(http.StatusOK, modalityToMap(*item))
}

func (h *AcademyHandler) GetTeachers(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	modalityID, err := uuid.Parse(c.Param("modalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid modality id"})
		return
	}
	items, err := h.repo.ListModalities(c.Request.Context(), academyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teachers"})
		return
	}
	for _, m := range items {
		if m.ID == modalityID {
			c.JSON(http.StatusOK, gin.H{"teacher_ids": m.TeacherIDs})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "modality not found"})
}

func (h *AcademyHandler) SetTeachers(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	modalityID, err := uuid.Parse(c.Param("modalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid modality id"})
		return
	}
	var req setTeachersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.SetModalityTeachers(c.Request.Context(), academyID, modalityID, userID, req.TeacherIDs)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set teachers"})
		}
		return
	}
	c.JSON(http.StatusOK, modalityToMap(*item))
}

func (h *AcademyHandler) UpdateGraduationConfig(c *gin.Context) {
	academyID, err := uuid.Parse(c.Param("academyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid academy id"})
		return
	}
	modalityID, err := uuid.Parse(c.Param("modalityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid modality id"})
		return
	}
	var req updateGraduationConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	cfgs := make([]repository.BeltConfigInput, 0, len(req.BeltConfigs))
	for _, item := range req.BeltConfigs {
		cfgs = append(cfgs, repository.BeltConfigInput{
			BeltID:              item.BeltID,
			BeltName:            item.BeltName,
			MinClasses:          item.MinClasses,
			MinMonths:           item.MinMonths,
			MinClassesPerDegree: item.MinClassesPerDegree,
			RequiresExam:        item.RequiresExam,
			ExamFee:             item.ExamFee,
			Notes:               item.Notes,
		})
	}
	userID := middleware.MustGetUserID(c)
	item, err := h.repo.UpdateGraduationConfig(c.Request.Context(), academyID, modalityID, userID, repository.UpdateGraduationConfigInput{
		UseDefaultGraduation: req.UseDefaultGraduation,
		BeltConfigs:          cfgs,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrModalityNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "modality not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update graduation config"})
		}
		return
	}
	c.JSON(http.StatusOK, modalityToMap(*item))
}

func mapAcademies(in []repository.Academy) []gin.H {
	out := make([]gin.H, 0, len(in))
	for _, item := range in {
		out = append(out, academyToMap(item))
	}
	return out
}

func academyToMap(a repository.Academy) gin.H {
	return gin.H{
		"id":                      a.ID.String(),
		"owner_id":                a.OwnerID.String(),
		"name":                    a.Name,
		"description":             a.Description,
		"logo_url":                a.LogoURL,
		"address":                 a.Address,
		"city":                    a.City,
		"state":                   a.State,
		"phone":                   a.Phone,
		"email":                   a.Email,
		"website":                 a.Website,
		"subscription_plan":       a.SubscriptionPlan,
		"subscription_started_at": a.SubscriptionStartedAt,
		"subscription_ends_at":    a.SubscriptionEndsAt,
		"max_students":            a.MaxStudents,
		"max_teachers":            a.MaxTeachers,
		"max_modalities":          a.MaxModalities,
		"is_active":               a.IsActive,
		"created_at":              a.CreatedAt,
		"updated_at":              a.UpdatedAt,
		"academy_modalities":      mapModalities(a.Modalities),
	}
}

func mapModalities(in []repository.AcademyModality) []gin.H {
	out := make([]gin.H, 0, len(in))
	for _, item := range in {
		out = append(out, modalityToMap(item))
	}
	return out
}

func modalityToMap(m repository.AcademyModality) gin.H {
	belts := make([]gin.H, 0, len(m.BeltConfigs))
	for _, cfg := range m.BeltConfigs {
		belts = append(belts, gin.H{
			"id":                     cfg.ID.String(),
			"modality_id":            cfg.ModalityID.String(),
			"belt_id":                cfg.BeltID,
			"belt_name":              cfg.BeltName,
			"min_classes":            cfg.MinClasses,
			"min_months":             cfg.MinMonths,
			"min_classes_per_degree": cfg.MinClassesPerDegree,
			"requires_exam":          cfg.RequiresExam,
			"exam_fee":               cfg.ExamFee,
			"notes":                  cfg.Notes,
		})
	}
	return gin.H{
		"id":                     m.ID.String(),
		"academy_id":             m.AcademyID.String(),
		"martial_art_type":       m.MartialArtType,
		"master_id":              m.MasterID,
		"use_default_graduation": m.UseDefaultGraduaton,
		"graduation_updated_at":  m.GraduationUpdatedAt,
		"is_active":              m.IsActive,
		"belt_configs":           belts,
		"teacher_ids":            m.TeacherIDs,
	}
}

func academyMemberToMap(m repository.AcademyMemberWithUser) gin.H {
	return gin.H{
		"id":         m.ID.String(),
		"academy_id": m.AcademyID.String(),
		"user_id":    m.UserID.String(),
		"role":       m.Role,
		"status":     m.Status,
		"joined_at":  m.JoinedAt,
		"created_at": m.CreatedAt,
		"updated_at": m.UpdatedAt,
		"user": gin.H{
			"email":        m.UserEmail,
			"display_name": m.UserDisplayName,
			"photo_url":    m.UserPhotoURL,
		},
	}
}

func academyStudentToMap(s repository.AcademyStudent) gin.H {
	modalities := make([]gin.H, 0, len(s.Modalities))
	for _, m := range s.Modalities {
		modalities = append(modalities, studentModalityToMap(m))
	}
	return gin.H{
		"id":         s.Member.ID.String(),
		"academy_id": s.Member.AcademyID.String(),
		"user_id":    s.Member.UserID.String(),
		"role":       s.Member.Role,
		"status":     s.Member.Status,
		"joined_at":  s.Member.JoinedAt,
		"created_at": s.Member.CreatedAt,
		"updated_at": s.Member.UpdatedAt,
		"user": gin.H{
			"email":        s.Member.UserEmail,
			"display_name": s.Member.UserDisplayName,
			"photo_url":    s.Member.UserPhotoURL,
		},
		"modalities": modalities,
	}
}

func studentModalityToMap(m repository.StudentModality) gin.H {
	history := make([]gin.H, 0, len(m.GraduationHistory))
	for _, h := range m.GraduationHistory {
		history = append(history, graduationHistoryToMap(h))
	}
	return gin.H{
		"id":                      m.ID.String(),
		"member_id":               m.MemberID.String(),
		"modality_id":             m.ModalityID.String(),
		"martial_art_type":        m.MartialArtType,
		"assigned_teacher_id":     m.AssignedTeacherID,
		"belt_id":                 m.BeltID,
		"degree":                  m.Degree,
		"promotion_date":          m.PromotionDate,
		"total_classes":           m.TotalClasses,
		"classes_at_current_belt": m.ClassesAtCurrentBelt,
		"enrolled_at":             m.EnrolledAt,
		"graduation_history":      history,
	}
}

func graduationHistoryToMap(h repository.GraduationHistoryItem) gin.H {
	return gin.H{
		"id":                  h.ID.String(),
		"student_modality_id": h.StudentModalityID.String(),
		"belt_id":             h.BeltID,
		"degree":              h.Degree,
		"promoted_at":         h.PromotedAt,
		"promoted_by":         h.PromotedBy,
		"notes":               h.Notes,
	}
}

func checkInToMap(c repository.CheckIn) gin.H {
	return gin.H{
		"id":                  c.ID.String(),
		"student_modality_id": c.StudentModalityID.String(),
		"class_schedule_id":   c.ClassScheduleID,
		"checked_in_at":       c.CheckedInAt,
		"checked_in_by":       c.CheckedInBy,
		"class_type":          c.ClassType,
		"notes":               c.Notes,
		"created_at":          c.CreatedAt,
	}
}

func classScheduleToMap(s repository.ClassSchedule) gin.H {
	return gin.H{
		"id":            s.ID.String(),
		"academy_id":    s.AcademyID.String(),
		"modality_id":   s.ModalityID,
		"instructor_id": s.InstructorID,
		"day_of_week":   s.DayOfWeek,
		"start_time":    s.StartTime,
		"end_time":      s.EndTime,
		"class_type":    s.ClassType,
		"is_active":     s.IsActive,
		"max_students":  s.MaxStudents,
		"notes":         s.Notes,
		"created_at":    s.CreatedAt,
		"updated_at":    s.UpdatedAt,
	}
}

func studentGroupToMap(g repository.StudentGroup) gin.H {
	return gin.H{
		"id":          g.ID.String(),
		"academy_id":  g.AcademyID.String(),
		"name":        g.Name,
		"description": g.Description,
		"is_active":   g.IsActive,
		"created_at":  g.CreatedAt,
		"updated_at":  g.UpdatedAt,
	}
}
