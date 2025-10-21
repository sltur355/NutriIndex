package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResearchBiomarker (м-м связь) API

// UpdateResearchBiomarkerAPI - PUT /api/research_biomarkers/:research_id/biomarkers/:biomarker_id
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
		"status":  "success",
		"message": "Research biomarker updated successfully",
	})
}

// DeleteResearchBiomarkerAPI - DELETE /api/research_biomarkers/:research_id/biomarkers/:biomarker_id
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
		"status":  "success",
		"message": "Biomarker removed from research successfully",
	})
}
