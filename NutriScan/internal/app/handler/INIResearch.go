package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"LAB1/internal/app/ds"
	"LAB1/internal/app/middleware"
	"LAB1/internal/app/role"

	"github.com/gin-gonic/gin"
)

// GetINIResearchCartAPI godoc
// @Summary Получить информацию о корзине исследований
// @Description Возвращает ID черновика и количество биомаркеров в корзине. Для неавторизованных пользователей возвращает request_id=-1, count=0
// @Tags Исследования
// @Produce json
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Router /researches/cart [get]
func (h *INIController) GetINIResearchCartAPI(ctx *gin.Context) {
	creatorID, exists := middleware.GetUserID(ctx)

	// Если пользователь НЕ авторизован - возвращаем -1, 0
	if !exists {
		ctx.JSON(http.StatusOK, gin.H{
			"request_id": -1,
			"count":      0,
		})
		return
	}

	// Если авторизован - возвращаем реальные данные
	requestID, count, err := h.INIModel.GetDraftRequestInfo(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"request_id": requestID,
		"count":      count,
	})
}

// GetINIResearchAPI godoc
// @Summary Получить исследование по ID
// @Description Возвращает детальную информацию об исследовании с биомаркерами
// @Tags Исследования
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "ID исследования"
// @Success 200 {object} ds.ResearchWithBiomarkersResponse "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /researches/{id} [get]
func (h *INIController) GetINIResearchAPI(ctx *gin.Context) {
	// Получаем userID и userRole из контекста
	userID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	userRoleStr, exists := middleware.GetUserRole(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user role not found"))
		return
	}
	userRole := role.FromString(userRoleStr)

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	research, biomarkers, err := h.INIModel.GetINIResearchWithBiomarkers(uint(id), userID, userRole)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	// Получаем статистику по биомаркерам исследования
	calculatedCount, totalCount, err := h.INIModel.GetResearchBiomarkerStats(research.ID)
	if err != nil {
		// В случае ошибки просто ставим 0
		calculatedCount = 0
		totalCount = 0
	}

	// Создаем DTO ответ с правильной структурой
	response := ds.ResearchWithBiomarkersResponse{
		ID:              research.ID,
		Status:          research.Status,
		CreatedAt:       research.CreatedAt,
		PatientName:     research.PatientName,
		PatientBirth:    research.PatientBirth,
		PatientGender:   research.PatientGender,
		INIResult:       research.INIResult,
		FormedAt:        research.FormedAt,
		CompletedAt:     research.CompletedAt,
		CalculatedCount: calculatedCount,
		TotalBiomarkers: totalCount,
	}

	// Добавляем логины создателя и модератора
	if research.User != nil {
		response.CreatorLogin = research.User.Login
	}
	if research.Moderator != nil {
		response.ModeratorLogin = research.Moderator.Login
	}

	// Преобразуем биомаркеры в DTO формат
	response.Biomarkers = make([]ds.BiomarkerItemResponse, 0, len(biomarkers))
	for _, rb := range biomarkers {
		if rb.Biomarker != nil {
			biomarkerItem := ds.BiomarkerItemResponse{
				ID:           rb.Biomarker.ID,
				Name:         rb.Biomarker.Name,
				Description:  rb.Biomarker.Description,
				MeasureUnit:  rb.Biomarker.MeasureUnit,
				MinValue:     rb.Biomarker.MinValue,
				MaxValue:     rb.Biomarker.MaxValue,
				ImageURL:     rb.Biomarker.ImageURL,
				PatientValue: rb.PatientValue,
				Significance: rb.Biomarker.Significance,
			}
			response.Biomarkers = append(response.Biomarkers, biomarkerItem)
		}
	}

	ctx.JSON(http.StatusOK, response)
}

// GetINIResearchesAPI godoc
// @Summary Получить список исследований
// @Description Возвращает список исследований с возможностью фильтрации по статусу и дате
// @Tags Исследования
// @Security ApiKeyAuth
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param start_date query string false "Начальная дата (формат: 2006-01-02)"
// @Param end_date query string false "Конечная дата (формат: 2006-01-02)"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Router /researches [get]
func (h *INIController) GetINIResearchesAPI(ctx *gin.Context) {
	// Получаем userID и userRole из контекста
	userID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	userRoleStr, exists := middleware.GetUserRole(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user role not found"))
		return
	}
	userRole := role.FromString(userRoleStr)

	var startDate, endDate *time.Time
	status := ctx.Query("status")

	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	researches, err := h.INIModel.GetINIResearchesWithFilters(userID, userRole, status, startDate, endDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Преобразуем в DTO для списка
	researchList := make([]ds.ResearchListResponse, 0, len(researches))
	for _, research := range researches {
		item := ds.ResearchListResponse{
			ID:          research.ID,
			Status:      research.Status,
			CreatedAt:   research.CreatedAt,
			PatientName: research.PatientName,
			FormedAt:    research.FormedAt,
			CompletedAt: research.CompletedAt,
			INIResult:   research.INIResult,
		}

		// Добавляем логины
		if research.User != nil {
			item.CreatorLogin = research.User.Login
		}
		if research.Moderator != nil {
			item.ModeratorLogin = research.Moderator.Login
		}

		researchList = append(researchList, item)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  researchList,
		"total": len(researchList),
	})
}

// UpdateINIResearchAPI godoc
// @Summary Обновить информацию об исследовании
// @Description Обновляет информацию о пациенте в исследовании (только для врачей)
// @Tags Исследования
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID исследования"
// @Param patient_info body object true "Информация о пациенте" SchemaExample({"patient_name": "Иван Иванов", "patient_birth": "01.01.1990", "patient_gender": "Мужской"})
// @Success 200 {object} map[string]interface{} "Исследование обновлено"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /researches/{id} [put]
func (h *INIController) UpdateINIResearchAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	var patientInfo struct {
		PatientName   *string `json:"patient_name,omitempty"`
		PatientBirth  *string `json:"patient_birth,omitempty"`
		PatientGender *string `json:"patient_gender,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&patientInfo); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	err = h.INIModel.UpdateINIResearchPatientInfo(uint(id), patientInfo)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Research updated successfully",
	})
}

// FormINIResearchAPI godoc
// @Summary Сформировать исследование
// @Description Переводит исследование из статуса черновика в сформированный (только создатель)
// @Tags Исследования
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "ID исследования"
// @Success 200 {object} map[string]interface{} "Исследование сформировано"
// @Failure 400 {object} map[string]interface{} "Неверные данные или пустое исследование"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /researches/{id}/form [put]
func (h *INIController) FormINIResearchAPI(ctx *gin.Context) {
	// Получаем userID из контекста
	userID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	err = h.INIModel.FormINIResearch(uint(id), userID)
	if err != nil {
		if strings.Contains(err.Error(), "доступен только черновик") ||
			strings.Contains(err.Error(), "необходимо заполнить") ||
			strings.Contains(err.Error(), "исследование пусто") {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Research formed successfully",
	})
}

// CompleteOrRejectINIResearchAPI godoc
// @Summary Завершить или отклонить исследование
// @Description Завершает или отклоняет сформированное исследование (только для модераторов)
// @Tags Исследования
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID исследования"
// @Param action body object true "Действие" SchemaExample({"action": "complete"})
// @Success 200 {object} map[string]interface{} "Исследование завершено/отклонено"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /researches/{id}/complete [put]
func (h *INIController) CompleteOrRejectINIResearchAPI(ctx *gin.Context) {
	// Получаем moderatorID из контекста
	moderatorID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	var requestBody struct {
		Action string `json:"action" binding:"required"` // "complete" или "reject"
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	err = h.INIModel.CompleteOrRejectINIResearch(uint(id), requestBody.Action, moderatorID)
	if err != nil {
		if strings.Contains(err.Error(), "только сформированные исследования") ||
			strings.Contains(err.Error(), "неверное действие") ||
			strings.Contains(err.Error(), "ошибка при расчете") {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	message := "Research completed successfully"
	if requestBody.Action == "reject" {
		message = "Research rejected successfully"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// DeleteINIResearchAPI godoc
// @Summary Удалить исследование
// @Description Удаляет исследование (только создатель для черновиков)
// @Tags Исследования
// @Security ApiKeyAuth
// @Produce json
// @Param id path int true "ID исследования"
// @Success 200 {object} map[string]interface{} "Исследование удалено"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /researches/{id} [delete]
func (h *INIController) DeleteINIResearchAPI(ctx *gin.Context) {
	// Получаем userID из контекста
	userID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	err = h.INIModel.DeleteINIResearch(uint(id), userID)
	if err != nil {
		if strings.Contains(err.Error(), "не найдено") || strings.Contains(err.Error(), "не может быть удалено") {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Research deleted successfully",
	})
}

// CompleteINIResearchAsyncAPI godoc
// @Summary Завершить исследование асинхронно
// @Description Запускает асинхронный расчет INI через внешний сервис
// @Tags Исследования
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID исследования"
// @Success 200 {object} map[string]interface{} "Асинхронный расчет запущен"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /researches/{id}/complete-async [put]
func (h *INIController) CompleteINIResearchAsyncAPI(ctx *gin.Context) {
	// Получаем moderatorID из контекста
	moderatorID, exists := middleware.GetUserID(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("user not authenticated"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	// Запускаем асинхронный расчет
	err = h.INIModel.CompleteINIResearchAsync(uint(id), moderatorID)
	if err != nil {
		if strings.Contains(err.Error(), "только сформированные исследования") {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Async INI calculation started",
		"research_id": id,
		"status":      "processing",
	})
}

// UpdateINIResultAPI godoc
// @Summary Обновить результат INI расчета
// @Description Принимает результат расчета от асинхронного сервиса
// @Tags Исследования
// @Accept json
// @Produce json
// @Param result body ds.UpdateINIResultRequest true "Результат расчета INI"
// @Success 200 {object} map[string]interface{} "Результат обновлен"
// @Failure 400 {object} map[string]interface{} "Неверные данные или неверный ключ"
// @Failure 404 {object} map[string]interface{} "Исследование не найдено"
// @Router /api/async/update-ini-result [post]
func (h *INIController) UpdateINIResultAPI(ctx *gin.Context) {
	var req struct {
		ResearchID uint    `json:"research_id"`
		INIResult  float64 `json:"ini_result"`
		SecretKey  string  `json:"secret_key"`
	}
	if ctx.BindJSON(&req) != nil {
		ctx.JSON(400, gin.H{"error": "bad request"})
		return
	}
	// Псевдо-авторизация
	if req.SecretKey != "nutriscan_async_key_2024" {
		ctx.JSON(401, gin.H{"error": "invalid key"})
		return
	}

	err := h.INIModel.UpdateINIResult(req.ResearchID, req.INIResult)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(200, gin.H{
		"message":     "INI result updated successfully",
		"research_id": req.ResearchID,
		"ini_result":  req.INIResult,
	})
}
