package handler

import (
	"net/http"
	"strconv"
	"time"

	"LAB1/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *INIController) GetBiomarkers(ctx *gin.Context) {
	var biomarkers []ds.Biomarker
	var err error

	FindBiomarker := ctx.Query("query")
	if FindBiomarker == "" {
		biomarkers, err = h.INIModel.GetBiomarkers()
	} else {
		biomarkers, err = h.INIModel.GetBiomarkersByName(FindBiomarker)
	}
	if err != nil {
		logrus.Error(err)
	}

	// Получаем информацию о черновике (как в примере с хроникой)
	draftResearch, researchBiomarkers, err := h.INIModel.GetDraftResearchInfo()
	var draftResearchID uint = 0
	var biomarkersCount int = 0

	if err == nil {
		draftResearchID = draftResearch.ID
		biomarkersCount = len(researchBiomarkers)
	}

	ctx.HTML(http.StatusOK, "Biomarkers.html", gin.H{
		"time":              time.Now().Format("15:04:05"),
		"biomarkers":        biomarkers,
		"FindBiomarker":     FindBiomarker,
		"BiomarkersCount":   biomarkersCount,
		"CurrentResearchID": draftResearchID, // ← как в chronicleResources.html
	})
}

func (h *INIController) GetDetailedBiomarker(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	biomarker, err := h.INIModel.GetDetailedBiomarker(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "DetailedBiomarker.html", gin.H{
		"biomarker": biomarker,
	})
}
