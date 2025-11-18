package ds

import (
	"time"
)

type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Login        string     `gorm:"type:varchar(25);unique;not null" json:"login"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Role         string     `gorm:"type:varchar(20);default:'patient'" json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	IsDeleted    bool       `gorm:"default:false" json:"is_deleted"`
}

// UpdateUserRequest запрос на обновление пользователя
// @Description Данные для обновления профиля пользователя
type UpdateUserRequest struct {
	Login string `json:"login,omitempty" example:"new_login"`
	Role  string `json:"role,omitempty" example:"doctor" enums:"patient,doctor"`
}
