package middleware

import (
	"LAB1/internal/app/ds"
	"LAB1/internal/app/role"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
)

const (
	UserIDKey    = "user_id"
	UserRoleKey  = "user_role"
	UserLoginKey = "user_login"
)

// AuthMiddleware проверяет наличие и валидность JWT токена с проверкой blacklist
func AuthMiddleware(jwtSecret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Проверяем формат Bearer {token}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use: Bearer {token}",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Валидируем токен с проверкой blacklist
		claims, err := validateTokenWithBlacklist(c.Request.Context(), tokenString, jwtSecret, redisClient)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контекст
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserLoginKey, claims.Subject)

		c.Next()
	}
}

// validateTokenWithBlacklist валидирует JWT токен и проверяет blacklist в Redis
func validateTokenWithBlacklist(ctx context.Context, tokenString, jwtSecret string, redisClient *redis.Client) (*ds.JWTClaims, error) {
	// Если Redis доступен, проверяем blacklist
	if redisClient != nil {
		isBlacklisted, err := checkJWTInBlacklist(ctx, tokenString, redisClient)
		if err != nil {
			return nil, fmt.Errorf("redis error: %v", err)
		}
		if isBlacklisted {
			return nil, errors.New("token revoked")
		}
	}

	// Затем стандартная JWT валидация
	token, err := jwt.ParseWithClaims(tokenString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	if claims, ok := token.Claims.(*ds.JWTClaims); ok && token.Valid {
		// Дополнительная проверка срока действия
		if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
			return nil, errors.New("token expired")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// checkJWTInBlacklist проверяет находится ли JWT токен в черном списке
func checkJWTInBlacklist(ctx context.Context, jwtStr string, redisClient *redis.Client) (bool, error) {
	key := "nutriscan_service.jwt." + jwtStr
	result, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// WriteJWTToBlacklist добавляет JWT токен в черный список (экспортируемая функция)
func WriteJWTToBlacklist(ctx context.Context, jwtStr string, ttl int64, redisClient *redis.Client) error {
	key := "nutriscan_service.jwt." + jwtStr
	return redisClient.Set(ctx, key, "blacklisted", time.Duration(ttl)*time.Second).Err()
}

// RoleMiddleware проверяет, что пользователь имеет достаточные права
// Должен использоваться после AuthMiddleware
func RoleMiddleware(requiredRole role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleStr, exists := c.Get(UserRoleKey)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User authentication required",
			})
			c.Abort()
			return
		}

		userRole := role.FromString(userRoleStr.(string))
		if !userRole.HasPermission(requiredRole) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions. Required: " + requiredRole.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuthMiddleware пытается извлечь данные из токена, но не требует его наличия
func OptionalAuthMiddleware(jwtSecret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			claims, err := validateTokenWithBlacklist(c.Request.Context(), tokenString, jwtSecret, redisClient)
			if err == nil {
				c.Set(UserIDKey, claims.UserID)
				c.Set(UserRoleKey, claims.Role)
				c.Set(UserLoginKey, claims.Subject)
			}
		}

		c.Next()
	}
}

// validateToken валидирует JWT токен и возвращает claims
func validateToken(tokenString, jwtSecret string) (*ds.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ds.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// Вспомогательные функции для извлечения данных из контекста

// GetUserID извлекает ID пользователя из контекста
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	// Безопасное преобразование типов
	switch v := userID.(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	case float64:
		return uint(v), true
	default:
		return 0, false
	}
}

// GetUserRole извлекает роль пользователя из контекста
func GetUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}

	if roleStr, ok := userRole.(string); ok {
		return roleStr, true
	}
	return "", false
}

// GetUserLogin извлекает логин пользователя из контекста
func GetUserLogin(c *gin.Context) (string, bool) {
	userLogin, exists := c.Get(UserLoginKey)
	if !exists {
		return "", false
	}

	if loginStr, ok := userLogin.(string); ok {
		return loginStr, true
	}
	return "", false
}

// IsAuthenticated проверяет аутентифицирован ли пользователь
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(UserIDKey)
	return exists
}

// SessionAuthMiddleware проверяет сессионные куки
func SessionAuthMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Next() // Продолжаем без аутентификации
			return
		}

		// Проверяем сессию в Redis
		sessionData, err := redisClient.HGetAll(c.Request.Context(), "nutriscan_service:sessions:"+sessionID).Result()
		if err != nil || len(sessionData) == 0 {
			c.Next() // Сессия не найдена
			return
		}

		// Сохраняем данные пользователя в контекст
		if userID, err := strconv.ParseUint(sessionData["user_id"], 10, 32); err == nil {
			c.Set(UserIDKey, uint(userID))
		}
		c.Set(UserRoleKey, sessionData["role"])
		c.Set(UserLoginKey, sessionData["login"])

		c.Next()
	}
}
