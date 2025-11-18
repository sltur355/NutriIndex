package ds

import "time"

// ResearchWithBiomarkersResponse - DTO для ответа с исследованием и биомаркерами
type ResearchWithBiomarkersResponse struct {
	ID             uint              `json:"id"`
	Status         INIResearchStatus `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
	PatientName    *string           `json:"patient_name,omitempty"`
	PatientBirth   *string           `json:"patient_birth,omitempty"`
	PatientGender  *string           `json:"patient_gender,omitempty"`
	INIResult      *float64          `json:"ini_result,omitempty"`
	FormedAt       *time.Time        `json:"formed_at,omitempty"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	CreatorLogin   string            `json:"creator_login,omitempty"`
	ModeratorLogin string            `json:"moderator_login,omitempty"`

	// Биомаркеры с нужными полями для фронтенда
	Biomarkers []BiomarkerItemResponse `json:"biomarkers"`
}

// BiomarkerItemResponse - DTO для биомаркера в исследовании
type BiomarkerItemResponse struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	MeasureUnit  string   `json:"measure_unit"`
	MinValue     float64  `json:"min_value"`
	MaxValue     float64  `json:"max_value"`
	ImageURL     *string  `json:"image_url,omitempty"`
	PatientValue *float64 `json:"patient_value,omitempty"`
	Significance float64  `json:"significance"`
}

// ResearchListResponse - DTO для списка исследований
type ResearchListResponse struct {
	ID             uint              `json:"id"`
	Status         INIResearchStatus `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
	PatientName    *string           `json:"patient_name,omitempty"`
	FormedAt       *time.Time        `json:"formed_at,omitempty"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	INIResult      *float64          `json:"ini_result,omitempty"`
	CreatorLogin   string            `json:"creator_login"`
	ModeratorLogin string            `json:"moderator_login,omitempty"`
}
