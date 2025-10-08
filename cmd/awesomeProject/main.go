package main

import (
	"LAB1/internal/app/config"
	"LAB1/internal/app/dsn"
	"LAB1/internal/app/handler"
	"LAB1/internal/app/repository"
	"LAB1/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	// Загружаем конфигурацию
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Получаем DSN строку
	postgresString := dsn.FromEnv()
	logrus.Info("Connecting to database...")

	// Инициализируем репозиторий
	repo, err := repository.NewINIModel(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Инициализируем обработчики
	hand := handler.NewINIController(repo)

	// Создаем и запускаем приложение
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
