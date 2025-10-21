package repository

import (
	"LAB1/internal/app/ds"
	"errors"

	"gorm.io/gorm"
)

func (r *INIModel) CreateUser(user ds.User) (ds.User, error) {
	// Устанавливаем дефолтное значение для логина если пусто
	if user.Login == "" {
		return ds.User{}, errors.New("login is required")
	}

	err := r.db.Create(&user).Error
	return user, err
}

func (r *INIModel) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, errors.New("user not found")
		}
		return ds.User{}, err
	}
	return user, nil
}

func (r *INIModel) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, errors.New("user not found")
		}
		return ds.User{}, err
	}
	return user, nil
}

func (r *INIModel) UpdateUser(id uint, user ds.User) error {
	// Разрешаем обновлять только определенные поля
	updates := map[string]interface{}{
		"login": user.Login,
	}

	// is_moderator может обновлять только администратор (в реальной системе)
	// Для лабораторной разрешаем обновление
	updates["is_moderator"] = user.IsModerator

	return r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *INIModel) CheckCredentials(login, password string) (ds.User, error) {
	user, err := r.GetUserByLogin(login)
	if err != nil {
		return ds.User{}, err
	}
	if user.Password != password {
		return ds.User{}, errors.New("invalid credentials")
	}
	return user, nil
}
