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

	// Регистрируем API маршруты вместо старых HTML маршрутов
	a.Handler.RegisterAPI(a.Router) // ← ИЗМЕНИЛИ: RegisterAPI вместо RegisterHandler

	//Регситририуем SwaggerUI
	a.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Убираем статику и шаблоны для SPA
	// a.Handler.RegisterStatic(a.Router) // ← КОММЕНТИРУЕМ

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
