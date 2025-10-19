package ds

//услуги
import (
	"database/sql"
	"time"
)

type Biomarker struct {
	ID              uint           `gorm:"primaryKey;autoIncrement"`
	Name            string         `gorm:"type:varchar(100);not null"`
	Description     string         `gorm:"type:text"`
	FullDescription string         `gorm:"type:text"`
	MeasureUnit     string         `gorm:"type:varchar(50);not null"`
	MinValue        float64        `gorm:"type:decimal(10,2);not null"`
	MaxValue        float64        `gorm:"type:decimal(10,2);not null"`
	Significance    float64        `gorm:"type:decimal(5,4);not null"`
	ImageURL        sql.NullString `gorm:"type:varchar(255);default:null"`
	IsActive        bool           `gorm:"type:boolean;default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
