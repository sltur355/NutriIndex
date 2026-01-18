package ds

import (
	"time"
)

// AsyncINICalculationRequest запрос на асинхронный расчет INI
type AsyncINICalculationRequest struct {
	ResearchID   uint   `json:"research_id"`
	BiomarkerIDs []uint `json:"biomarker_ids"`
	SecretKey    string `json:"secret_key"` // Ключ для авторизации
}

// AsyncINICalculationResponse ответ от асинхронного сервиса
type AsyncINICalculationResponse struct {
	Success      bool    `json:"success"`
	INIResult    float64 `json:"ini_result,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
	CalculatedAt string  `json:"calculated_at"`
}

// UpdateINIResultRequest запрос на обновление INI результата
type UpdateINIResultRequest struct {
	ResearchID uint    `json:"research_id"`
	INIResult  float64 `json:"ini_result"`
	SecretKey  string  `json:"secret_key"` // Ключ для авторизации
}

// ResearchListResponseWithCalcCount - DTO для списка исследований с количеством рассчитанных биомаркеров
type ResearchListResponseWithCalcCount struct {
	ID              uint              `json:"id"`
	Status          INIResearchStatus `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	PatientName     *string           `json:"patient_name,omitempty"`
	FormedAt        *time.Time        `json:"formed_at,omitempty"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
	INIResult       *float64          `json:"ini_result,omitempty"`
	CreatorLogin    string            `json:"creator_login"`
	ModeratorLogin  string            `json:"moderator_login,omitempty"`
	CalculatedCount int               `json:"calculated_count"` // Количество рассчитанных биомаркеров
	TotalBiomarkers int               `json:"total_biomarkers"` // Общее количество биомаркеров
}
