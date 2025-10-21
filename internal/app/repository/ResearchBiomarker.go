package repository

import (
	"LAB1/internal/app/ds"
	"errors"

	"gorm.io/gorm"
)

// ResearchBiomarker (м-м связь) методы

// DeleteResearchBiomarker удаляет биомаркер из исследования (без PK м-м)
func (r *INIModel) DeleteResearchBiomarker(researchID uint, biomarkerID uint) error {
	// Проверяем что исследование существует и является черновиком
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status = ?", researchID, "черновик").First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("draft research not found")
		}
		return err
	}

	// Удаляем связь между исследованием и биомаркером
	tx := r.db.Where("id_research = ? AND id_biomarker = ?", researchID, biomarkerID).Delete(&ds.ResearchBiomarker{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("biomarker not found in research")
	}
	return nil
}

// UpdateResearchBiomarker обновляет м-м запись (значение пациента)
func (r *INIModel) UpdateResearchBiomarker(researchID uint, biomarkerID uint, updates map[string]interface{}) error {
	// Проверяем что исследование существует и является черновиком
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status = ?", researchID, "черновик").First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("draft research not found")
		}
		return err
	}

	// Фильтруем системные поля из обновления
	allowedFields := map[string]bool{
		"patient_value": true, // Единственное поле которое можно обновлять
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("no valid fields to update")
	}

	// Обновляем связь между исследованием и биомаркером
	tx := r.db.Model(&ds.ResearchBiomarker{}).
		Where("id_research = ? AND id_biomarker = ?", researchID, biomarkerID).
		Updates(filteredUpdates)

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("biomarker not found in research")
	}
	return nil
}
