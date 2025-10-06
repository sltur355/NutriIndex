package handler

import (
	"LAB1/internal/app/repository"
	"net/http"
	"strconv"
	"time"

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

func (h *INIController) GetBiomarkers(ctx *gin.Context) {
	// Серверная фильтрация по имени
	FindBiomarker := ctx.Query("query")
	var (
		biomarkers []repository.Biomarker
		err        error
	)
	if FindBiomarker == "" {
		biomarkers, err = h.INIModel.GetBiomarkers()
	} else {
		biomarkers, err = h.INIModel.GetBiomarkersByName(FindBiomarker)
	}
	if err != nil {
		logrus.Error(err)
	}

	// Получаем количество позиций в корзине
	BiomarkersCount, err := h.INIModel.GetINIresearchItemsCount(1)
	if err != nil {
		logrus.Error(err)
		BiomarkersCount = 0 // Устанавливаем 0 в случае ошибки
	}

	ctx.HTML(http.StatusOK, "Biomarkers.html", gin.H{
		"time":            time.Now().Format("15:04:05"),
		"biomarkers":      biomarkers,
		"FindBiomarker":   FindBiomarker,
		"BiomarkersCount": BiomarkersCount,
	})
}

// Новые обработчики под явные маршруты биомаркеров
func (h *INIController) GetDetailedBiomarker(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid biomarker ID"})
		return
	}

	biomarker, err := h.INIModel.GetDetailedBiomarker(id)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Biomarker not found"})
		return
	}

	ctx.HTML(http.StatusOK, "DetailedBiomarker.html", gin.H{
		"biomarker": biomarker,
	})
}

func (h *INIController) GetINIresearch(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid INI research ID:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid INI research ID"})
		return
	}

	INIresearch, err := h.INIModel.GetINIresearch(id)
	if err != nil {
		logrus.Error("Error getting INI research:", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "INI research not found"})
		return
	}

	logrus.Info("INI research data:", INIresearch)

	ctx.HTML(http.StatusOK, "INIresearch.html", gin.H{
		"INIresearch": INIresearch,
	})
}
