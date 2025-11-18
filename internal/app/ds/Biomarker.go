package ds

import (
	"time"
)

// @Description Медицинский биомаркер для исследований
type Biomarker struct {
	ID              uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string  `gorm:"type:varchar(100);not null" json:"name"`
	Description     string  `gorm:"type:text" json:"description"`
	FullDescription string  `gorm:"type:text" json:"full_description"`
	MeasureUnit     string  `gorm:"type:varchar(50);not null" json:"measure_unit"`
	MinValue        float64 `gorm:"type:decimal(10,2);not null" json:"min_value"`
	MaxValue        float64 `gorm:"type:decimal(10,2);not null" json:"max_value"`
	Significance    float64 `gorm:"type:decimal(5,4);not null" json:"significance"`
	ImageURL        *string `gorm:"type:varchar(255);default:null" json:"image_url,omitempty"`
	IsActive        bool    `gorm:"type:boolean;default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
