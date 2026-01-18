package pkg

import (
	"fmt"

	"LAB1/internal/app/config"
	"LAB1/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "LAB1/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type NutriScanApplication struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.INIController
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.INIController) *NutriScanApplication {
	return &NutriScanApplication{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *NutriScanApplication) RunApp() {
	logrus.Info("Server start up")

	// Регистрируем API маршруты
	a.Handler.RegisterAPI(a.Router)

	// Регистрируем SwaggerUI
	a.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Гарантируем использование порта из конфигурации
	port := a.Config.ServicePort
	host := a.Config.ServiceHost

	serverAddress := fmt.Sprintf("%s:%d", host, port)

	logrus.Infof("Listening and serving HTTP on %s", serverAddress)

	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
