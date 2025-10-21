package repository

import (
	"LAB1/internal/app/ds"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GetINIResearchesWithFilters - GET список исследований с фильтрацией
func (r *INIModel) GetINIResearchesWithFilters(status string, dateFrom, dateTo *time.Time) ([]ds.INIResearch, error) {
	var researches []ds.INIResearch
	// Исключаем удаленные и черновики (как в задании)
	query := r.db.Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if dateFrom != nil {
		query = query.Where("formed_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("formed_at <= ?", *dateTo)
	}

	err := query.Preload("User").Preload("Moderator").Order("formed_at DESC").Find(&researches).Error
	return researches, err
}

// GetResearchByID - GET одно исследование по ID
func (r *INIModel) GetINIResearchByID(id uint) (ds.INIResearch, error) {
	var research ds.INIResearch
	err := r.db.Preload("User").Preload("Moderator").
		Where("id = ? AND status != ?", id, ds.StatusDeleted).
		First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.INIResearch{}, errors.New("research not found")
		}
		return ds.INIResearch{}, err
	}
	return research, nil
}

// GetResearchWithBiomarkers - GET исследование с биомаркерами
func (r *INIModel) GetINIResearchWithBiomarkers(id uint) (ds.INIResearch, []ds.ResearchBiomarker, error) {
	research, err := r.GetINIResearchByID(id)
	if err != nil {
		return ds.INIResearch{}, nil, err
	}

	var biomarkers []ds.ResearchBiomarker
	err = r.db.Preload("Biomarker").Where("id_research = ?", id).Find(&biomarkers).Error
	if err != nil {
		return ds.INIResearch{}, nil, err
	}

	return research, biomarkers, nil
}

// UpdateResearchPatientInfo - PUT обновление информации о пациенте
func (r *INIModel) UpdateINIResearchPatientInfo(id uint, patientInfo struct {
	PatientName   string `json:"patient_name"`
	PatientBirth  string `json:"patient_birth"`
	PatientGender string `json:"patient_gender"`
}) error {
	var existingResearch ds.INIResearch
	err := r.db.Where("id = ? AND status != ?", id, ds.StatusDeleted).First(&existingResearch).Error
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"patient_name":   patientInfo.PatientName,
		"patient_birth":  patientInfo.PatientBirth,
		"patient_gender": patientInfo.PatientGender,
	}

	return r.db.Model(&ds.INIResearch{}).Where("id = ?", id).Updates(updates).Error
}

// GetDraftRequestInfo возвращает ID черновика и количество биомаркеров в корзине
func (r *INIModel) GetDraftRequestInfo() (uint, int, error) {
	creatorID := GetFixedCreatorID()

	var research ds.INIResearch
	err := r.db.Where("created_by = ? AND status = ?", creatorID, "черновик").First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, nil // Нет черновика
		}
		return 0, 0, err
	}

	var count int64
	err = r.db.Model(&ds.ResearchBiomarker{}).Where("id_research = ?", research.ID).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}

	return research.ID, int(count), nil
}

// UpdateINIResearchStatus - общий метод смены статуса исследования
func (r *INIModel) UpdateINIResearchStatus(id uint, newStatus ds.INIResearchStatus, moderatorID *uint) error {
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status != ?", id, ds.StatusDeleted).First(&research).Error
	if err != nil {
		return err
	}

	if !r.isValidINIStatusTransition(research.Status, newStatus) {
		return fmt.Errorf("недопустимый переход статуса с %s на %s", research.Status, newStatus)
	}

	updates := map[string]interface{}{
		"status": newStatus,
	}

	switch newStatus {
	case ds.StatusFormed:
		updates["formed_at"] = time.Now()
	case ds.StatusCompleted, ds.StatusRejected:
		updates["completed_at"] = time.Now()
		if moderatorID != nil {
			updates["moderator_id"] = *moderatorID
		}
	}

	return r.db.Model(&ds.INIResearch{}).Where("id = ?", id).Updates(updates).Error
}

// isValidINIStatusTransition - проверка допустимости перехода между статусами
func (r *INIModel) isValidINIStatusTransition(current, new ds.INIResearchStatus) bool {
	validTransitions := map[ds.INIResearchStatus][]ds.INIResearchStatus{
		ds.StatusDraft:     {ds.StatusDeleted, ds.StatusFormed},
		ds.StatusFormed:    {ds.StatusCompleted, ds.StatusRejected},
		ds.StatusCompleted: {},
		ds.StatusRejected:  {},
		ds.StatusDeleted:   {},
	}

	allowedStatuses, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == new {
			return true
		}
	}
	return false
}

// FormINIResearch - формирование исследования (только создатель)
func (r *INIModel) FormINIResearch(id uint, creatorID uint) error {
	// Проверяем что это черновик текущего пользователя
	var research ds.INIResearch
	err := r.db.Where("id = ? AND created_by = ? AND status = ?", id, creatorID, ds.StatusDraft).First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("доступен только черновик текущего пользователя")
		}
		return err
	}

	// Улучшенная проверка обязательных полей пациента (trim пробелов)
	if strings.TrimSpace(research.PatientName) == "" ||
		strings.TrimSpace(research.PatientBirth) == "" ||
		strings.TrimSpace(research.PatientGender) == "" {
		return errors.New("необходимо заполнить все поля пациента: ФИО, дата рождения, пол")
	}

	// Проверяем что есть хотя бы один биомаркер
	var biomarkersCount int64
	err = r.db.Model(&ds.ResearchBiomarker{}).Where("id_research = ?", id).Count(&biomarkersCount).Error
	if err != nil {
		return err
	}
	if biomarkersCount == 0 {
		return errors.New("исследование пусто - добавьте хотя бы один биомаркер")
	}

	return r.UpdateINIResearchStatus(id, ds.StatusFormed, nil)
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

	return iniResult, nil
}

// CompleteOrRejectINIResearch - завершение/отклонение исследования (только модератор)
func (r *INIModel) CompleteOrRejectINIResearch(id uint, action string, moderatorID uint) error {
	// Проверяем что исследование в статусе "сформирован"
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status = ?", id, ds.StatusFormed).First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("только сформированные исследования могут быть завершены или отклонены")
		}
		return err
	}

	switch action {
	case "complete":
		// Выполняем расчет INI индекса при завершении
		iniResult, err := r.CalculateINIResult(int(id)) // ← Исправленный вызов
		if err != nil {
			return fmt.Errorf("ошибка при расчете INI индекса: %v", err)
		}

		// Сохраняем результат INI и завершаем исследование
		err = r.db.Model(&ds.INIResearch{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":       ds.StatusCompleted,
			"ini_result":   iniResult,
			"completed_at": time.Now(),
			"moderator_id": moderatorID,
		}).Error
		return err

	case "reject":
		return r.UpdateINIResearchStatus(id, ds.StatusRejected, &moderatorID)

	default:
		return errors.New("неверное действие. Используйте 'complete' или 'reject'")
	}
}

// DeleteINIResearch - удаление исследования (только создатель для черновика)
func (r *INIModel) DeleteINIResearch(id uint, creatorID uint) error {
	tx := r.db.Model(&ds.INIResearch{}).
		Where("id = ? AND created_by = ? AND status = ?", id, creatorID, ds.StatusDraft).
		Update("status", ds.StatusDeleted)

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("исследование не найдено или не может быть удалено")
	}
	return nil
}

// AddBiomarkerToDraftResearch добавляет биомаркер в черновик исследования
func (r *INIModel) AddBiomarkerToDraftINIResearch(biomarkerID uint) (uint, error) {
	creatorID := GetFixedCreatorID()

	// Ищем существующий черновик
	var research ds.INIResearch
	err := r.db.Where("created_by = ? AND status = ?", creatorID, "черновик").First(&research).Error

	if err != nil {
		// Создаем новый черновик
		research = ds.INIResearch{
			Status:        "черновик",
			CreatedBy:     creatorID,
			PatientName:   "Новый пациент",
			PatientBirth:  "01.01.2000",
			PatientGender: "Мужской",
			CreatedAt:     time.Now(),
		}

		err = r.db.Create(&research).Error
		if err != nil {
			return 0, err
		}
	}

	// Добавляем биомаркер в исследование
	researchBiomarker := ds.ResearchBiomarker{
		IDResearch:   research.ID,
		IDBiomarker:  biomarkerID,
		PatientValue: 0, // Значение по умолчанию
	}

	err = r.db.Create(&researchBiomarker).Error
	if err != nil {
		return 0, err
	}

	return research.ID, nil
}
