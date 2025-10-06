package api

import (
	"LAB1/internal/app/handler"
	"LAB1/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewINIModel()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	h := handler.NewINIController(repo)

	r := gin.Default()
	// добавляем наш html/шаблон
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")
	// слева название папки, в которую выгрузится наша статика
	// справа путь к папке, в которой лежит статика

	// Маршруты для биомаркеров:

	r.GET("/biomarkers", h.GetBiomarkers)
	r.GET("/biomarkers/:id", h.GetDetailedBiomarker)
	r.GET("/INIresearch/:id", h.GetINIresearch)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}
