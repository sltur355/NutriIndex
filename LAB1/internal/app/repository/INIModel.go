package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type INIModel struct {
	db *gorm.DB
}

func NewINIModel(dsn string) (*INIModel, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
	if err != nil {
		return nil, err
	}

	// Возвращаем объект Repository с подключенной базой данных
	return &INIModel{
		db: db,
	}, nil
}
