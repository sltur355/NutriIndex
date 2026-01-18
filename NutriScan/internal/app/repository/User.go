package repository

import (
	"LAB1/internal/app/ds"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GenerateToken создает JWT токен для пользователя
func (r *INIModel) GenerateToken(user *ds.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // 24 часа
			IssuedAt:  time.Now().Unix(),
			Issuer:    "nutriscan-app",
		},
		UserID: user.ID,
		Role:   user.Role,
	})

	return token.SignedString([]byte(r.jwtSecret))
}

// HashPassword хэширует пароль с использованием bcrypt
func (r *INIModel) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword проверяет пароль с bcrypt
func (r *INIModel) CheckPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// CreateUser создает нового пользователя с хэшированным паролем
func (r *INIModel) CreateUser(user ds.User) (ds.User, error) {
	// Проверяем обязательные поля
	if user.Login == "" {
		return ds.User{}, errors.New("login is required")
	}
	if user.PasswordHash == "" {
		return ds.User{}, errors.New("password is required")
	}

	// Хэшируем пароль перед сохранением (user.PasswordHash содержит plain password)
	hashedPassword, err := r.HashPassword(user.PasswordHash)
	if err != nil {
		return ds.User{}, err
	}
	user.PasswordHash = hashedPassword

	// Устанавливаем createdAt если не установлено
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	err = r.db.Create(&user).Error
	return user, err
}

// CheckCredentials проверяет логин и пароль пользователя
func (r *INIModel) CheckCredentials(login, password string) (ds.User, error) {
	user, err := r.GetUserByLogin(login)
	if err != nil {
		return ds.User{}, err
	}

	if !r.CheckPassword(user.PasswordHash, password) {
		return ds.User{}, errors.New("invalid credentials")
	}

	return user, nil
}

// GetUserByID ищет пользователя по ID
func (r *INIModel) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ? AND is_deleted = false", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, errors.New("user not found")
		}
		return ds.User{}, err
	}
	return user, nil
}

// GetUserByLogin ищет пользователя по логину
func (r *INIModel) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ? AND is_deleted = false", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, errors.New("user not found")
		}
		return ds.User{}, err
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя
func (r *INIModel) UpdateUser(id uint, updates map[string]interface{}) error {
	return r.db.Model(&ds.User{}).Where("id = ? AND is_deleted = false", id).Updates(updates).Error
}

// ParseToken парсит и валидирует JWT токен
func (r *INIModel) ParseToken(tokenString string) (*ds.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ds.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
