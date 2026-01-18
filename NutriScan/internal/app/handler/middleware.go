package handler

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware добавляет CORS headers для всех запросов
func (h *INIController) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// разрешаем запросы с GitHub Pages
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
