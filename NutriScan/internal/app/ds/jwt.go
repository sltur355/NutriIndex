package ds

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserID uint   `json:"user_id"`
	Role   string `json:"role"` // "patient", "doctor"
}

// LoginRequest запрос на аутентификацию
// @Description Данные для входа в систему
type LoginRequest struct {
	Login    string `json:"login" binding:"required" example:"doctor_user"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

// RegisterRequest запрос на регистрацию
// @Description Данные для регистрации нового пользователя
type RegisterRequest struct {
	Login    string `json:"login" binding:"required" example:"new_user"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
	Role     string `json:"role,omitempty" example:"patient" enums:"patient,doctor"`
}

// AuthResponse ответ с токеном
// @Description Ответ с JWT токеном после успешной аутентификации
type AuthResponse struct {
	AccessToken string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string    `json:"token_type" example:"Bearer"`
	ExpiresAt   time.Time `json:"expires_at" example:"2023-12-31T23:59:59Z"`
	UserID      uint      `json:"user_id" example:"1"`
	Role        string    `json:"role" example:"doctor"`
}
