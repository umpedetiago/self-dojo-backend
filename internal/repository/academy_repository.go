package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAcademyNotFound = errors.New("academy not found")
var ErrForbidden = errors.New("forbidden")
var ErrModalityNotFound = errors.New("modality not found")
var ErrMembershipNotFound = errors.New("membership not found")
var ErrMembershipAlreadyPending = errors.New("membership already pending")
var ErrMembershipAlreadyApproved = errors.New("membership already approved")
var ErrStudentNotFound = errors.New("student not found")
var ErrStudentModalityNotFound = errors.New("student modality not found")
var ErrStudentAlreadyEnrolled = errors.New("student already enrolled in modality")
var ErrClassScheduleNotFound = errors.New("class schedule not found")
var ErrStudentGroupNotFound = errors.New("student group not found")

type Academy struct {
	ID                    uuid.UUID
	OwnerID               uuid.UUID
	Name                  string
	Description           string
	LogoURL               string
	Address               string
	City                  string
	State                 string
	Phone                 string
	Email                 string
	Website               string
	SubscriptionPlan      string
	SubscriptionStartedAt *time.Time
	SubscriptionEndsAt    *time.Time
	MaxStudents           int
	MaxTeachers           int
	MaxModalities         int
	IsActive              bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Modalities            []AcademyModality
}

type AcademyModality struct {
	ID                  uuid.UUID
	AcademyID           uuid.UUID
	MartialArtType      string
	MasterID            *uuid.UUID
	UseDefaultGraduaton bool
	GraduationUpdatedAt *time.Time
	IsActive            bool
	BeltConfigs         []BeltConfig
	TeacherIDs          []uuid.UUID
}

type BeltConfig struct {
	ID                  uuid.UUID
	ModalityID          uuid.UUID
	BeltID              string
	BeltName            string
	MinClasses          int
	MinMonths           *int
	MinClassesPerDegree *int
	RequiresExam        bool
	ExamFee             *float64
	Notes               string
}

type AcademyStats struct {
	ApprovedMembers int
	PendingRequests int
}

type AcademyMember struct {
	ID        uuid.UUID
	AcademyID uuid.UUID
	UserID    uuid.UUID
	Role      string
	Status    string
	JoinedAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AcademyMemberWithUser struct {
	AcademyMember
	UserEmail       string
	UserDisplayName string
	UserPhotoURL    string
}

type StudentModality struct {
	ID                   uuid.UUID
	MemberID             uuid.UUID
	ModalityID           uuid.UUID
	MartialArtType       string
	AssignedTeacherID    *uuid.UUID
	BeltID               string
	Degree               int
	PromotionDate        *time.Time
	TotalClasses         int
	ClassesAtCurrentBelt int
	EnrolledAt           time.Time
	GraduationHistory    []GraduationHistoryItem
}

type GraduationHistoryItem struct {
	ID                uuid.UUID
	StudentModalityID uuid.UUID
	BeltID            string
	Degree            int
	PromotedAt        time.Time
	PromotedBy        *uuid.UUID
	Notes             string
}

type AcademyStudent struct {
	Member     AcademyMemberWithUser
	Modalities []StudentModality
}

type EnrollStudentInput struct {
	AcademyModalityID uuid.UUID
	InitialBeltID     string
	InitialDegree     int
}

type UpdateStudentModalityInput struct {
	AssignedTeacherID    *uuid.UUID
	TotalClasses         *int
	ClassesAtCurrentBelt *int
}

type PromoteStudentInput struct {
	NewBeltID string
	Degree    int
	Notes     *string
}

type CreateCheckInInput struct {
	StudentModalityID uuid.UUID
	ClassScheduleID   *uuid.UUID
	ClassType         *string
	Notes             *string
}

type CheckIn struct {
	ID                uuid.UUID
	StudentModalityID uuid.UUID
	ClassScheduleID   *uuid.UUID
	CheckedInAt       time.Time
	CheckedInBy       *uuid.UUID
	ClassType         string
	Notes             string
	CreatedAt         time.Time
}

type ClassSchedule struct {
	ID           uuid.UUID
	AcademyID    uuid.UUID
	ModalityID   *uuid.UUID
	InstructorID *uuid.UUID
	DayOfWeek    int
	StartTime    string
	EndTime      string
	ClassType    string
	IsActive     bool
	MaxStudents  *int
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StudentGroup struct {
	ID          uuid.UUID
	AcademyID   uuid.UUID
	Name        string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateClassScheduleInput struct {
	ModalityID   *uuid.UUID
	InstructorID *uuid.UUID
	DayOfWeek    int
	StartTime    string
	EndTime      string
	ClassType    *string
	IsActive     *bool
	MaxStudents  *int
	Notes        *string
}

type UpdateClassScheduleInput struct {
	ModalityID   *uuid.UUID
	InstructorID *uuid.UUID
	DayOfWeek    *int
	StartTime    *string
	EndTime      *string
	ClassType    *string
	IsActive     *bool
	MaxStudents  *int
	Notes        *string
}

type CreateStudentGroupInput struct {
	Name        string
	Description *string
	IsActive    *bool
}

type UpdateStudentGroupInput struct {
	Name        *string
	Description *string
	IsActive    *bool
}

type CreateAcademyInput struct {
	Name        string
	Description *string
	Address     *string
	City        *string
	State       *string
	Phone       *string
	Email       *string
	Website     *string
	Modalities  []string
}

type UpdateAcademyInput struct {
	Name        *string
	Description *string
	LogoURL     *string
	Address     *string
	City        *string
	State       *string
	Phone       *string
	Email       *string
	Website     *string
	IsActive    *bool
}

type CreateModalityInput struct {
	MartialArtType       string
	MasterID             *uuid.UUID
	UseDefaultGraduation *bool
}

type UpdateModalityInput struct {
	MasterID             *uuid.UUID
	UseDefaultGraduation *bool
	IsActive             *bool
}

type UpdateGraduationConfigInput struct {
	UseDefaultGraduation bool
	BeltConfigs          []BeltConfigInput
}

type BeltConfigInput struct {
	BeltID              string
	BeltName            string
	MinClasses          int
	MinMonths           *int
	MinClassesPerDegree *int
	RequiresExam        bool
	ExamFee             *float64
	Notes               *string
}

type AcademyRepository struct {
	pool *pgxpool.Pool
}

func NewAcademyRepository(pool *pgxpool.Pool) *AcademyRepository {
	return &AcademyRepository{pool: pool}
}

func (r *AcademyRepository) Create(ctx context.Context, ownerID uuid.UUID, in CreateAcademyInput) (*Academy, error) {
	now := time.Now().UTC()
	trialEnd := now.Add(7 * 24 * time.Hour)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
INSERT INTO academies (
  owner_id, name, description, address, city, state, phone, email, website,
  subscription_plan, subscription_started_at, subscription_ends_at, max_students, max_teachers, max_modalities
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'trial',$10,$11,10,3,3)
RETURNING
  id, owner_id, name, COALESCE(description, ''), COALESCE(logo_url, ''), COALESCE(address, ''), COALESCE(city, ''),
  COALESCE(state, ''), COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''), subscription_plan,
  subscription_started_at, subscription_ends_at, max_students, max_teachers, max_modalities, is_active, created_at, updated_at
`, ownerID, strings.TrimSpace(in.Name), in.Description, in.Address, in.City, in.State, in.Phone, in.Email, in.Website, now, trialEnd)

	academy := Academy{}
	if err := row.Scan(
		&academy.ID, &academy.OwnerID, &academy.Name, &academy.Description, &academy.LogoURL, &academy.Address, &academy.City,
		&academy.State, &academy.Phone, &academy.Email, &academy.Website, &academy.SubscriptionPlan,
		&academy.SubscriptionStartedAt, &academy.SubscriptionEndsAt, &academy.MaxStudents, &academy.MaxTeachers, &academy.MaxModalities,
		&academy.IsActive, &academy.CreatedAt, &academy.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert academy: %w", err)
	}

	for _, modality := range in.Modalities {
		mType := strings.TrimSpace(modality)
		if mType == "" {
			continue
		}
		_, err := tx.Exec(ctx, `
INSERT INTO academy_modalities (academy_id, martial_art_type, use_default_graduation, is_active)
VALUES ($1, $2, true, true)
ON CONFLICT (academy_id, martial_art_type) DO NOTHING
`, academy.ID, mType)
		if err != nil {
			return nil, fmt.Errorf("insert modality: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return r.GetByID(ctx, academy.ID)
}

func (r *AcademyRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]Academy, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
  id, owner_id, name, COALESCE(description, ''), COALESCE(logo_url, ''), COALESCE(address, ''), COALESCE(city, ''),
  COALESCE(state, ''), COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''), subscription_plan,
  subscription_started_at, subscription_ends_at, max_students, max_teachers, max_modalities, is_active, created_at, updated_at
FROM academies
WHERE owner_id = $1
ORDER BY created_at ASC
`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list academies by owner: %w", err)
	}
	defer rows.Close()

	var items []Academy
	for rows.Next() {
		var a Academy
		if err := rows.Scan(
			&a.ID, &a.OwnerID, &a.Name, &a.Description, &a.LogoURL, &a.Address, &a.City, &a.State, &a.Phone, &a.Email,
			&a.Website, &a.SubscriptionPlan, &a.SubscriptionStartedAt, &a.SubscriptionEndsAt, &a.MaxStudents, &a.MaxTeachers,
			&a.MaxModalities, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan academy: %w", err)
		}
		modalities, err := r.listModalities(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		a.Modalities = modalities
		items = append(items, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate academies: %w", err)
	}
	return items, nil
}

func (r *AcademyRepository) Search(ctx context.Context, query, city, modality string, limit int) ([]Academy, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := strings.TrimSpace(strings.ToLower(query))
	c := strings.TrimSpace(strings.ToLower(city))
	m := strings.TrimSpace(modality)

	rows, err := r.pool.Query(ctx, `
SELECT DISTINCT
  a.id, a.owner_id, a.name, COALESCE(a.description, ''), COALESCE(a.logo_url, ''), COALESCE(a.address, ''), COALESCE(a.city, ''),
  COALESCE(a.state, ''), COALESCE(a.phone, ''), COALESCE(a.email, ''), COALESCE(a.website, ''), a.subscription_plan,
  a.subscription_started_at, a.subscription_ends_at, a.max_students, a.max_teachers, a.max_modalities, a.is_active, a.created_at, a.updated_at
FROM academies a
LEFT JOIN academy_modalities am ON am.academy_id = a.id
WHERE a.is_active = true
  AND ($1 = '' OR LOWER(a.name) LIKE '%' || $1 || '%' OR LOWER(a.city) LIKE '%' || $1 || '%')
  AND ($2 = '' OR LOWER(a.city) LIKE '%' || $2 || '%')
  AND ($3 = '' OR am.martial_art_type = $3)
ORDER BY a.name ASC
LIMIT $4
`, q, c, m, limit)
	if err != nil {
		return nil, fmt.Errorf("search academies: %w", err)
	}
	defer rows.Close()

	var items []Academy
	for rows.Next() {
		var a Academy
		if err := rows.Scan(
			&a.ID, &a.OwnerID, &a.Name, &a.Description, &a.LogoURL, &a.Address, &a.City, &a.State, &a.Phone, &a.Email,
			&a.Website, &a.SubscriptionPlan, &a.SubscriptionStartedAt, &a.SubscriptionEndsAt, &a.MaxStudents, &a.MaxTeachers,
			&a.MaxModalities, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan academy search: %w", err)
		}
		modalities, err := r.listModalities(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		a.Modalities = modalities
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *AcademyRepository) GetByID(ctx context.Context, academyID uuid.UUID) (*Academy, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  id, owner_id, name, COALESCE(description, ''), COALESCE(logo_url, ''), COALESCE(address, ''), COALESCE(city, ''),
  COALESCE(state, ''), COALESCE(phone, ''), COALESCE(email, ''), COALESCE(website, ''), subscription_plan,
  subscription_started_at, subscription_ends_at, max_students, max_teachers, max_modalities, is_active, created_at, updated_at
FROM academies
WHERE id = $1
`, academyID)

	var a Academy
	if err := row.Scan(
		&a.ID, &a.OwnerID, &a.Name, &a.Description, &a.LogoURL, &a.Address, &a.City, &a.State, &a.Phone, &a.Email,
		&a.Website, &a.SubscriptionPlan, &a.SubscriptionStartedAt, &a.SubscriptionEndsAt, &a.MaxStudents, &a.MaxTeachers,
		&a.MaxModalities, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAcademyNotFound
		}
		return nil, fmt.Errorf("get academy by id: %w", err)
	}

	modalities, err := r.listModalities(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	a.Modalities = modalities
	return &a, nil
}

func (r *AcademyRepository) Update(ctx context.Context, academyID, ownerID uuid.UUID, in UpdateAcademyInput) (*Academy, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE academies
SET
  name = COALESCE($3, name),
  description = COALESCE($4, description),
  logo_url = COALESCE($5, logo_url),
  address = COALESCE($6, address),
  city = COALESCE($7, city),
  state = COALESCE($8, state),
  phone = COALESCE($9, phone),
  email = COALESCE($10, email),
  website = COALESCE($11, website),
  is_active = COALESCE($12, is_active),
  updated_at = now()
WHERE id = $1 AND owner_id = $2
RETURNING id
`, academyID, ownerID, in.Name, in.Description, in.LogoURL, in.Address, in.City, in.State, in.Phone, in.Email, in.Website, in.IsActive)

	var updatedID uuid.UUID
	if err := row.Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
				return nil, err
			}
			return nil, ErrAcademyNotFound
		}
		return nil, fmt.Errorf("update academy: %w", err)
	}
	return r.GetByID(ctx, updatedID)
}

func (r *AcademyRepository) GetStats(ctx context.Context, academyID, ownerID uuid.UUID) (*AcademyStats, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}

	stats := AcademyStats{}
	err := r.pool.QueryRow(ctx, `
SELECT
  COUNT(*) FILTER (WHERE status = 'approved') AS approved_members,
  COUNT(*) FILTER (WHERE status = 'pending') AS pending_requests
FROM academy_members
WHERE academy_id = $1
`, academyID).Scan(&stats.ApprovedMembers, &stats.PendingRequests)
	if err != nil {
		return nil, fmt.Errorf("get academy stats: %w", err)
	}
	return &stats, nil
}

func (r *AcademyRepository) GetMembership(ctx context.Context, academyID, userID uuid.UUID) (*AcademyMember, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, academy_id, user_id, role, status, joined_at, created_at, updated_at
FROM academy_members
WHERE academy_id = $1 AND user_id = $2
`, academyID, userID)

	var m AcademyMember
	if err := row.Scan(&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMembershipNotFound
		}
		return nil, fmt.Errorf("get membership: %w", err)
	}
	return &m, nil
}

func (r *AcademyRepository) GetAcademyMemberByID(ctx context.Context, memberID uuid.UUID) (*AcademyMember, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, academy_id, user_id, role, status, joined_at, created_at, updated_at
FROM academy_members
WHERE id = $1
`, memberID)
	var m AcademyMember
	if err := row.Scan(&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMembershipNotFound
		}
		return nil, fmt.Errorf("get membership by id: %w", err)
	}
	return &m, nil
}

func (r *AcademyRepository) CreateMembershipRequest(ctx context.Context, academyID, userID uuid.UUID) (*AcademyMemberWithUser, error) {
	// Ensure academy exists
	if _, err := r.GetByID(ctx, academyID); err != nil {
		return nil, err
	}

	existing, err := r.GetMembership(ctx, academyID, userID)
	if err == nil {
		if existing.Status == "approved" {
			return nil, ErrMembershipAlreadyApproved
		}
		if existing.Status == "pending" {
			return nil, ErrMembershipAlreadyPending
		}
	}
	if err != nil && !errors.Is(err, ErrMembershipNotFound) {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO academy_members (academy_id, user_id, role, status, created_at, updated_at)
VALUES ($1, $2, 'student', 'pending', now(), now())
RETURNING id
`, academyID, userID)
	var memberID uuid.UUID
	if err := row.Scan(&memberID); err != nil {
		return nil, fmt.Errorf("create membership request: %w", err)
	}
	return r.getMembershipWithUserByID(ctx, memberID)
}

func (r *AcademyRepository) CancelMembershipRequest(ctx context.Context, academyID, userID uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `
DELETE FROM academy_members
WHERE academy_id = $1 AND user_id = $2 AND status = 'pending'
`, academyID, userID)
	if err != nil {
		return fmt.Errorf("cancel membership request: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrMembershipNotFound
	}
	return nil
}

func (r *AcademyRepository) ListMembershipRequests(ctx context.Context, academyID, ownerID uuid.UUID, status string) ([]AcademyMemberWithUser, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	s := strings.TrimSpace(status)
	rows, err := r.pool.Query(ctx, `
SELECT
  am.id, am.academy_id, am.user_id, am.role, am.status, am.joined_at, am.created_at, am.updated_at,
  u.email, COALESCE(u.display_name, ''), COALESCE(u.photo_url, '')
FROM academy_members am
JOIN users u ON u.id = am.user_id
WHERE am.academy_id = $1
  AND ($2 = '' OR am.status = $2)
ORDER BY am.created_at DESC
`, academyID, s)
	if err != nil {
		return nil, fmt.Errorf("list membership requests: %w", err)
	}
	defer rows.Close()

	var out []AcademyMemberWithUser
	for rows.Next() {
		var m AcademyMemberWithUser
		if err := rows.Scan(
			&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
			&m.UserEmail, &m.UserDisplayName, &m.UserPhotoURL,
		); err != nil {
			return nil, fmt.Errorf("scan membership request: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) ApproveMembershipRequest(ctx context.Context, academyID, memberID, ownerID uuid.UUID) (*AcademyMemberWithUser, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
UPDATE academy_members
SET status = 'approved', joined_at = now(), updated_at = now()
WHERE id = $1 AND academy_id = $2
RETURNING id
`, memberID, academyID)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMembershipNotFound
		}
		return nil, fmt.Errorf("approve membership request: %w", err)
	}
	return r.getMembershipWithUserByID(ctx, id)
}

func (r *AcademyRepository) RejectMembershipRequest(ctx context.Context, academyID, memberID, ownerID uuid.UUID) error {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return err
	}
	cmd, err := r.pool.Exec(ctx, `
DELETE FROM academy_members
WHERE id = $1 AND academy_id = $2
`, memberID, academyID)
	if err != nil {
		return fmt.Errorf("reject membership request: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrMembershipNotFound
	}
	return nil
}

func (r *AcademyRepository) ListStudents(ctx context.Context, academyID, ownerID uuid.UUID) ([]AcademyStudent, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
SELECT
  am.id, am.academy_id, am.user_id, am.role, am.status, am.joined_at, am.created_at, am.updated_at,
  u.email, COALESCE(u.display_name, ''), COALESCE(u.photo_url, '')
FROM academy_members am
JOIN users u ON u.id = am.user_id
WHERE am.academy_id = $1 AND am.status = 'approved'
ORDER BY COALESCE(u.display_name, u.email) ASC
`, academyID)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	defer rows.Close()

	var out []AcademyStudent
	for rows.Next() {
		var m AcademyMemberWithUser
		if err := rows.Scan(
			&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
			&m.UserEmail, &m.UserDisplayName, &m.UserPhotoURL,
		); err != nil {
			return nil, fmt.Errorf("scan student member: %w", err)
		}
		modalities, err := r.listStudentModalitiesByMemberID(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, AcademyStudent{Member: m, Modalities: modalities})
	}
	return out, rows.Err()
}

func (r *AcademyRepository) GetStudent(ctx context.Context, academyID, memberID, ownerID uuid.UUID) (*AcademyStudent, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
SELECT
  am.id, am.academy_id, am.user_id, am.role, am.status, am.joined_at, am.created_at, am.updated_at,
  u.email, COALESCE(u.display_name, ''), COALESCE(u.photo_url, '')
FROM academy_members am
JOIN users u ON u.id = am.user_id
WHERE am.id = $1 AND am.academy_id = $2
`, memberID, academyID)
	var m AcademyMemberWithUser
	if err := row.Scan(
		&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
		&m.UserEmail, &m.UserDisplayName, &m.UserPhotoURL,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, fmt.Errorf("get student: %w", err)
	}
	modalities, err := r.listStudentModalitiesByMemberID(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return &AcademyStudent{Member: m, Modalities: modalities}, nil
}

func (r *AcademyRepository) EnrollStudentInModality(ctx context.Context, academyID, memberID, ownerID uuid.UUID, in EnrollStudentInput) (*StudentModality, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	if _, err := r.GetStudent(ctx, academyID, memberID, ownerID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `SELECT id FROM academy_modalities WHERE id = $1 AND academy_id = $2`, in.AcademyModalityID, academyID)
	var modalityID uuid.UUID
	if err := row.Scan(&modalityID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModalityNotFound
		}
		return nil, fmt.Errorf("validate academy modality: %w", err)
	}

	insertRow := r.pool.QueryRow(ctx, `
INSERT INTO student_modalities (
  member_id, modality_id, belt_id, degree, total_classes, classes_at_current_belt, enrolled_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, 0, 0, now(), now(), now())
RETURNING id
`, memberID, modalityID, strings.TrimSpace(in.InitialBeltID), in.InitialDegree)
	var studentModalityID uuid.UUID
	if err := insertRow.Scan(&studentModalityID); err != nil {
		// 23505 = unique violation on (member_id, modality_id)
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, ErrStudentAlreadyEnrolled
		}
		return nil, fmt.Errorf("enroll student in modality: %w", err)
	}
	return r.getStudentModalityByID(ctx, studentModalityID)
}

func (r *AcademyRepository) UpdateStudentModality(ctx context.Context, academyID, memberID, studentModalityID, ownerID uuid.UUID, in UpdateStudentModalityInput) (*StudentModality, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
UPDATE student_modalities sm
SET
  assigned_teacher_id = COALESCE($4::uuid, assigned_teacher_id),
  total_classes = COALESCE($5::integer, total_classes),
  classes_at_current_belt = COALESCE($6::integer, classes_at_current_belt),
  updated_at = now()
FROM academy_members am
WHERE sm.id = $1
  AND sm.member_id = $2
  AND am.id = sm.member_id
  AND am.academy_id = $3
RETURNING sm.id
`, studentModalityID, memberID, academyID, in.AssignedTeacherID, in.TotalClasses, in.ClassesAtCurrentBelt)

	var updatedID uuid.UUID
	if err := row.Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentModalityNotFound
		}
		return nil, fmt.Errorf("update student modality: %w", err)
	}
	return r.getStudentModalityByID(ctx, updatedID)
}

func (r *AcademyRepository) DeleteStudentModality(ctx context.Context, academyID, memberID, studentModalityID, ownerID uuid.UUID) error {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return err
	}
	cmd, err := r.pool.Exec(ctx, `
DELETE FROM student_modalities sm
USING academy_members am
WHERE sm.id = $1
  AND sm.member_id = $2
  AND am.id = sm.member_id
  AND am.academy_id = $3
`, studentModalityID, memberID, academyID)
	if err != nil {
		return fmt.Errorf("delete student modality: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrStudentModalityNotFound
	}
	return nil
}

func (r *AcademyRepository) PromoteStudent(ctx context.Context, studentModalityID, actorUserID uuid.UUID, in PromoteStudentInput) (*StudentModality, error) {
	if err := r.assertCanManageStudentModality(ctx, studentModalityID, actorUserID); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin promotion tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	notes := ""
	if in.Notes != nil {
		notes = *in.Notes
	}

	_, err = tx.Exec(ctx, `
INSERT INTO graduation_history (student_modality_id, belt_id, degree, promoted_at, promoted_by, notes, created_at)
VALUES ($1, $2, $3, now(), $4, $5, now())
`, studentModalityID, strings.TrimSpace(in.NewBeltID), in.Degree, actorUserID, notes)
	if err != nil {
		return nil, fmt.Errorf("insert graduation history: %w", err)
	}

	cmd, err := tx.Exec(ctx, `
UPDATE student_modalities
SET
  belt_id = $2,
  degree = $3,
  promotion_date = now(),
  classes_at_current_belt = 0,
  updated_at = now()
WHERE id = $1
`, studentModalityID, strings.TrimSpace(in.NewBeltID), in.Degree)
	if err != nil {
		return nil, fmt.Errorf("update student modality promotion: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, ErrStudentModalityNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit promotion tx: %w", err)
	}
	return r.getStudentModalityByID(ctx, studentModalityID)
}

func (r *AcademyRepository) ListGraduationHistory(ctx context.Context, studentModalityID, actorUserID uuid.UUID) ([]GraduationHistoryItem, error) {
	if err := r.assertCanManageStudentModality(ctx, studentModalityID, actorUserID); err != nil {
		return nil, err
	}
	return r.listGraduationHistoryByStudentModality(ctx, studentModalityID)
}

func (r *AcademyRepository) CreateCheckIn(ctx context.Context, actorUserID uuid.UUID, in CreateCheckInInput) (*CheckIn, error) {
	if err := r.assertCanAccessStudentModality(ctx, in.StudentModalityID, actorUserID); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin check-in tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := r.assertCheckInGroupRestriction(ctx, tx, in.StudentModalityID, in.ClassScheduleID); err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, `
INSERT INTO check_ins (student_modality_id, class_schedule_id, checked_in_at, checked_in_by, class_type, notes, created_at)
VALUES ($1, $2, now(), $3, $4, $5, now())
RETURNING id
`, in.StudentModalityID, in.ClassScheduleID, actorUserID, in.ClassType, in.Notes)
	var checkInID uuid.UUID
	if err := row.Scan(&checkInID); err != nil {
		return nil, fmt.Errorf("insert check-in: %w", err)
	}

	_, err = tx.Exec(ctx, `
UPDATE student_modalities
SET
  total_classes = total_classes + 1,
  classes_at_current_belt = classes_at_current_belt + 1,
  updated_at = now()
WHERE id = $1
`, in.StudentModalityID)
	if err != nil {
		return nil, fmt.Errorf("increment student classes on check-in: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit check-in tx: %w", err)
	}
	return r.getCheckInByID(ctx, checkInID)
}

func (r *AcademyRepository) ListCheckIns(ctx context.Context, studentModalityID, actorUserID uuid.UUID, startDate, endDate *time.Time) ([]CheckIn, error) {
	if err := r.assertCanAccessStudentModality(ctx, studentModalityID, actorUserID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
SELECT id, student_modality_id, class_schedule_id, checked_in_at, checked_in_by, COALESCE(class_type, ''), COALESCE(notes, ''), created_at
FROM check_ins
WHERE student_modality_id = $1
  AND ($2::timestamptz IS NULL OR checked_in_at >= $2)
  AND ($3::timestamptz IS NULL OR checked_in_at <= $3)
ORDER BY checked_in_at DESC
`, studentModalityID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("list check-ins: %w", err)
	}
	defer rows.Close()

	var out []CheckIn
	for rows.Next() {
		var c CheckIn
		if err := rows.Scan(&c.ID, &c.StudentModalityID, &c.ClassScheduleID, &c.CheckedInAt, &c.CheckedInBy, &c.ClassType, &c.Notes, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan check-in: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) CreateClassSchedule(ctx context.Context, academyID, actorUserID uuid.UUID, in CreateClassScheduleInput) (*ClassSchedule, error) {
	if _, err := r.assertOwnership(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	if in.DayOfWeek < 0 || in.DayOfWeek > 6 {
		return nil, fmt.Errorf("day_of_week must be between 0 and 6")
	}
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	classType := "regular"
	if in.ClassType != nil && strings.TrimSpace(*in.ClassType) != "" {
		classType = strings.TrimSpace(*in.ClassType)
	}
	var notes any = nil
	if in.Notes != nil {
		notes = *in.Notes
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO class_schedules (
  academy_id, modality_id, instructor_id, day_of_week, start_time, end_time, class_type, is_active, max_students, notes, created_at, updated_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now())
RETURNING id
`, academyID, in.ModalityID, in.InstructorID, in.DayOfWeek, strings.TrimSpace(in.StartTime), strings.TrimSpace(in.EndTime), classType, isActive, in.MaxStudents, notes)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create class schedule: %w", err)
	}
	return r.getClassScheduleByID(ctx, id)
}

func (r *AcademyRepository) ListClassSchedules(ctx context.Context, academyID, actorUserID uuid.UUID, modalityID *uuid.UUID, dayOfWeek *int, isActive *bool) ([]ClassSchedule, error) {
	if err := r.assertCanAccessAcademy(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
SELECT
  id, academy_id, modality_id, instructor_id, day_of_week, start_time::text, end_time::text, COALESCE(class_type, ''),
  is_active, max_students, COALESCE(notes, ''), created_at, updated_at
FROM class_schedules
WHERE academy_id = $1
  AND ($2::uuid IS NULL OR modality_id = $2)
  AND ($3::int IS NULL OR day_of_week = $3)
  AND ($4::boolean IS NULL OR is_active = $4)
ORDER BY day_of_week ASC, start_time ASC
`, academyID, modalityID, dayOfWeek, isActive)
	if err != nil {
		return nil, fmt.Errorf("list class schedules: %w", err)
	}
	defer rows.Close()
	var out []ClassSchedule
	for rows.Next() {
		var s ClassSchedule
		if err := rows.Scan(&s.ID, &s.AcademyID, &s.ModalityID, &s.InstructorID, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ClassType, &s.IsActive, &s.MaxStudents, &s.Notes, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan class schedule: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) GetClassSchedule(ctx context.Context, academyID, scheduleID, actorUserID uuid.UUID) (*ClassSchedule, error) {
	if err := r.assertCanAccessAcademy(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
SELECT
  id, academy_id, modality_id, instructor_id, day_of_week, start_time::text, end_time::text, COALESCE(class_type, ''),
  is_active, max_students, COALESCE(notes, ''), created_at, updated_at
FROM class_schedules
WHERE id = $1 AND academy_id = $2
`, scheduleID, academyID)
	var s ClassSchedule
	if err := row.Scan(&s.ID, &s.AcademyID, &s.ModalityID, &s.InstructorID, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ClassType, &s.IsActive, &s.MaxStudents, &s.Notes, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClassScheduleNotFound
		}
		return nil, fmt.Errorf("get class schedule: %w", err)
	}
	return &s, nil
}

func (r *AcademyRepository) UpdateClassSchedule(ctx context.Context, academyID, scheduleID, actorUserID uuid.UUID, in UpdateClassScheduleInput) (*ClassSchedule, error) {
	if _, err := r.assertOwnership(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
UPDATE class_schedules
SET
  modality_id = COALESCE($3::uuid, modality_id),
  instructor_id = COALESCE($4::uuid, instructor_id),
  day_of_week = COALESCE($5::int, day_of_week),
  start_time = COALESCE($6::text, start_time::text)::time,
  end_time = COALESCE($7::text, end_time::text)::time,
  class_type = COALESCE($8::text, class_type),
  is_active = COALESCE($9::boolean, is_active),
  max_students = COALESCE($10::int, max_students),
  notes = COALESCE($11::text, notes),
  updated_at = now()
WHERE id = $1 AND academy_id = $2
RETURNING id
`, scheduleID, academyID, in.ModalityID, in.InstructorID, in.DayOfWeek, in.StartTime, in.EndTime, in.ClassType, in.IsActive, in.MaxStudents, in.Notes)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClassScheduleNotFound
		}
		return nil, fmt.Errorf("update class schedule: %w", err)
	}
	return r.getClassScheduleByID(ctx, id)
}

func (r *AcademyRepository) DeleteClassSchedule(ctx context.Context, academyID, scheduleID, actorUserID uuid.UUID) error {
	if _, err := r.assertOwnership(ctx, academyID, actorUserID); err != nil {
		return err
	}
	cmd, err := r.pool.Exec(ctx, `DELETE FROM class_schedules WHERE id = $1 AND academy_id = $2`, scheduleID, academyID)
	if err != nil {
		return fmt.Errorf("delete class schedule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrClassScheduleNotFound
	}
	return nil
}

func (r *AcademyRepository) ListAvailableSchedulesForCheckIn(ctx context.Context, academyID, actorUserID uuid.UUID, dayOfWeek int) ([]ClassSchedule, error) {
	if err := r.assertCanAccessAcademy(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
SELECT
  id, academy_id, modality_id, instructor_id, day_of_week, start_time::text, end_time::text, COALESCE(class_type, ''),
  is_active, max_students, COALESCE(notes, ''), created_at, updated_at
FROM class_schedules
WHERE academy_id = $1
  AND day_of_week = $2
  AND is_active = true
ORDER BY start_time ASC
`, academyID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("list available schedules for check-in: %w", err)
	}
	defer rows.Close()
	var out []ClassSchedule
	for rows.Next() {
		var s ClassSchedule
		if err := rows.Scan(&s.ID, &s.AcademyID, &s.ModalityID, &s.InstructorID, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ClassType, &s.IsActive, &s.MaxStudents, &s.Notes, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan available class schedule: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) CreateStudentGroup(ctx context.Context, academyID, ownerID uuid.UUID, in CreateStudentGroupInput) (*StudentGroup, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}

	row := r.pool.QueryRow(ctx, `
INSERT INTO student_groups (academy_id, name, description, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, now(), now())
RETURNING id
`, academyID, name, in.Description, isActive)
	var groupID uuid.UUID
	if err := row.Scan(&groupID); err != nil {
		return nil, fmt.Errorf("create student group: %w", err)
	}
	return r.getStudentGroupByID(ctx, groupID)
}

func (r *AcademyRepository) ListStudentGroups(ctx context.Context, academyID, actorUserID uuid.UUID) ([]StudentGroup, error) {
	if err := r.assertCanAccessAcademy(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	return r.listStudentGroupsByAcademy(ctx, academyID)
}

func (r *AcademyRepository) UpdateStudentGroup(ctx context.Context, academyID, groupID, ownerID uuid.UUID, in UpdateStudentGroupInput) (*StudentGroup, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, `
UPDATE student_groups
SET
  name = COALESCE($3::text, name),
  description = COALESCE($4::text, description),
  is_active = COALESCE($5::boolean, is_active),
  updated_at = now()
WHERE id = $1 AND academy_id = $2
RETURNING id
`, groupID, academyID, in.Name, in.Description, in.IsActive)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentGroupNotFound
		}
		return nil, fmt.Errorf("update student group: %w", err)
	}
	return r.getStudentGroupByID(ctx, id)
}

func (r *AcademyRepository) DeleteStudentGroup(ctx context.Context, academyID, groupID, ownerID uuid.UUID) error {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return err
	}
	cmd, err := r.pool.Exec(ctx, `DELETE FROM student_groups WHERE id = $1 AND academy_id = $2`, groupID, academyID)
	if err != nil {
		return fmt.Errorf("delete student group: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrStudentGroupNotFound
	}
	return nil
}

func (r *AcademyRepository) ListStudentGroupMembers(ctx context.Context, academyID, groupID, ownerID uuid.UUID) ([]AcademyMemberWithUser, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	group, err := r.getStudentGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.AcademyID != academyID {
		return nil, ErrStudentGroupNotFound
	}
	return r.listStudentGroupMembersByGroupID(ctx, groupID)
}

func (r *AcademyRepository) SetStudentGroupMembers(ctx context.Context, academyID, groupID, ownerID uuid.UUID, memberIDs []uuid.UUID) ([]AcademyMemberWithUser, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	group, err := r.getStudentGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.AcademyID != academyID {
		return nil, ErrStudentGroupNotFound
	}
	if err := r.validateMemberIDsForAcademy(ctx, academyID, memberIDs); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin set student group members tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM student_group_members WHERE group_id = $1`, groupID); err != nil {
		return nil, fmt.Errorf("clear group members: %w", err)
	}
	for _, memberID := range memberIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO student_group_members (group_id, academy_member_id, created_at)
VALUES ($1, $2, now())
`, groupID, memberID); err != nil {
			return nil, fmt.Errorf("insert group member: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit set student group members tx: %w", err)
	}
	return r.listStudentGroupMembersByGroupID(ctx, groupID)
}

func (r *AcademyRepository) ListClassScheduleGroups(ctx context.Context, academyID, scheduleID, actorUserID uuid.UUID) ([]StudentGroup, error) {
	if err := r.assertCanAccessAcademy(ctx, academyID, actorUserID); err != nil {
		return nil, err
	}
	if err := r.assertClassScheduleInAcademy(ctx, academyID, scheduleID); err != nil {
		return nil, err
	}
	return r.listStudentGroupsByClassScheduleID(ctx, scheduleID)
}

func (r *AcademyRepository) SetClassScheduleGroups(ctx context.Context, academyID, scheduleID, ownerID uuid.UUID, groupIDs []uuid.UUID) ([]StudentGroup, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	if err := r.assertClassScheduleInAcademy(ctx, academyID, scheduleID); err != nil {
		return nil, err
	}
	if err := r.validateGroupIDsForAcademy(ctx, academyID, groupIDs); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin set class schedule groups tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM class_schedule_groups WHERE class_schedule_id = $1`, scheduleID); err != nil {
		return nil, fmt.Errorf("clear class schedule groups: %w", err)
	}
	for _, groupID := range groupIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO class_schedule_groups (class_schedule_id, group_id, created_at)
VALUES ($1, $2, now())
`, scheduleID, groupID); err != nil {
			return nil, fmt.Errorf("insert class schedule group: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit set class schedule groups tx: %w", err)
	}
	return r.listStudentGroupsByClassScheduleID(ctx, scheduleID)
}

func (r *AcademyRepository) ListModalities(ctx context.Context, academyID uuid.UUID) ([]AcademyModality, error) {
	return r.listModalities(ctx, academyID)
}

func (r *AcademyRepository) CreateModality(ctx context.Context, academyID, ownerID uuid.UUID, in CreateModalityInput) (*AcademyModality, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	useDefault := true
	if in.UseDefaultGraduation != nil {
		useDefault = *in.UseDefaultGraduation
	}
	row := r.pool.QueryRow(ctx, `
INSERT INTO academy_modalities (academy_id, martial_art_type, master_id, use_default_graduation, is_active)
VALUES ($1, $2, $3, $4, true)
RETURNING id
`, academyID, strings.TrimSpace(in.MartialArtType), in.MasterID, useDefault)
	var modalityID uuid.UUID
	if err := row.Scan(&modalityID); err != nil {
		return nil, fmt.Errorf("create modality: %w", err)
	}
	return r.getModalityByID(ctx, modalityID)
}

func (r *AcademyRepository) UpdateModality(ctx context.Context, academyID, modalityID, ownerID uuid.UUID, in UpdateModalityInput) (*AcademyModality, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
UPDATE academy_modalities
SET
  master_id = COALESCE($3::uuid, master_id),
  use_default_graduation = COALESCE($4::boolean, use_default_graduation),
  is_active = COALESCE($5::boolean, is_active),
  updated_at = now()
WHERE id = $1 AND academy_id = $2
RETURNING id
`, modalityID, academyID, in.MasterID, in.UseDefaultGraduation, in.IsActive)

	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModalityNotFound
		}
		return nil, fmt.Errorf("update modality: %w", err)
	}
	return r.getModalityByID(ctx, id)
}

func (r *AcademyRepository) DeleteModality(ctx context.Context, academyID, modalityID, ownerID uuid.UUID) error {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return err
	}
	cmd, err := r.pool.Exec(ctx, `DELETE FROM academy_modalities WHERE id = $1 AND academy_id = $2`, modalityID, academyID)
	if err != nil {
		return fmt.Errorf("delete modality: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrModalityNotFound
	}
	return nil
}

func (r *AcademyRepository) SetModalityTeachers(ctx context.Context, academyID, modalityID, ownerID uuid.UUID, teacherIDs []uuid.UUID) (*AcademyModality, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx teachers: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM modality_teachers WHERE modality_id = $1 AND role = 'teacher'`, modalityID); err != nil {
		return nil, fmt.Errorf("clear modality teachers: %w", err)
	}
	for _, tID := range teacherIDs {
		if _, err := tx.Exec(ctx, `
INSERT INTO modality_teachers (modality_id, user_id, role)
VALUES ($1, $2, 'teacher')
ON CONFLICT (modality_id, user_id) DO UPDATE SET role = 'teacher'
`, modalityID, tID); err != nil {
			return nil, fmt.Errorf("insert modality teacher: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit teachers tx: %w", err)
	}
	return r.getModalityByID(ctx, modalityID)
}

func (r *AcademyRepository) UpdateGraduationConfig(ctx context.Context, academyID, modalityID, ownerID uuid.UUID, in UpdateGraduationConfigInput) (*AcademyModality, error) {
	if _, err := r.assertOwnership(ctx, academyID, ownerID); err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx graduation config: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cmd, err := tx.Exec(ctx, `
UPDATE academy_modalities
SET use_default_graduation = $3, graduation_updated_at = now(), updated_at = now()
WHERE id = $1 AND academy_id = $2
`, modalityID, academyID, in.UseDefaultGraduation)
	if err != nil {
		return nil, fmt.Errorf("update modality graduation config: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, ErrModalityNotFound
	}

	if _, err := tx.Exec(ctx, `DELETE FROM belt_configs WHERE modality_id = $1`, modalityID); err != nil {
		return nil, fmt.Errorf("clear belt configs: %w", err)
	}
	for _, cfg := range in.BeltConfigs {
		notes := ""
		if cfg.Notes != nil {
			notes = *cfg.Notes
		}
		_, err := tx.Exec(ctx, `
INSERT INTO belt_configs (modality_id, belt_id, belt_name, min_classes, min_months, min_classes_per_degree, requires_exam, exam_fee, notes)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`, modalityID, strings.TrimSpace(cfg.BeltID), strings.TrimSpace(cfg.BeltName), cfg.MinClasses, cfg.MinMonths, cfg.MinClassesPerDegree, cfg.RequiresExam, cfg.ExamFee, notes)
		if err != nil {
			return nil, fmt.Errorf("insert belt config: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit graduation tx: %w", err)
	}
	return r.getModalityByID(ctx, modalityID)
}

func (r *AcademyRepository) assertOwnership(ctx context.Context, academyID, ownerID uuid.UUID) (*Academy, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, owner_id FROM academies WHERE id = $1`, academyID)
	var aID, oID uuid.UUID
	if err := row.Scan(&aID, &oID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAcademyNotFound
		}
		return nil, fmt.Errorf("assert ownership: %w", err)
	}
	if oID != ownerID {
		return nil, ErrForbidden
	}
	return &Academy{ID: aID, OwnerID: oID}, nil
}

func (r *AcademyRepository) assertCanAccessAcademy(ctx context.Context, academyID, actorUserID uuid.UUID) error {
	row := r.pool.QueryRow(ctx, `SELECT owner_id FROM academies WHERE id = $1`, academyID)
	var ownerID uuid.UUID
	if err := row.Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAcademyNotFound
		}
		return fmt.Errorf("assert academy access owner lookup: %w", err)
	}
	if ownerID == actorUserID {
		return nil
	}

	memberRow := r.pool.QueryRow(ctx, `
SELECT 1
FROM academy_members
WHERE academy_id = $1 AND user_id = $2 AND status = 'approved'
`, academyID, actorUserID)
	var one int
	if err := memberRow.Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("assert academy access member lookup: %w", err)
	}
	return nil
}

func (r *AcademyRepository) listModalities(ctx context.Context, academyID uuid.UUID) ([]AcademyModality, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, academy_id, martial_art_type, master_id, use_default_graduation, graduation_updated_at, is_active
FROM academy_modalities
WHERE academy_id = $1
ORDER BY martial_art_type ASC
`, academyID)
	if err != nil {
		return nil, fmt.Errorf("list modalities: %w", err)
	}
	defer rows.Close()

	var out []AcademyModality
	for rows.Next() {
		var m AcademyModality
		if err := rows.Scan(&m.ID, &m.AcademyID, &m.MartialArtType, &m.MasterID, &m.UseDefaultGraduaton, &m.GraduationUpdatedAt, &m.IsActive); err != nil {
			return nil, fmt.Errorf("scan modality: %w", err)
		}
		beltCfgs, err := r.listBeltConfigs(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		m.BeltConfigs = beltCfgs
		teachers, err := r.listTeacherIDs(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		m.TeacherIDs = teachers
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) getModalityByID(ctx context.Context, modalityID uuid.UUID) (*AcademyModality, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, academy_id, martial_art_type, master_id, use_default_graduation, graduation_updated_at, is_active
FROM academy_modalities
WHERE id = $1
`, modalityID)
	var m AcademyModality
	if err := row.Scan(&m.ID, &m.AcademyID, &m.MartialArtType, &m.MasterID, &m.UseDefaultGraduaton, &m.GraduationUpdatedAt, &m.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModalityNotFound
		}
		return nil, fmt.Errorf("get modality by id: %w", err)
	}
	beltCfgs, err := r.listBeltConfigs(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	m.BeltConfigs = beltCfgs
	teachers, err := r.listTeacherIDs(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	m.TeacherIDs = teachers
	return &m, nil
}

func (r *AcademyRepository) getMembershipWithUserByID(ctx context.Context, memberID uuid.UUID) (*AcademyMemberWithUser, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  am.id, am.academy_id, am.user_id, am.role, am.status, am.joined_at, am.created_at, am.updated_at,
  u.email, COALESCE(u.display_name, ''), COALESCE(u.photo_url, '')
FROM academy_members am
JOIN users u ON u.id = am.user_id
WHERE am.id = $1
`, memberID)
	var m AcademyMemberWithUser
	if err := row.Scan(
		&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
		&m.UserEmail, &m.UserDisplayName, &m.UserPhotoURL,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMembershipNotFound
		}
		return nil, fmt.Errorf("get membership with user by id: %w", err)
	}
	return &m, nil
}

func (r *AcademyRepository) getStudentModalityByID(ctx context.Context, studentModalityID uuid.UUID) (*StudentModality, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  sm.id, sm.member_id, sm.modality_id, am.martial_art_type, sm.assigned_teacher_id, sm.belt_id, sm.degree,
  sm.promotion_date, sm.total_classes, sm.classes_at_current_belt, sm.enrolled_at
FROM student_modalities sm
JOIN academy_modalities am ON am.id = sm.modality_id
WHERE sm.id = $1
`, studentModalityID)
	var s StudentModality
	if err := row.Scan(
		&s.ID, &s.MemberID, &s.ModalityID, &s.MartialArtType, &s.AssignedTeacherID, &s.BeltID, &s.Degree,
		&s.PromotionDate, &s.TotalClasses, &s.ClassesAtCurrentBelt, &s.EnrolledAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentModalityNotFound
		}
		return nil, fmt.Errorf("get student modality by id: %w", err)
	}
	history, err := r.listGraduationHistoryByStudentModality(ctx, s.ID)
	if err != nil {
		return nil, err
	}
	s.GraduationHistory = history
	return &s, nil
}

func (r *AcademyRepository) listStudentModalitiesByMemberID(ctx context.Context, memberID uuid.UUID) ([]StudentModality, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
  sm.id, sm.member_id, sm.modality_id, am.martial_art_type, sm.assigned_teacher_id, sm.belt_id, sm.degree,
  sm.promotion_date, sm.total_classes, sm.classes_at_current_belt, sm.enrolled_at
FROM student_modalities sm
JOIN academy_modalities am ON am.id = sm.modality_id
WHERE sm.member_id = $1
ORDER BY am.martial_art_type ASC
`, memberID)
	if err != nil {
		return nil, fmt.Errorf("list student modalities by member: %w", err)
	}
	defer rows.Close()

	var out []StudentModality
	for rows.Next() {
		var s StudentModality
		if err := rows.Scan(
			&s.ID, &s.MemberID, &s.ModalityID, &s.MartialArtType, &s.AssignedTeacherID, &s.BeltID, &s.Degree,
			&s.PromotionDate, &s.TotalClasses, &s.ClassesAtCurrentBelt, &s.EnrolledAt,
		); err != nil {
			return nil, fmt.Errorf("scan student modality: %w", err)
		}
		history, err := r.listGraduationHistoryByStudentModality(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		s.GraduationHistory = history
		out = append(out, s)
	}
	return out, rows.Err()
}

// ListUserStudentModalities retorna todas as modalidades em que o usuário está
// matriculado (status approved), já com histórico de graduações carregado.
func (r *AcademyRepository) ListUserStudentModalities(ctx context.Context, userID uuid.UUID) ([]StudentModality, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
  sm.id, sm.member_id, sm.modality_id, am.martial_art_type, sm.assigned_teacher_id, sm.belt_id, sm.degree,
  sm.promotion_date, sm.total_classes, sm.classes_at_current_belt, sm.enrolled_at
FROM academy_members m
JOIN student_modalities sm ON sm.member_id = m.id
JOIN academy_modalities am ON am.id = sm.modality_id
WHERE m.user_id = $1
  AND m.status = 'approved'
ORDER BY am.martial_art_type ASC, sm.enrolled_at ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list student modalities by user: %w", err)
	}
	defer rows.Close()

	var out []StudentModality
	for rows.Next() {
		var s StudentModality
		if err := rows.Scan(
			&s.ID, &s.MemberID, &s.ModalityID, &s.MartialArtType, &s.AssignedTeacherID, &s.BeltID, &s.Degree,
			&s.PromotionDate, &s.TotalClasses, &s.ClassesAtCurrentBelt, &s.EnrolledAt,
		); err != nil {
			return nil, fmt.Errorf("scan student modality by user: %w", err)
		}
		history, err := r.listGraduationHistoryByStudentModality(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		s.GraduationHistory = history
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) listGraduationHistoryByStudentModality(ctx context.Context, studentModalityID uuid.UUID) ([]GraduationHistoryItem, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, student_modality_id, belt_id, degree, promoted_at, promoted_by, COALESCE(notes, '')
FROM graduation_history
WHERE student_modality_id = $1
ORDER BY promoted_at DESC
`, studentModalityID)
	if err != nil {
		return nil, fmt.Errorf("list graduation history: %w", err)
	}
	defer rows.Close()
	var out []GraduationHistoryItem
	for rows.Next() {
		var item GraduationHistoryItem
		if err := rows.Scan(&item.ID, &item.StudentModalityID, &item.BeltID, &item.Degree, &item.PromotedAt, &item.PromotedBy, &item.Notes); err != nil {
			return nil, fmt.Errorf("scan graduation history: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) assertCanManageStudentModality(ctx context.Context, studentModalityID, actorUserID uuid.UUID) error {
	row := r.pool.QueryRow(ctx, `
SELECT a.owner_id
FROM student_modalities sm
JOIN academy_members am ON am.id = sm.member_id
JOIN academies a ON a.id = am.academy_id
WHERE sm.id = $1
`, studentModalityID)
	var ownerID uuid.UUID
	if err := row.Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStudentModalityNotFound
		}
		return fmt.Errorf("assert manage student modality: %w", err)
	}
	if ownerID != actorUserID {
		return ErrForbidden
	}
	return nil
}

func (r *AcademyRepository) assertCanAccessStudentModality(ctx context.Context, studentModalityID, actorUserID uuid.UUID) error {
	row := r.pool.QueryRow(ctx, `
SELECT a.owner_id, am.user_id
FROM student_modalities sm
JOIN academy_members am ON am.id = sm.member_id
JOIN academies a ON a.id = am.academy_id
WHERE sm.id = $1
`, studentModalityID)
	var ownerID uuid.UUID
	var memberUserID uuid.UUID
	if err := row.Scan(&ownerID, &memberUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStudentModalityNotFound
		}
		return fmt.Errorf("assert access student modality: %w", err)
	}
	if ownerID != actorUserID && memberUserID != actorUserID {
		return ErrForbidden
	}
	return nil
}

func (r *AcademyRepository) assertCheckInGroupRestriction(ctx context.Context, tx pgx.Tx, studentModalityID uuid.UUID, classScheduleID *uuid.UUID) error {
	if classScheduleID == nil {
		return nil
	}

	var academyMemberID uuid.UUID
	var studentAcademyID uuid.UUID
	row := tx.QueryRow(ctx, `
SELECT sm.member_id, am.academy_id
FROM student_modalities sm
JOIN academy_members am ON am.id = sm.member_id
WHERE sm.id = $1
`, studentModalityID)
	if err := row.Scan(&academyMemberID, &studentAcademyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStudentModalityNotFound
		}
		return fmt.Errorf("check check-in student group restriction (student): %w", err)
	}

	var scheduleAcademyID uuid.UUID
	row = tx.QueryRow(ctx, `SELECT academy_id FROM class_schedules WHERE id = $1`, *classScheduleID)
	if err := row.Scan(&scheduleAcademyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrClassScheduleNotFound
		}
		return fmt.Errorf("check check-in student group restriction (schedule): %w", err)
	}
	if scheduleAcademyID != studentAcademyID {
		return ErrForbidden
	}

	var restrictionCount int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
FROM class_schedule_groups
WHERE class_schedule_id = $1
`, *classScheduleID).Scan(&restrictionCount); err != nil {
		return fmt.Errorf("check class schedule group restrictions: %w", err)
	}
	if restrictionCount == 0 {
		return nil
	}

	var one int
	row = tx.QueryRow(ctx, `
SELECT 1
FROM class_schedule_groups csg
JOIN student_group_members sgm ON sgm.group_id = csg.group_id
WHERE csg.class_schedule_id = $1
  AND sgm.academy_member_id = $2
LIMIT 1
`, *classScheduleID, academyMemberID)
	if err := row.Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("check student group membership for check-in: %w", err)
	}
	return nil
}

func (r *AcademyRepository) assertClassScheduleInAcademy(ctx context.Context, academyID, scheduleID uuid.UUID) error {
	var foundID uuid.UUID
	row := r.pool.QueryRow(ctx, `SELECT id FROM class_schedules WHERE id = $1 AND academy_id = $2`, scheduleID, academyID)
	if err := row.Scan(&foundID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrClassScheduleNotFound
		}
		return fmt.Errorf("assert class schedule in academy: %w", err)
	}
	return nil
}

func (r *AcademyRepository) validateMemberIDsForAcademy(ctx context.Context, academyID uuid.UUID, memberIDs []uuid.UUID) error {
	if len(memberIDs) == 0 {
		return nil
	}
	var count int
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM academy_members
WHERE academy_id = $1
  AND id = ANY($2::uuid[])
  AND role = 'student'
  AND status = 'approved'
`, academyID, memberIDs).Scan(&count)
	if err != nil {
		return fmt.Errorf("validate student group member ids: %w", err)
	}
	if count != len(memberIDs) {
		return ErrStudentNotFound
	}
	return nil
}

func (r *AcademyRepository) validateGroupIDsForAcademy(ctx context.Context, academyID uuid.UUID, groupIDs []uuid.UUID) error {
	if len(groupIDs) == 0 {
		return nil
	}
	var count int
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM student_groups
WHERE academy_id = $1
  AND id = ANY($2::uuid[])
`, academyID, groupIDs).Scan(&count)
	if err != nil {
		return fmt.Errorf("validate class schedule group ids: %w", err)
	}
	if count != len(groupIDs) {
		return ErrStudentGroupNotFound
	}
	return nil
}

func (r *AcademyRepository) getCheckInByID(ctx context.Context, checkInID uuid.UUID) (*CheckIn, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, student_modality_id, class_schedule_id, checked_in_at, checked_in_by, COALESCE(class_type, ''), COALESCE(notes, ''), created_at
FROM check_ins
WHERE id = $1
`, checkInID)
	var c CheckIn
	if err := row.Scan(&c.ID, &c.StudentModalityID, &c.ClassScheduleID, &c.CheckedInAt, &c.CheckedInBy, &c.ClassType, &c.Notes, &c.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentModalityNotFound
		}
		return nil, fmt.Errorf("get check-in by id: %w", err)
	}
	return &c, nil
}

func (r *AcademyRepository) getClassScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*ClassSchedule, error) {
	row := r.pool.QueryRow(ctx, `
SELECT
  id, academy_id, modality_id, instructor_id, day_of_week, start_time::text, end_time::text, COALESCE(class_type, ''),
  is_active, max_students, COALESCE(notes, ''), created_at, updated_at
FROM class_schedules
WHERE id = $1
`, scheduleID)
	var s ClassSchedule
	if err := row.Scan(&s.ID, &s.AcademyID, &s.ModalityID, &s.InstructorID, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.ClassType, &s.IsActive, &s.MaxStudents, &s.Notes, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClassScheduleNotFound
		}
		return nil, fmt.Errorf("get class schedule by id: %w", err)
	}
	return &s, nil
}

func (r *AcademyRepository) getStudentGroupByID(ctx context.Context, groupID uuid.UUID) (*StudentGroup, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, academy_id, name, COALESCE(description, ''), is_active, created_at, updated_at
FROM student_groups
WHERE id = $1
`, groupID)
	var g StudentGroup
	if err := row.Scan(&g.ID, &g.AcademyID, &g.Name, &g.Description, &g.IsActive, &g.CreatedAt, &g.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentGroupNotFound
		}
		return nil, fmt.Errorf("get student group by id: %w", err)
	}
	return &g, nil
}

func (r *AcademyRepository) listStudentGroupsByAcademy(ctx context.Context, academyID uuid.UUID) ([]StudentGroup, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, academy_id, name, COALESCE(description, ''), is_active, created_at, updated_at
FROM student_groups
WHERE academy_id = $1
ORDER BY name ASC
`, academyID)
	if err != nil {
		return nil, fmt.Errorf("list student groups by academy: %w", err)
	}
	defer rows.Close()

	out := make([]StudentGroup, 0)
	for rows.Next() {
		var g StudentGroup
		if err := rows.Scan(&g.ID, &g.AcademyID, &g.Name, &g.Description, &g.IsActive, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan student group: %w", err)
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) listStudentGroupsByClassScheduleID(ctx context.Context, scheduleID uuid.UUID) ([]StudentGroup, error) {
	rows, err := r.pool.Query(ctx, `
SELECT sg.id, sg.academy_id, sg.name, COALESCE(sg.description, ''), sg.is_active, sg.created_at, sg.updated_at
FROM class_schedule_groups csg
JOIN student_groups sg ON sg.id = csg.group_id
WHERE csg.class_schedule_id = $1
ORDER BY sg.name ASC
`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("list class schedule groups: %w", err)
	}
	defer rows.Close()

	out := make([]StudentGroup, 0)
	for rows.Next() {
		var g StudentGroup
		if err := rows.Scan(&g.ID, &g.AcademyID, &g.Name, &g.Description, &g.IsActive, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan class schedule group: %w", err)
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) listStudentGroupMembersByGroupID(ctx context.Context, groupID uuid.UUID) ([]AcademyMemberWithUser, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
  am.id, am.academy_id, am.user_id, am.role, am.status, am.joined_at, am.created_at, am.updated_at,
  u.email, COALESCE(u.display_name, ''), COALESCE(u.photo_url, '')
FROM student_group_members sgm
JOIN academy_members am ON am.id = sgm.academy_member_id
JOIN users u ON u.id = am.user_id
WHERE sgm.group_id = $1
ORDER BY COALESCE(u.display_name, u.email) ASC
`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list student group members: %w", err)
	}
	defer rows.Close()

	out := make([]AcademyMemberWithUser, 0)
	for rows.Next() {
		var m AcademyMemberWithUser
		if err := rows.Scan(
			&m.ID, &m.AcademyID, &m.UserID, &m.Role, &m.Status, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
			&m.UserEmail, &m.UserDisplayName, &m.UserPhotoURL,
		); err != nil {
			return nil, fmt.Errorf("scan student group member: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) listBeltConfigs(ctx context.Context, modalityID uuid.UUID) ([]BeltConfig, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, modality_id, belt_id, COALESCE(belt_name, ''), min_classes, min_months, min_classes_per_degree, requires_exam, exam_fee, COALESCE(notes, '')
FROM belt_configs
WHERE modality_id = $1
ORDER BY belt_id ASC
`, modalityID)
	if err != nil {
		return nil, fmt.Errorf("list belt configs: %w", err)
	}
	defer rows.Close()
	var out []BeltConfig
	for rows.Next() {
		var b BeltConfig
		if err := rows.Scan(&b.ID, &b.ModalityID, &b.BeltID, &b.BeltName, &b.MinClasses, &b.MinMonths, &b.MinClassesPerDegree, &b.RequiresExam, &b.ExamFee, &b.Notes); err != nil {
			return nil, fmt.Errorf("scan belt config: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *AcademyRepository) listTeacherIDs(ctx context.Context, modalityID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
SELECT user_id
FROM modality_teachers
WHERE modality_id = $1 AND role = 'teacher'
ORDER BY user_id
`, modalityID)
	if err != nil {
		return nil, fmt.Errorf("list teacher ids: %w", err)
	}
	defer rows.Close()
	var out []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan teacher id: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
