package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"LAB1/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// GetResearchCartAPI - GET /api/researches/cart (иконка корзины)
func (h *INIController) GetINIResearchCartAPI(ctx *gin.Context) {
	requestID, count, err := h.INIModel.GetDraftRequestInfo()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"request_id": requestID,
		"count":      count,
	})
}

// GetResearchesAPI - GET /api/researches (список с фильтрацией)
func (h *INIController) GetINIResearchesAPI(ctx *gin.Context) {
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

	researches, err := h.INIModel.GetINIResearchesWithFilters(status, startDate, endDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   researches,
	})
}

// GetResearchAPI - GET /api/researches/:id
func (h *INIController) GetINIResearchAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	research, biomarkers, err := h.INIModel.GetINIResearchWithBiomarkers(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"research":   research,
		"biomarkers": biomarkers,
	})
}

// UpdateResearchAPI - PUT /api/researches/:id
func (h *INIController) UpdateINIResearchAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	var patientInfo struct {
		PatientName   string `json:"patient_name"`
		PatientBirth  string `json:"patient_birth"`
		PatientGender string `json:"patient_gender"`
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
		"status":  "success",
		"message": "Research updated successfully",
	})
}

// FormINIResearchAPI - PUT /api/researches/:id/form
func (h *INIController) FormINIResearchAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	err = h.INIModel.FormINIResearch(uint(id), repository.GetFixedCreatorID())
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
		"status":  "success",
		"message": "Research formed successfully",
	})
}

// CompleteOrRejectINIResearchAPI - PUT /api/researches/:id/complete
func (h *INIController) CompleteOrRejectINIResearchAPI(ctx *gin.Context) {
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

	err = h.INIModel.CompleteOrRejectINIResearch(uint(id), requestBody.Action, repository.GetFixedModeratorID())
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
		"status":  "success",
		"message": message,
	})
}

// DeleteINIResearchAPI - DELETE /api/researches/:id
func (h *INIController) DeleteINIResearchAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid ID format"))
		return
	}

	err = h.INIModel.DeleteINIResearch(uint(id), repository.GetFixedCreatorID())
	if err != nil {
		if strings.Contains(err.Error(), "не найдено") || strings.Contains(err.Error(), "не может быть удалено") {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Research deleted successfully",
	})
}
