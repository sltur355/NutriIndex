// C:\RIP\LAB1\cmd\awesomeProject\main.go
package main

import (
	"LAB1/internal/app/config"
	"LAB1/internal/app/dsn"
	"LAB1/internal/app/handler"
	"LAB1/internal/app/repository"
	"LAB1/internal/pkg"
	"context"
	"strconv" // ДОБАВЛЯЕМ этот импорт

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
// @description Введите "Bearer [ваш_JWT_токен]" для авторизации

// @x-extension-openapi {"example": "value"}
func main() {
	router := gin.Default()
	ctx := context.Background()

	// Загружаем конфигурацию
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	logrus.Infof("Service will run on: %s:%d", conf.ServiceHost, conf.ServicePort)

	// Получаем DSN строку
	postgresString := dsn.FromEnv()
	logrus.Info("Connecting to database...")

	// Инициализируем репозиторий с MinIO
	repo, err := repository.NewINIModel(
		postgresString,
		conf.Minio.Endpoint,
		conf.Minio.AccessKey,
		conf.Minio.SecretKey,
		conf.Minio.Bucket,
		conf.Minio.UseSSL,
	)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Инициализируем Redis клиент
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Host + ":" + strconv.Itoa(conf.Redis.Port),
		Password: conf.Redis.Password,
		DB:       0,
	})

	// Проверяем подключение к Redis
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("error connecting to Redis: %v", err)
	}
	logrus.Info("Connected to Redis successfully")

	// Инициализируем обработчики с Redis клиентом
	hand := handler.NewINIController(repo, redisClient)

	// Создаем и запускаем приложение
	application := pkg.NewApp(conf, router, hand)

	logrus.Infof("Starting NutriScan API server on %s:%d", conf.ServiceHost, conf.ServicePort)
	application.RunApp()
}
