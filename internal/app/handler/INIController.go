package handler

import (
	"LAB1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type INIController struct {
	INIModel *repository.INIModel // ← ИСПРАВЛЕНО: INIModel вместо Repository
}

func NewINIController(r *repository.INIModel) *INIController {
	return &INIController{
		INIModel: r, // ← ИСПРАВЛЕНО
	}
}

// RegisterHandler регистрирует маршруты
func (h *INIController) RegisterHandler(router *gin.Engine) {
	// GET методы
	router.GET("/biomarkers", h.GetBiomarkers)
	router.GET("/biomarkers/:id", h.GetDetailedBiomarker)
	router.GET("/INIresearch/:id", h.GetINIresearch)

	// POST методы согласно заданию
	router.POST("/INIresearch/:id/biomarkers", h.AddBiomarkerToResearch) // Добавление услуги в заявку (ORM)
	router.POST("/INIresearch/:id/delete", h.DeleteResearch)             // Логическое удаление заявки (SQL)
}

// RegisterStatic регистрирует статику и шаблоны
func (h *INIController) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
}

// errorHandler для удобного вывода ошибок
func (h *INIController) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
