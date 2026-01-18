package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"LAB1/internal/app/ds"
	"LAB1/internal/app/middleware"

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
	name := ctx.Query("name")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "12"))
	useIndex, _ := strconv.ParseBool(ctx.DefaultQuery("use_index", "true"))

	const MAX_LIMIT = 100
	if limit <= 0 {
		limit = 12 // значение по умолчанию
	}
	if limit > MAX_LIMIT {
		limit = MAX_LIMIT // максимальное ограничение
	}
	if page <= 0 {
		page = 1
	}

	// Добавляем обработку ошибки
	biomarkers, total, execTime, err := h.INIModel.GetBiomarkers(name, page, limit, useIndex)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Рассчитываем количество страниц
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	ctx.JSON(http.StatusOK, gin.H{
		"data": biomarkers,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": totalPages,
		},
		"execution_time_ms": execTime.Milliseconds(),
		"use_index":         useIndex,
	})
}

// GetBiomarkerAPI godoc
// @Summary Получить биомаркер по ID
// @Description Возвращает детальную информацию о биомаркере
// @Tags Биомаркеры
// @Produce json
// @Param id path int true "ID биомаркера"
// @Success 200 {object} ds.Biomarker "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 404 {object} map[string]interface{} "Биомаркер не найден"
// @Router /biomarkers/{id} [get]
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

	ctx.JSON(http.StatusOK, biomarker)
}

// UploadBiomarkerImageAPI godoc
// @Summary Загрузить изображение биомаркера
// @Description Загружает изображение для биомаркера в MinIO (только для администраторов)
// @Tags Биомаркеры
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID биомаркера"
// @Param image formData file true "Изображение биомаркера"
// @Success 200 {object} map[string]interface{} "Изображение загружено"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Биомаркер не найден"
// @Router /biomarkers/{id}/image [post]
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

	imageURL := fmt.Sprintf("http://localhost:9000/%s/%s", "biomarkers", fileName)

	// Обновляем биомаркер в БД
	err = h.INIModel.UpdateBiomarkerImage(uint(id), imageURL)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Image uploaded successfully",
		"image_url": imageURL,
	})
}

// CreateBiomarkerAPI godoc
// @Summary Создать новый биомаркер
// @Description Создает новый биомаркер в системе (только для администраторов)
// @Tags Биомаркеры
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param biomarker body ds.Biomarker true "Данные биомаркера"
// @Success 201 {object} ds.Biomarker "Биомаркер создан"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Router /biomarkers [post]
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

	ctx.JSON(http.StatusCreated, createdBiomarker)
}

// UpdateBiomarkerAPI godoc
// @Summary Обновить биомаркер
// @Description Обновляет информацию о биомаркере (только для администраторов)
// @Tags Биомаркеры
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID биомаркера"
// @Param biomarker body ds.Biomarker true "Обновленные данные биомаркера"
// @Success 200 {object} map[string]interface{} "Биомаркер обновлен"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Биомаркер не найден"
// @Router /biomarkers/{id} [put]
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
		"message": "Biomarker updated successfully",
	})
}

// DeleteBiomarkerAPI godoc
// @Summary Удалить биомаркер
// @Description Удаляет биомаркер и его изображение из MinIO (только для администраторов)
// @Tags Биомаркеры
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "ID биомаркера"
// @Success 200 {object} map[string]interface{} "Биомаркер удален"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Биомаркер не найден"
// @Router /biomarkers/{id} [delete]
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
		"message": "Biomarker deleted successfully",
	})
}

// AddBiomarkerToINIResearchAPI godoc
// @Summary Добавить биомаркер в исследование
// @Description Добавляет биомаркер в черновик исследования (только для врачей)
// @Tags Исследования
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "ID биомаркера"
// @Success 200 {object} map[string]interface{} "Биомаркер добавлен в исследование"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Биомаркер не найден"
// @Router /biomarkers/{id}/add_to_research [post]
func (h *INIController) AddBiomarkerToINIResearchAPI(ctx *gin.Context) {
	// Получаем creatorID из контекста
	creatorID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

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

	// Добавляем биомаркер в черновик исследования с реальным creatorID
	researchID, err := h.INIModel.AddBiomarkerToDraftINIResearch(uint(biomarkerID), creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Biomarker added to draft research successfully",
		"research_id": researchID,
	})
}

// GetBiomarkersCompareAPI - сравнение производительности с индексом и без
// GetBiomarkersCompareAPI - сравнение производительности с индексом и без
func (h *INIController) GetBiomarkersCompareAPI(ctx *gin.Context) {
	name := ctx.Query("name")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// Выполняем оба запроса
	biomarkersWithIndex, totalWith, timeWithIndex, err1 := h.INIModel.GetBiomarkers(name, page, limit, true)
	biomarkersWithoutIndex, totalWithout, timeWithoutIndex, err2 := h.INIModel.GetBiomarkers(name, page, limit, false)

	if err1 != nil || err2 != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка сравнения: %v, %v", err1, err2))
		return
	}

	// Рассчитываем эффективность индекса
	var indexEfficiency float64
	if timeWithIndex.Milliseconds() > 0 {
		indexEfficiency = (float64(timeWithoutIndex.Milliseconds()) - float64(timeWithIndex.Milliseconds())) /
			float64(timeWithIndex.Milliseconds()) * 100
	}

	totalPagesWith := int(math.Ceil(float64(totalWith) / float64(limit)))
	totalPagesWithout := int(math.Ceil(float64(totalWithout) / float64(limit)))

	ctx.JSON(http.StatusOK, gin.H{
		"with_index": gin.H{
			"data": biomarkersWithIndex,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": totalWith,
				"pages": totalPagesWith,
			},
			"execution_time_ms": timeWithIndex.Milliseconds(),
		},
		"without_index": gin.H{
			"data": biomarkersWithoutIndex,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": totalWithout,
				"pages": totalPagesWithout,
			},
			"execution_time_ms": timeWithoutIndex.Milliseconds(),
		},
		"comparison": gin.H{
			"speedup_ratio":      float64(timeWithoutIndex.Milliseconds()) / float64(timeWithIndex.Milliseconds()),
			"time_difference_ms": timeWithoutIndex.Milliseconds() - timeWithIndex.Milliseconds(),
			"index_efficiency":   indexEfficiency,
		},
	})
}
