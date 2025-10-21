package handler

import (
	"LAB1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type INIController struct {
	INIModel *repository.INIModel
}

func NewINIController(r *repository.INIModel) *INIController {
	return &INIController{
		INIModel: r,
	}
}

// RegisterHandler регистрирует маршруты
func (h *INIController) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Домен биомаркеров
		biomarkers := api.Group("/biomarkers")
		{
			biomarkers.GET("", h.GetBiomarkersAPI)
			biomarkers.GET("/:id", h.GetBiomarkerAPI)
			biomarkers.POST("", h.CreateBiomarkerAPI)
			biomarkers.PUT("/:id", h.UpdateBiomarkerAPI)
			biomarkers.DELETE("/:id", h.DeleteBiomarkerAPI)
			biomarkers.POST("/:id/add_to_research", h.AddBiomarkerToINIResearchAPI)
			biomarkers.POST("/:id/image", h.UploadBiomarkerImageAPI)
		}

		// Домен исследований
		researches := api.Group("/researches")
		{
			researches.GET("/cart", h.GetINIResearchCartAPI)
			researches.GET("", h.GetINIResearchesAPI)
			researches.GET("/:id", h.GetINIResearchAPI)
			researches.PUT("/:id", h.UpdateINIResearchAPI)
			researches.PUT("/:id/form", h.FormINIResearchAPI)
			researches.PUT("/:id/complete", h.CompleteOrRejectINIResearchAPI)
			researches.DELETE("/:id", h.DeleteINIResearchAPI)
		}

		// Домен М-М (ResearchBiomarker)
		researchBiomarkers := api.Group("/research_biomarkers")
		{
			researchBiomarkers.PUT("/:research_id/biomarkers/:biomarker_id", h.UpdateResearchBiomarkerAPI)
			researchBiomarkers.DELETE("/:research_id/biomarkers/:biomarker_id", h.DeleteResearchBiomarkerAPI)
		}

		// Домен пользователей
		users := api.Group("/users")
		{
			users.POST("/register", h.RegisterUserAPI)
			users.POST("/auth", h.AuthenticateUserAPI)
			users.POST("/logout", h.LogoutUserAPI)
			users.GET("/profile", h.GetUserProfileAPI)
			users.PUT("/profile", h.UpdateUserProfileAPI)
		}
	}
}
func (h *INIController) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
