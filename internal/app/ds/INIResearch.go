package ds

//заявки
import (
	"database/sql"
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
	ID        uint              `gorm:"primaryKey;autoIncrement"`
	Status    INIResearchStatus `gorm:"type:varchar(20);not null;check:status IN ('черновик','удалён','сформирован','завершён','отклонён')"`
	CreatedAt time.Time         `gorm:"not null"`
	CreatedBy uint              `gorm:"not null"`

	// Nullable поля (как в примере)
	FormedAt    sql.NullTime  `gorm:"default:null"`
	CompletedAt sql.NullTime  `gorm:"default:null"`
	ModeratorID sql.NullInt64 `gorm:"default:null"`

	// Поля по предметной области
	PatientName   string  `gorm:"type:varchar(100);not null"`
	PatientBirth  string  `gorm:"type:varchar(10);not null"`
	PatientGender string  `gorm:"type:varchar(10);not null"`
	INIResult     float64 `gorm:"type:decimal(5,2);default:null"`

	// Связи (явные, как в примере)
	User      User `gorm:"foreignKey:CreatedBy"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
}
