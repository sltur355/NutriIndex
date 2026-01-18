package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResearchBiomarker (м-м связь) API

// UpdateResearchBiomarkerAPI godoc
// @Summary Обновить значение биомаркера в исследовании
// @Description Обновляет значение пациента для биомаркера в исследовании (только для врачей)
// @Tags Исследования
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param research_id path int true "ID исследования"
// @Param biomarker_id path int true "ID биомаркера"
// @Param updates body object true "Обновляемые поля" SchemaExample({"patient_value": 42.5})
// @Success 200 {object} map[string]interface{} "Значение обновлено"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование или биомаркер не найдены"
// @Router /research_biomarkers/{research_id}/biomarkers/{biomarker_id} [put]
func (h *INIController) UpdateResearchBiomarkerAPI(ctx *gin.Context) {
	researchIDStr := ctx.Param("research_id")
	biomarkerIDStr := ctx.Param("biomarker_id")

	researchID, err := strconv.ParseUint(researchIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid research ID format"))
		return
	}
	biomarkerID, err := strconv.ParseUint(biomarkerIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid biomarker ID format"))
		return
	}

	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	err = h.INIModel.UpdateResearchBiomarker(uint(researchID), uint(biomarkerID), updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no valid fields") {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Research biomarker updated successfully",
	})
}

// DeleteResearchBiomarkerAPI godoc
// @Summary Удалить биомаркер из исследования
// @Description Удаляет биомаркер из исследования (только для врачей)
// @Tags Исследования
// @Security ApiKeyAuth
// @Produce json
// @Param research_id path int true "ID исследования"
// @Param biomarker_id path int true "ID биомаркера"
// @Success 200 {object} map[string]interface{} "Биомаркер удален из исследования"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Failure 404 {object} map[string]interface{} "Исследование или биомаркер не найдены"
// @Router /research_biomarkers/{research_id}/biomarkers/{biomarker_id} [delete]
func (h *INIController) DeleteResearchBiomarkerAPI(ctx *gin.Context) {
	researchIDStr := ctx.Param("research_id")
	biomarkerIDStr := ctx.Param("biomarker_id")

	researchID, err := strconv.ParseUint(researchIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid research ID format"))
		return
	}
	biomarkerID, err := strconv.ParseUint(biomarkerIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid biomarker ID format"))
		return
	}

	err = h.INIModel.DeleteResearchBiomarker(uint(researchID), uint(biomarkerID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Biomarker removed from research successfully",
	})
}
