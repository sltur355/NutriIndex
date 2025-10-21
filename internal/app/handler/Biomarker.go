package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"LAB1/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// GetBiomarkersAPI godoc
// @Summary Получить список биомаркеров
// @Description Возвращает список всех биомаркеров с возможностью фильтрации по названию
// @Tags Биомаркеры
// @Produce json
// @Param name query string false "Фильтр по названию биомаркера"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /biomarkers [get]
func (h *INIController) GetBiomarkersAPI(ctx *gin.Context) {
	name := ctx.Query("name") // Фильтр по названию

	biomarkers, err := h.INIModel.GetBiomarkers(name)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   biomarkers,
		"total":  len(biomarkers),
	})
}

// GetBiomarkerAPI - GET /api/biomarkers/:id
func (h *INIController) GetBiomarkerAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	biomarker, err := h.INIModel.GetBiomarker(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   biomarker,
	})
}

// CreateBiomarkerAPI - POST /api/biomarkers
func (h *INIController) CreateBiomarkerAPI(ctx *gin.Context) {
	var biomarker ds.Biomarker
	if err := ctx.ShouldBindJSON(&biomarker); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	createdBiomarker, err := h.INIModel.CreateBiomarker(biomarker)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   createdBiomarker,
	})
}

func (h *INIController) UpdateBiomarkerAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var biomarker ds.Biomarker
	if err := ctx.ShouldBindJSON(&biomarker); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.INIModel.UpdateBiomarker(uint(id), biomarker)
	if err != nil {
		if err.Error() == "biomarker not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Biomarker updated successfully",
	})
}

// DeleteBiomarkerAPI - DELETE /api/biomarkers/:id
func (h *INIController) DeleteBiomarkerAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.INIModel.DeleteBiomarker(uint(id))
	if err != nil {
		if err.Error() == "biomarker not found" {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Biomarker deleted successfully",
	})
}

// UploadBiomarkerImageAPI - POST /api/biomarkers/:id/image
func (h *INIController) UploadBiomarkerImageAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем существование биомаркера
	_, err = h.INIModel.GetBiomarker(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Получаем файл из формы
	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("no image file provided"))
		return
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	defer src.Close()

	// Генерируем имя файла на латинице
	fileName := h.INIModel.GenerateImageFileName(file.Filename)

	// Загружаем в MinIO
	err = h.INIModel.UploadFileToMinIO(
		context.Background(),
		fileName,
		src,
		file.Size,
		file.Header.Get("Content-Type"),
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Формируем URL изображения КАК В ПРИМЕРЕ С ХРОНИКОЙ
	imageURL := fmt.Sprintf("http://127.0.0.1:9000/%s/%s", "biomarkers", fileName)

	// Обновляем биомаркер в БД
	err = h.INIModel.UpdateBiomarkerImage(uint(id), imageURL)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Image uploaded successfully",
		"image_url": imageURL,
	})
}

// AddBiomarkerToResearchAPI - POST /api/biomarkers/:id/add_to_research
func (h *INIController) AddBiomarkerToINIResearchAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	biomarkerID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid biomarker ID format"))
		return
	}

	// Проверяем существование биомаркера
	_, err = h.INIModel.GetBiomarker(uint(biomarkerID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	// Добавляем биомаркер в черновик исследования
	researchID, err := h.INIModel.AddBiomarkerToDraftINIResearch(uint(biomarkerID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Biomarker added to draft research successfully",
		"research_id": researchID,
	})
}
