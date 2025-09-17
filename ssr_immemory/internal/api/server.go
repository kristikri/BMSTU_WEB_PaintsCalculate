package api

import (
  "github.com/gin-gonic/gin"
  "github.com/sirupsen/logrus"
  "log"
  "ssr_immemory/internal/app/handler"
  "ssr_immemory/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/hello", handler.GetOrders)
	r.GET("/order/:id", handler.GetOrder) 
    r.GET("/calculate", handler.GetCalculate)
	r.Run()
	log.Println("Server down")
}