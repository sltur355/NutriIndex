package repository

import (
	"LAB1/internal/app/ds"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (r *INIModel) GetINIresearch(id int) (map[string]interface{}, error) {
	var research ds.INIResearch
	// Проверяем что заявка существует и НЕ в статусе "удалён"
	err := r.db.Where("id = ? AND status != ?", id, "удалён").First(&research).Error
	if err != nil {
		return nil, fmt.Errorf("исследование не найдено")
	}

	// Загружаем связанные биомаркеры
	var researchBiomarkers []ds.ResearchBiomarker
	err = r.db.Where("id_research = ?", id).Find(&researchBiomarkers).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки биомаркеров исследования")
	}

	// Формируем данные для шаблона
	var INIresearchItems []map[string]interface{}
	for _, rb := range researchBiomarkers {
		// Загружаем информацию о биомаркере отдельно
		var biomarker ds.Biomarker
		err := r.db.Where("id = ?", rb.IDBiomarker).First(&biomarker).Error
		if err != nil {
			logrus.Warnf("Biomarker not found: ID=%d", rb.IDBiomarker)
			continue
		}

		item := map[string]interface{}{
			"BiomarkerID":    rb.IDBiomarker,
			"BiomarkerImage": biomarker.ImageURL.String,
			"BiomarkerName":  biomarker.Name,
			"MeasureUnit":    biomarker.MeasureUnit,
			"MinValue":       biomarker.MinValue,
			"MaxValue":       biomarker.MaxValue,
			"Significance":   biomarker.Significance,
			"PatientValue":   rb.PatientValue,
		}
		INIresearchItems = append(INIresearchItems, item)
	}

	result := map[string]interface{}{
		"ID":               research.ID,
		"PatientName":      research.PatientName,
		"PatientBirth":     research.PatientBirth,
		"PatientGender":    research.PatientGender,
		"INIResult":        research.INIResult,
		"INIresearchItems": INIresearchItems,
	}

	return result, nil
}

func (r *INIModel) GetINIresearchItemsCount(id int) (int, error) {
	var count int64
	err := r.db.Model(&ds.ResearchBiomarker{}).Where("id_research = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// Получить черновик для пользователя
func (r *INIModel) GetDraftResearch(userID int) (*ds.INIResearch, error) {
	var research ds.INIResearch
	err := r.db.Where("created_by = ? AND status = ?", userID, "черновик").First(&research).Error
	if err != nil {
		return nil, err // Нет черновика - это нормально
	}
	return &research, nil
}

// Добавить биомаркер в исследование
func (r *INIModel) AddBiomarkerToResearch(researchID, biomarkerID int, patientValue float64) error {
	researchBiomarker := ds.ResearchBiomarker{
		IDResearch:   uint(researchID),
		IDBiomarker:  uint(biomarkerID),
		PatientValue: patientValue,
	}
	return r.db.Create(&researchBiomarker).Error
}

// Логическое удаление исследования (SQL UPDATE)
func (r *INIModel) DeleteResearch(id int) error {
	return r.db.Exec("UPDATE ini_researches SET status = 'удалён' WHERE id = ?", id).Error
}

// GetDraftResearchInfo получает черновик исследования и связанные биомаркеры
func (r *INIModel) GetDraftResearchInfo() (ds.INIResearch, []ds.ResearchBiomarker, error) {
	creatorID := uint(1) // Пока фиксированный пользователь

	var research ds.INIResearch
	// Ищем только НЕ удаленные черновики
	err := r.db.Where("created_by = ? AND status = ?", creatorID, "черновик").First(&research).Error
	if err != nil {
		return ds.INIResearch{}, nil, err
	}

	var researchBiomarkers []ds.ResearchBiomarker
	err = r.db.Where("id_research = ?", research.ID).Find(&researchBiomarkers).Error
	if err != nil {
		return ds.INIResearch{}, nil, err
	}

	return research, researchBiomarkers, nil
}

// CreateResearch создает новое исследование
func (r *INIModel) CreateResearch(biomarkerID uint) (ds.INIResearch, error) {
	research := ds.INIResearch{
		Status:        "черновик",
		CreatedBy:     1, // userID
		PatientName:   "Новый пациент",
		PatientBirth:  "01.01.2000",
		PatientGender: "Мужской",
	}

	err := r.db.Create(&research).Error
	if err != nil {
		return ds.INIResearch{}, err
	}

	// Добавляем первый биомаркер
	err = r.AddBiomarkerToResearch(int(research.ID), int(biomarkerID), 0)
	if err != nil {
		return ds.INIResearch{}, err
	}

	return research, nil
}

// CalculateINIResult рассчитывает итоговый INI результат
func (r *INIModel) CalculateINIResult(researchID int) (float64, error) {
	// Получаем все биомаркеры исследования
	var researchBiomarkers []ds.ResearchBiomarker
	err := r.db.Preload("Biomarker").Where("id_research = ?", researchID).Find(&researchBiomarkers).Error
	if err != nil {
		return 0, err
	}

	var totalScore float64

	for _, rb := range researchBiomarkers {
		// Нормализуем значение пациента относительно диапазона нормы
		valueRange := rb.Biomarker.MaxValue - rb.Biomarker.MinValue
		if valueRange == 0 {
			continue // избегаем деления на ноль
		}

		normalizedValue := (rb.PatientValue - rb.Biomarker.MinValue) / valueRange

		// Ограничиваем значение между 0 и 1
		if normalizedValue < 0 {
			normalizedValue = 0
		}
		if normalizedValue > 1 {
			normalizedValue = 1
		}

		// Умножаем на значимость биомаркера и добавляем к общему score
		totalScore += normalizedValue * rb.Biomarker.Significance
	}

	// Преобразуем в процентное значение (0-100)
	iniResult := totalScore * 100

	// Сохраняем результат в заявку
	err = r.db.Model(&ds.INIResearch{}).Where("id = ?", researchID).Update("ini_result", iniResult).Error
	if err != nil {
		return 0, err
	}

	return iniResult, nil
}
