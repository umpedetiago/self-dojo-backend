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
	"go-api/internal/storage"
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
	academyRepo := repository.NewAcademyRepository(pool)
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

	var avatarStorage storage.Provider
	if cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != "" {
		supabaseStorage, err := storage.NewSupabaseProvider(storage.SupabaseConfig{
			BaseURL:        cfg.SupabaseURL,
			ServiceRoleKey: cfg.SupabaseServiceRoleKey,
			Bucket:         cfg.AvatarBucket,
		})
		if err != nil {
			log.Fatal(err)
		}
		avatarStorage = supabaseStorage
	}

	authHandler := handlers.NewAuthHandler(userRepo, passwordResetRepo, emailSender, avatarStorage, cfg.JWTSecret)
	academyHandler := handlers.NewAcademyHandler(academyRepo)

	auth := server.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", middleware.AuthRequired(cfg.JWTSecret), authHandler.Logout)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	server.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)
	server.PUT("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Update)
	server.GET("/v1/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)
	server.PUT("/v1/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Update)
	server.POST("/v1/me/avatar", middleware.AuthRequired(cfg.JWTSecret), authHandler.UploadAvatar)

	v1 := server.Group("/v1", middleware.AuthRequired(cfg.JWTSecret))
	{
		v1.GET("/academies", academyHandler.ListOwnerAcademies)
		v1.POST("/academies", academyHandler.CreateAcademy)
		v1.GET("/academies/search", academyHandler.SearchAcademies)
		v1.GET("/academies/:academyId", academyHandler.GetAcademy)
		v1.PATCH("/academies/:academyId", academyHandler.UpdateAcademy)
		v1.GET("/academies/:academyId/stats", academyHandler.GetAcademyStats)
		v1.GET("/academies/:academyId/membership/me", academyHandler.GetMyMembershipStatus)
		v1.POST("/academies/:academyId/membership-requests", academyHandler.CreateMembershipRequest)
		v1.DELETE("/academies/:academyId/membership-requests/me", academyHandler.CancelMyMembershipRequest)
		v1.GET("/academies/:academyId/membership-requests", academyHandler.ListMembershipRequests)
		v1.POST("/academies/:academyId/membership-requests/:memberId/approve", academyHandler.ApproveMembershipRequest)
		v1.POST("/academies/:academyId/membership-requests/:memberId/reject", academyHandler.RejectMembershipRequest)
		v1.GET("/academies/:academyId/students", academyHandler.ListStudents)
		v1.GET("/academies/:academyId/students/:memberId", academyHandler.GetStudent)
		v1.POST("/academies/:academyId/students/:memberId/modalities", academyHandler.EnrollStudentInModality)
		v1.PATCH("/academies/:academyId/students/:memberId/modalities/:studentModalityId", academyHandler.UpdateStudentModality)
		v1.DELETE("/academies/:academyId/students/:memberId/modalities/:studentModalityId", academyHandler.DeleteStudentModality)
		v1.POST("/student-modalities/:studentModalityId/promotions", academyHandler.PromoteStudent)
		v1.GET("/student-modalities/:studentModalityId/graduation-history", academyHandler.ListGraduationHistory)
		v1.GET("/student-modalities/:studentModalityId/check-ins", academyHandler.ListCheckIns)
		v1.POST("/check-ins", academyHandler.CreateCheckIn)
		v1.POST("/academies/:academyId/class-schedules", academyHandler.CreateClassSchedule)
		v1.GET("/academies/:academyId/class-schedules", academyHandler.ListClassSchedules)
		v1.GET("/academies/:academyId/class-schedules/available-for-checkin", academyHandler.ListAvailableSchedulesForCheckIn)
		v1.GET("/academies/:academyId/class-schedules/:scheduleId", academyHandler.GetClassSchedule)
		v1.PATCH("/academies/:academyId/class-schedules/:scheduleId", academyHandler.UpdateClassSchedule)
		v1.DELETE("/academies/:academyId/class-schedules/:scheduleId", academyHandler.DeleteClassSchedule)
		v1.GET("/academies/:academyId/class-schedules/:scheduleId/groups", academyHandler.ListClassScheduleGroups)
		v1.PUT("/academies/:academyId/class-schedules/:scheduleId/groups", academyHandler.SetClassScheduleGroups)

		v1.GET("/academies/:academyId/student-groups", academyHandler.ListStudentGroups)
		v1.POST("/academies/:academyId/student-groups", academyHandler.CreateStudentGroup)
		v1.PATCH("/academies/:academyId/student-groups/:groupId", academyHandler.UpdateStudentGroup)
		v1.DELETE("/academies/:academyId/student-groups/:groupId", academyHandler.DeleteStudentGroup)
		v1.GET("/academies/:academyId/student-groups/:groupId/members", academyHandler.ListStudentGroupMembers)
		v1.PUT("/academies/:academyId/student-groups/:groupId/members", academyHandler.SetStudentGroupMembers)

		v1.POST("/academies/:academyId/modalities", academyHandler.CreateModality)
		v1.GET("/academies/:academyId/modalities", academyHandler.ListModalities)
		v1.PATCH("/academies/:academyId/modalities/:modalityId", academyHandler.UpdateModality)
		v1.DELETE("/academies/:academyId/modalities/:modalityId", academyHandler.DeleteModality)
		v1.PUT("/academies/:academyId/modalities/:modalityId/master", academyHandler.SetMaster)
		v1.PUT("/academies/:academyId/modalities/:modalityId/graduation-config", academyHandler.UpdateGraduationConfig)
		v1.GET("/academies/:academyId/modalities/:modalityId/teachers", academyHandler.GetTeachers)
		v1.PUT("/academies/:academyId/modalities/:modalityId/teachers", academyHandler.SetTeachers)
	}

	if err := server.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}

}
