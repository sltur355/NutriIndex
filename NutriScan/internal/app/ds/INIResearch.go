package ds

import (
	"time"
)

type INIResearchStatus string

const (
	StatusDraft     INIResearchStatus = "черновик"
	StatusDeleted   INIResearchStatus = "удалён"
	StatusFormed    INIResearchStatus = "сформирован"
	StatusCompleted INIResearchStatus = "завершён"
	StatusRejected  INIResearchStatus = "отклонён"
)

type INIResearch struct {
	ID        uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	Status    INIResearchStatus `gorm:"type:varchar(20);not null;check:status IN ('черновик','удалён','сформирован','завершён','отклонён')" json:"status"`
	CreatedAt time.Time         `gorm:"not null" json:"created_at"`
	CreatedBy uint              `gorm:"not null" json:"created_by"`

	// Nullable поля - заменяем sql.Null* на указатели
	FormedAt    *time.Time `gorm:"default:null" json:"formed_at,omitempty"`
	CompletedAt *time.Time `gorm:"default:null" json:"completed_at,omitempty"`
	ModeratorID *uint      `gorm:"default:null" json:"moderator_id,omitempty"`

	// Поля по предметной области - теперь nullable
	PatientName   *string  `gorm:"type:varchar(100);default:null" json:"patient_name,omitempty"`
	PatientBirth  *string  `gorm:"type:varchar(10);default:null" json:"patient_birth,omitempty"`
	PatientGender *string  `gorm:"type:varchar(10);default:null" json:"patient_gender,omitempty"`
	INIResult     *float64 `gorm:"type:decimal(5,2);default:null" json:"ini_result,omitempty"`

	// Связи
	User      *User `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
	Moderator *User `gorm:"foreignKey:ModeratorID" json:"moderator,omitempty"`
}
