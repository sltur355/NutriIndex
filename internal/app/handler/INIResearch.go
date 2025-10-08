package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
		// Показываем страницу с ошибкой (как в примере с хроникой)
		ctx.HTML(http.StatusOK, "INIresearch.html", gin.H{
			"error": "Заявка не найдена",
		})
		return
	}

	ctx.HTML(http.StatusOK, "INIresearch.html", gin.H{
		"INIresearch": INIresearch,
	})
}

// AddBiomarkerToResearch - добавление биомаркера в исследование (ORM)
func (h *INIController) AddBiomarkerToResearch(ctx *gin.Context) {
	biomarkerIDStr := ctx.PostForm("biomarker_id")
	biomarkerID, err := strconv.Atoi(biomarkerIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	draftResearch, researchBiomarkers, err := h.INIModel.GetDraftResearchInfo()

	if err != nil {
		// Создаем новое исследование
		_, err := h.INIModel.CreateResearch(uint(biomarkerID))
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		// researchID не нужен - просто создаем и редиректим
	} else {
		researchID := int(draftResearch.ID)

		// Проверяем нет ли уже этого биомаркера
		biomarkerExists := false
		for _, rb := range researchBiomarkers {
			if rb.IDBiomarker == uint(biomarkerID) {
				biomarkerExists = true
				break
			}
		}

		if !biomarkerExists {
			// Получаем биомаркер через репозиторий
			biomarker, err := h.INIModel.GetDetailedBiomarker(biomarkerID)
			defaultValue := 0.0
			if err == nil {
				defaultValue = biomarker.MinValue // минимальное нормальное значение
			}

			err = h.INIModel.AddBiomarkerToResearch(researchID, biomarkerID, defaultValue)
			if err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
		}
	}

	// Редирект на главную страницу (как в примере с хроникой)
	ctx.Redirect(http.StatusFound, "/biomarkers")
}

// DeleteResearch - логическое удаление исследования (SQL)
func (h *INIController) DeleteResearch(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.INIModel.DeleteResearch(id) // ← ИСПРАВЛЕНО
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Перенаправляем на список биомаркеров после удаления
	ctx.Redirect(http.StatusFound, "/biomarkers")
}
