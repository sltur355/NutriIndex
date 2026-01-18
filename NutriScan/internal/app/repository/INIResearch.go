package repository

import (
	"LAB1/internal/app/ds"
	"LAB1/internal/app/role"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

// CallAsyncINIService вызывает асинхронный сервис расчета INI
func (r *INIModel) CallAsyncINIService(researchID uint, biomarkerIDs []uint) error {
	// URL Django-сервиса (из переменных окружения)
	djangoURL := os.Getenv("DJANGO_SERVICE_URL")
	if djangoURL == "" {
		// По умолчанию для локальной разработки
		djangoURL = "http://localhost:8001/api/async/calculate-ini/"
	}

	// Подготовка запроса
	requestData := ds.AsyncINICalculationRequest{
		ResearchID:   researchID,
		BiomarkerIDs: biomarkerIDs,
		SecretKey:    "nutriscan_async_key_2024", // Секретный ключ
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Асинхронный HTTP-вызов
	go func() {
		resp, err := http.Post(djangoURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("ERROR: Failed to call async service: %v\n", err)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("DEBUG: Async service response: %s\n", string(body))
	}()

	return nil
}

// CompleteINIResearchAsync - асинхронное завершение исследования
func (r *INIModel) CompleteINIResearchAsync(researchID uint, moderatorID uint) error {
	// Проверяем что исследование в статусе "сформирован"
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status = ?", researchID, ds.StatusFormed).First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("только сформированные исследования могут быть завершены")
		}
		return err
	}

	// Получаем все биомаркеры исследования
	var researchBiomarkers []ds.ResearchBiomarker
	err = r.db.Where("id_research = ?", researchID).Find(&researchBiomarkers).Error
	if err != nil {
		return err
	}

	// Собираем ID биомаркеров
	biomarkerIDs := make([]uint, 0, len(researchBiomarkers))
	for _, rb := range researchBiomarkers {
		biomarkerIDs = append(biomarkerIDs, rb.IDBiomarker)
	}

	// Вызываем асинхронный сервис расчета INI
	if err := r.CallAsyncINIService(researchID, biomarkerIDs); err != nil {
		return fmt.Errorf("failed to start async calculation: %v", err)
	}

	// Обновляем статус на "в процессе расчета" (можно добавить новый статус)
	fmt.Printf("INFO: Async INI calculation started for research %d\n", researchID)

	return nil
}

// GetINIResearchesWithFilters - GET список исследований с фильтрацией
func (r *INIModel) GetINIResearchesWithFilters(userID uint, userRole role.Role, status string, dateFrom, dateTo *time.Time) ([]ds.INIResearch, error) {
	var researches []ds.INIResearch

	// Базовый запрос - исключаем удаленные
	query := r.db.Where("status != ?", ds.StatusDeleted)

	// НОВАЯ ЛОГИКА ФИЛЬТРАЦИИ:
	// - Patient: видит только СВОИ исследования
	// - Doctor: видит ВСЕ исследования
	if userRole == role.Patient {
		query = query.Where("created_by = ?", userID)
	}
	// Doctor видит все исследования без фильтрации

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

// GetINIResearchByID - получает исследование по ID с проверкой прав доступа
func (r *INIModel) GetINIResearchByID(id uint, userID uint, userRole role.Role) (ds.INIResearch, error) {
	var research ds.INIResearch

	query := r.db.Preload("User").Preload("Moderator").
		Where("id = ? AND status != ?", id, ds.StatusDeleted)

	// Patient может видеть только свои исследования
	if userRole == role.Patient {
		query = query.Where("created_by = ?", userID)
	}

	err := query.First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.INIResearch{}, errors.New("research not found")
		}
		return ds.INIResearch{}, err
	}
	return research, nil
}

// GetResearchWithBiomarkers - GET исследование с биомаркерами
func (r *INIModel) GetINIResearchWithBiomarkers(id uint, userID uint, userRole role.Role) (ds.INIResearch, []ds.ResearchBiomarker, error) {
	research, err := r.GetINIResearchByID(id, userID, userRole)
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
	PatientName   *string `json:"patient_name,omitempty"`
	PatientBirth  *string `json:"patient_birth,omitempty"`
	PatientGender *string `json:"patient_gender,omitempty"`
}) error {
	var existingResearch ds.INIResearch
	err := r.db.Where("id = ? AND status != ?", id, ds.StatusDeleted).First(&existingResearch).Error
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}

	if patientInfo.PatientName != nil {
		updates["patient_name"] = patientInfo.PatientName
	}
	if patientInfo.PatientBirth != nil {
		updates["patient_birth"] = patientInfo.PatientBirth
	}
	if patientInfo.PatientGender != nil {
		updates["patient_gender"] = patientInfo.PatientGender
	}

	if len(updates) == 0 {
		return nil // Нет изменений
	}

	return r.db.Model(&ds.INIResearch{}).Where("id = ?", id).Updates(updates).Error
}

// GetDraftRequestInfo возвращает ID черновика и количество биомаркеров в корзине
func (r *INIModel) GetDraftRequestInfo(creatorID uint) (uint, int, error) {
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

	// ИЗМЕНЕНИЕ: Используем указатели на time.Time вместо sql.NullTime
	now := time.Now()
	switch newStatus {
	case ds.StatusFormed:
		updates["formed_at"] = &now
	case ds.StatusCompleted, ds.StatusRejected:
		updates["completed_at"] = &now
		if moderatorID != nil {
			updates["moderator_id"] = moderatorID
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
func (r *INIModel) FormINIResearch(id uint, userID uint) error {
	// Проверяем что это черновик и принадлежит пользователю
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status = ? AND created_by = ?", id, ds.StatusDraft, userID).First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("доступен только ваш черновик")
		}
		return err
	}

	// Проверка nullable полей пациента
	if research.PatientName == nil || *research.PatientName == "" ||
		research.PatientBirth == nil || *research.PatientBirth == "" ||
		research.PatientGender == nil || *research.PatientGender == "" {
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
func (r *INIModel) CalculateINIResult(researchID uint) (float64, error) {
	// Получаем все биомаркеры исследования
	var researchBiomarkers []ds.ResearchBiomarker
	err := r.db.Preload("Biomarker").Where("id_research = ?", researchID).Find(&researchBiomarkers).Error
	if err != nil {
		return 0, err
	}

	var totalScore float64

	for _, rb := range researchBiomarkers {
		// Проверяем что значение пациента не nil
		if rb.PatientValue == nil {
			continue // пропускаем биомаркеры без значения
		}

		// Нормализуем значение пациента относительно диапазона нормы
		valueRange := rb.Biomarker.MaxValue - rb.Biomarker.MinValue
		if valueRange == 0 {
			continue // избегаем деления на ноль
		}

		normalizedValue := (*rb.PatientValue - rb.Biomarker.MinValue) / valueRange

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
		iniResult, err := r.CalculateINIResult(id)
		if err != nil {
			return fmt.Errorf("ошибка при расчете INI индекса: %v", err)
		}

		modID := &moderatorID
		now := time.Now()
		err = r.db.Model(&ds.INIResearch{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":       ds.StatusCompleted,
			"ini_result":   iniResult,
			"completed_at": &now,
			"moderator_id": modID,
		}).Error
		return err

	case "reject":
		modID := &moderatorID
		return r.UpdateINIResearchStatus(id, ds.StatusRejected, modID)

	default:
		return errors.New("неверное действие. Используйте 'complete' или 'reject'")
	}
}

// DeleteINIResearch - удаление исследования (только создатель для черновика)
func (r *INIModel) DeleteINIResearch(id uint, userID uint) error {
	tx := r.db.Model(&ds.INIResearch{}).
		Where("id = ? AND status = ? AND created_by = ?", id, ds.StatusDraft, userID).
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
func (r *INIModel) AddBiomarkerToDraftINIResearch(biomarkerID uint, creatorID uint) (uint, error) {
	// Ищем существующий черновик
	var research ds.INIResearch
	err := r.db.Where("created_by = ? AND status = ?", creatorID, "черновик").First(&research).Error

	if err != nil {
		// Создаем новый черновик с nullable полями
		patientName := "Новый пациент"
		patientBirth := "01.01.2000"
		patientGender := "Мужской"

		research = ds.INIResearch{
			Status:        "черновик",
			CreatedBy:     creatorID,
			PatientName:   &patientName,
			PatientBirth:  &patientBirth,
			PatientGender: &patientGender,
			CreatedAt:     time.Now(),
			FormedAt:      nil,
			CompletedAt:   nil,
			ModeratorID:   nil,
		}

		err = r.db.Create(&research).Error
		if err != nil {
			return 0, err
		}
	}

	// Добавляем биомаркер в исследование с nil значением
	researchBiomarker := ds.ResearchBiomarker{
		IDResearch:   research.ID,
		IDBiomarker:  biomarkerID,
		PatientValue: nil, // Значение по умолчанию - nil
	}

	err = r.db.Create(&researchBiomarker).Error
	if err != nil {
		return 0, err
	}

	return research.ID, nil
}

// UpdateBiomarkerValue обновляет значение биомаркера в исследовании
func (r *INIModel) UpdateBiomarkerValue(researchID uint, biomarkerID uint, value *float64) error {
	return r.db.Model(&ds.ResearchBiomarker{}).
		Where("id_research = ? AND id_biomarker = ?", researchID, biomarkerID).
		Update("patient_value", value).Error
}

// repository/ini_research.go - добавить новые методы

// GetResearchBiomarkerCalculatedCount возвращает количество биомаркеров с заполненным patient_value
func (r *INIModel) GetResearchBiomarkerCalculatedCount(researchID uint) (int, int, error) {
	var totalCount int64
	var calculatedCount int64

	// Общее количество биомаркеров в исследовании
	err := r.db.Model(&ds.ResearchBiomarker{}).
		Where("id_research = ?", researchID).
		Count(&totalCount).Error
	if err != nil {
		return 0, 0, err
	}

	// Количество биомаркеров с заполненным patient_value
	err = r.db.Model(&ds.ResearchBiomarker{}).
		Where("id_research = ? AND patient_value IS NOT NULL", researchID).
		Count(&calculatedCount).Error
	if err != nil {
		return 0, 0, err
	}

	return int(calculatedCount), int(totalCount), nil
}

// UpdateINIResult обновляет INI результат исследования
// UpdateINIResult обновляет INI результат исследования
func (r *INIModel) UpdateINIResult(researchID uint, iniResult float64) error {
	// Проверяем что исследование существует
	var research ds.INIResearch
	err := r.db.Where("id = ? AND status != ?", researchID, ds.StatusDeleted).First(&research).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("research not found")
		}
		return err
	}

	// Обновляем результат и статус
	updates := map[string]interface{}{
		"ini_result":   iniResult,
		"status":       ds.StatusCompleted,
		"completed_at": time.Now(),
	}

	return r.db.Model(&ds.INIResearch{}).Where("id = ?", researchID).Updates(updates).Error
}

// GetResearchBiomarkerStats возвращает статистику по биомаркерам исследования
func (r *INIModel) GetResearchBiomarkerStats(researchID uint) (calculatedCount, totalCount int, err error) {
	var biomarkers []ds.ResearchBiomarker
	err = r.db.Where("id_research = ?", researchID).Find(&biomarkers).Error
	if err != nil {
		return 0, 0, err
	}

	totalCount = len(biomarkers)
	for _, b := range biomarkers {
		if b.PatientValue != nil {
			calculatedCount++
		}
	}

	return calculatedCount, totalCount, nil
}
