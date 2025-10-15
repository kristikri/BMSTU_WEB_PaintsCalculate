package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"ssr_immemory/internal/app/config"
	"ssr_immemory/internal/app/dsn"
	"ssr_immemory/internal/app/handler"
	"ssr_immemory/internal/app/repository"
	"ssr_immemory/internal/pkg"
	_ "ssr_immemory/docs" 
)

// @title Paint Service API
// @version 1.0
// @description API для управления красками и заявками
// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@paint.com
// @license.name MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println("Connecting to database with DSN:", postgresString)

	rep, errRep := repository.NewRepository(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}