package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"LAB1/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserProfileAPI godoc
// @Summary Получить профиль пользователя
// @Description Возвращает информацию о пользователе по ID
// @Tags Пользователи
// @Security ApiKeyAuth
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Success 200 {object} map[string]interface{} "Успешный ответ"
// @Failure 400 {object} map[string]interface{} "Неверный ID"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 404 {object} map[string]interface{} "Пользователь не найден"
// @Router /users/profile [get]
func (h *INIController) GetUserProfileAPI(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("user_id is required"))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid user_id format"))
		return
	}

	user, err := h.INIModel.GetUserByID(uint(userID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	userResponse := gin.H{
		"id":    user.ID,
		"login": user.Login,
		"role":  user.Role,
	}

	ctx.JSON(http.StatusOK, userResponse)
}

// LoginUserAPI godoc
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя в систему и возвращает JWT токен
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param login body ds.LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]interface{} "Успешная аутентификация"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Неверные учетные данные"
// @Router /auth/login [post]
func (h *INIController) LoginUserAPI(ctx *gin.Context) {
	var authData ds.LoginRequest

	if err := ctx.ShouldBindJSON(&authData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Используем CheckCredentials вместо прямой проверки
	user, err := h.INIModel.CheckCredentials(authData.Login, authData.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	// Генерируем JWT токен
	token, err := h.INIModel.GenerateToken(&user)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %v", err))
		return
	}

	// Создаем ответ с токеном
	userResponse := gin.H{
		"id":    user.ID,
		"login": user.Login,
		"role":  user.Role,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": userResponse,
		"auth": ds.AuthResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserID:      user.ID,
			Role:        user.Role,
		},
		"message": "Authentication successful",
	})
}

// RegisterUserAPI godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя в системе
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param register body ds.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} map[string]interface{} "Пользователь создан"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 409 {object} map[string]interface{} "Пользователь уже существует"
// @Router /auth/register [post]
func (h *INIController) RegisterUserAPI(ctx *gin.Context) {
	var req ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("login and password are required"))
		return
	}

	// Проверяем уникальность логина
	_, err := h.INIModel.GetUserByLogin(req.Login)
	if err == nil {
		h.errorHandler(ctx, http.StatusConflict, fmt.Errorf("user with this login already exists"))
		return
	}

	// Устанавливаем роль по умолчанию если не указана
	if req.Role == "" {
		req.Role = "patient" // По умолчанию регистрируем как пациента
	}

	// Создаем пользователя
	user := ds.User{
		Login:        req.Login,
		PasswordHash: req.Password, // Пароль будет хэширован в репозитории
		Role:         req.Role,
	}

	createdUser, err := h.INIModel.CreateUser(user)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Генерируем токен для автоматического входа после регистрации
	token, err := h.INIModel.GenerateToken(&createdUser)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %v", err))
		return
	}

	// Не возвращаем пароль в ответе
	createdUser.PasswordHash = ""

	ctx.JSON(http.StatusCreated, gin.H{
		"data": createdUser,
		"auth": ds.AuthResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserID:      createdUser.ID,
			Role:        createdUser.Role,
		},
		"message": "User registered successfully",
	})
}

// UpdateUserProfileAPI godoc
// @Summary Обновить профиль пользователя
// @Description Обновляет информацию о пользователе
// @Tags Пользователи
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Param updates body ds.UpdateUserRequest true "Обновляемые поля"
// @Success 200 {object} map[string]interface{} "Профиль обновлен"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Failure 404 {object} map[string]interface{} "Пользователь не найден"
// @Router /users/profile [put]
func (h *INIController) UpdateUserProfileAPI(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("user_id is required"))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid user_id format"))
		return
	}

	var updates ds.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Создаем map для обновления
	updateMap := make(map[string]interface{})
	if updates.Login != "" {
		updateMap["login"] = updates.Login
	}
	if updates.Role != "" {
		updateMap["role"] = updates.Role
	}

	if len(updateMap) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("no fields to update"))
		return
	}

	err = h.INIModel.UpdateUser(uint(userID), updateMap)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User profile updated successfully",
	})
}

// LogoutUserAPI godoc
// @Summary Выход из системы
// @Description Выполняет выход пользователя и добавляет токен в blacklist
// @Tags Аутентификация
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Успешный выход"
// @Failure 400 {object} map[string]interface{} "Неверный токен"
// @Failure 401 {object} map[string]interface{} "Не авторизован"
// @Router /auth/logout [post]
func (h *INIController) LogoutUserAPI(ctx *gin.Context) {
	// Получаем токен из заголовка Authorization
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("authorization header is required"))
		return
	}

	// Проверяем формат Bearer {token}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid authorization header format"))
		return
	}

	tokenString := parts[1]

	// Парсим токен чтобы получить expiration time
	token, err := h.INIModel.ParseToken(tokenString)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid token"))
		return
	}

	// Вычисляем оставшееся время жизни токена
	ttl := token.ExpiresAt - time.Now().Unix()
	if ttl <= 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("token already expired"))
		return
	}

	// Добавляем токен в Redis blacklist
	err = h.redisClient.Set(ctx.Request.Context(),
		"nutriscan_service.jwt."+tokenString,
		"blacklisted",
		time.Duration(ttl)*time.Second).Err()

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to logout: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User logged out successfully",
	})
}

// SessionLoginAPI godoc
// @Summary Аутентификация через сессии (куки)
// @Description Выполняет вход пользователя и устанавливает сессионную куку
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param login body ds.LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]interface{} "Успешная аутентификация"
// @Router /auth/session-login [post]
func (h *INIController) SessionLoginAPI(ctx *gin.Context) {
	var authData ds.LoginRequest

	if err := ctx.ShouldBindJSON(&authData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	user, err := h.INIModel.CheckCredentials(authData.Login, authData.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	// Создаем сессию в Redis
	sessionID := uuid.New().String()
	sessionData := map[string]interface{}{
		"user_id": user.ID,
		"login":   user.Login,
		"role":    user.Role,
	}

	// Сохраняем сессию в Redis на 24 часа
	err = h.redisClient.HSet(ctx.Request.Context(), "nutriscan_service:sessions:"+sessionID, sessionData).Err()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to create session: %v", err))
		return
	}
	h.redisClient.Expire(ctx.Request.Context(), "nutriscan_service:sessions:"+sessionID, 24*time.Hour)

	// Устанавливаем куку
	ctx.SetCookie("session_id", sessionID, 3600*24, "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Session created successfully",
		"user": gin.H{
			"id":    user.ID,
			"login": user.Login,
			"role":  user.Role,
		},
	})
}
