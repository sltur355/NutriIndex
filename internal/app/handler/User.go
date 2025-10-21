package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"LAB1/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// RegisterUserAPI - POST /api/users/register
func (h *INIController) RegisterUserAPI(ctx *gin.Context) {
	var user ds.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if user.Login == "" || user.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("login and password are required"))
		return
	}

	// Проверяем уникальность логина
	_, err := h.INIModel.GetUserByLogin(user.Login)
	if err == nil {
		h.errorHandler(ctx, http.StatusConflict, fmt.Errorf("user with this login already exists"))
		return
	}

	createdUser, err := h.INIModel.CreateUser(user)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Не возвращаем пароль в ответе
	createdUser.Password = ""
	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    createdUser,
		"message": "User registered successfully",
	})
}

// AuthenticateUserAPI - POST /api/users/auth
func (h *INIController) AuthenticateUserAPI(ctx *gin.Context) {
	var authData struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&authData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	user, err := h.INIModel.CheckCredentials(authData.Login, authData.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	user.Password = ""
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    user,
		"message": "Authentication successful",
	})
}

// GetUserProfileAPI - GET /api/users/profile
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

	user.Password = ""
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

// UpdateUserProfileAPI - PUT /api/users/profile
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

	var user ds.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Защищаем от изменения ID, пароля и других системных полей
	user.ID = 0
	user.Password = "" // Пароль нельзя менять через этот метод

	err = h.INIModel.UpdateUser(uint(userID), user)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User profile updated successfully",
	})
}

// LogoutUserAPI - POST /api/users/logout
func (h *INIController) LogoutUserAPI(ctx *gin.Context) {
	// Заглушка - в реальной системе здесь была бы инвалидация сессии/токена
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User logged out successfully",
	})
}
