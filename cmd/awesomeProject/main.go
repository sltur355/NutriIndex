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

// @title NutriScan API
// @version 1.0
// @description API для расчета индекса нутритивной недостаточности

// @contact.name API Support
// @contact.url http://localhost:8081
// @contact.email support@nutriscan.ru

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

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

	// Инициализируем репозиторий с MinIO (как в примере с Хроникой)
	repo, err := repository.NewINIModel(
		postgresString,
		conf.MinIO.Endpoint,
		conf.MinIO.AccessKeyID,
		conf.MinIO.SecretAccessKey,
		conf.MinIO.BucketName,
		conf.MinIO.UseSSL,
	)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Инициализируем обработчики
	hand := handler.NewINIController(repo)

	// Создаем и запускаем приложение
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
