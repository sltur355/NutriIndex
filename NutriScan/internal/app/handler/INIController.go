package handler

import (
	"LAB1/internal/app/middleware"
	"LAB1/internal/app/repository"
	"LAB1/internal/app/role"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type INIController struct {
	INIModel    *repository.INIModel
	JWTSecret   string
	redisClient *redis.Client // Добавляем Redis клиент
}

func NewINIController(r *repository.INIModel, redisClient *redis.Client) *INIController {
	// Получаем JWT секрет
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "fallback-secret-key-change-in-production"
	}

	return &INIController{
		INIModel:    r,
		JWTSecret:   jwtSecret,
		redisClient: redisClient,
	}
}

// RegisterAPI регистрирует маршруты
func (h *INIController) RegisterAPI(router *gin.Engine) {
	router.Use(h.CORSMiddleware())
	api := router.Group("/api")
	{
		// Публичные маршруты (доступны без аутентификации)
		public := api.Group("/")
		{
			// Аутентификация
			public.POST("/auth/login", h.LoginUserAPI)
			public.POST("/auth/register", h.RegisterUserAPI)
			public.POST("/auth/session-login", h.SessionLoginAPI)

			// Просмотр биомаркеров (доступно всем)
			public.GET("/biomarkers", h.GetBiomarkersAPI)
			public.GET("/biomarkers/:id", h.GetBiomarkerAPI)

			public.GET("/researches/cart", h.GetINIResearchCartAPI)
			public.POST("/async/update-ini-result", h.UpdateINIResultAPI)

			public.GET("/biomarkers/compare", h.GetBiomarkersCompareAPI)
		}

		// Защищенные маршруты (требуют аутентификации)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(h.JWTSecret, h.redisClient))
		{
			// Выход
			protected.POST("/auth/logout", h.LogoutUserAPI)

			// Профиль пользователя
			protected.GET("/users/profile", h.GetUserProfileAPI)
			protected.PUT("/users/profile", h.UpdateUserProfileAPI)

			// Исследования - ОДИН маршрут для всех, доступ по ролям внутри метода
			protected.GET("/researches", h.GetINIResearchesAPI)
			protected.GET("/researches/:id", h.GetINIResearchAPI)
			protected.PUT("/researches/:id", h.UpdateINIResearchAPI)
			protected.POST("/biomarkers/:id/add_to_research", h.AddBiomarkerToINIResearchAPI)
			protected.PUT("/research_biomarkers/:research_id/biomarkers/:biomarker_id", h.UpdateResearchBiomarkerAPI)
			protected.DELETE("/research_biomarkers/:research_id/biomarkers/:biomarker_id", h.DeleteResearchBiomarkerAPI)

		}

		// Маршруты только для пациентов
		patient := api.Group("/")
		patient.Use(middleware.AuthMiddleware(h.JWTSecret, h.redisClient))
		patient.Use(middleware.RoleMiddleware(role.Patient))
		{
			// Пациенты могут формировать и удалять только свои исследования
			patient.PUT("/researches/:id/form", h.FormINIResearchAPI)
			patient.DELETE("/researches/:id", h.DeleteINIResearchAPI)
		}

		// Маршруты только для врачей (модераторов)
		doctor := api.Group("/")
		doctor.Use(middleware.AuthMiddleware(h.JWTSecret, h.redisClient))
		doctor.Use(middleware.RoleMiddleware(role.Doctor))
		{
			// Врачи могут завершать/отклонять исследования
			doctor.PUT("/researches/:id/complete", h.CompleteOrRejectINIResearchAPI)

			// Управление биомаркерами (CRUD)
			doctor.POST("/biomarkers", h.CreateBiomarkerAPI)
			doctor.PUT("/biomarkers/:id", h.UpdateBiomarkerAPI)
			doctor.DELETE("/biomarkers/:id", h.DeleteBiomarkerAPI)
			doctor.POST("/biomarkers/:id/image", h.UploadBiomarkerImageAPI)
			doctor.PUT("/researches/:id/complete-async", h.CompleteINIResearchAsyncAPI)
		}
	}
}
func (h *INIController) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"error": err.Error(),
	})
}
