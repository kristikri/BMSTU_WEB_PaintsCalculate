package handler

import (
	"errors"
	"net/http"
	"ssr_immemory/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler godoc
// @title Paint Service API
// @version 1.0
// @description API для управления красками и заявками
// @contact.name API Support
// @contact.url http://localhost
// @contact.email support@paint.com
// @license.name MIT
// @host localhost
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.Use(CORSMiddleware())

	api := router.Group("/api/v1") 

	unauthorized := api.Group("/")
	unauthorized.POST("/users/register", h.CreateUser)
	unauthorized.POST("/users/signin", h.SignIn)
	unauthorized.GET("/paints", h.GetPaints)
	unauthorized.GET("/paint/:id", h.GetPaint)

	authorized := api.Group("/")
	authorized.Use(h.ModeratorMiddleware(false))
	authorized.POST("/paint/:id/add-to", h.AddPaintToRequest)
	authorized.POST("/paint/create-paint", h.CreatePaint)
	authorized.PUT("/paint/:id/change-paint", h.ChangePaint)
	authorized.DELETE("/paint/:id/delete-paint", h.DeletePaint)
	authorized.POST("/paint/:id/upload-image", h.UploadImage)
	unauthorized.PUT("/requests/:id/paint_quantity", h.UpdatePaintQuantity)

	optionalauthorized := api.Group("/")
	optionalauthorized.Use(h.WithOptionalAuthCheck())
	optionalauthorized.GET("/requests/paints-calculate", h.GetRequestCart)
	authorized.GET("/requests", h.GetPaintRequests)
	authorized.GET("/requests/:id", h.GetRequest)
	authorized.PUT("/requests/:id/change-paint_request", h.ChangeRequest)
	authorized.PUT("/requests/:id/form-paint_request", h.FormRequest)
	authorized.DELETE("/requests/:id/delete-paint_request", h.DeleteRequest)

	authorized.DELETE("/request-paints/:request_id/:paint_id", h.DeletePaintFromRequest)
	authorized.PUT("/request-paints/:request_id/:paint_id", h.ChangeRequestPaint)

	authorized.GET("/users/profile", h.GetProfile)
	authorized.PUT("/users/profile", h.ChangeProfile)
	authorized.POST("/users/signout", h.SignOut)

	moderator := api.Group("/")
	moderator.Use(h.ModeratorMiddleware(true))
	moderator.PUT("/requests/:id/complete-paint_request", h.ModerateRequest)

	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	router.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./resources/styles")
	router.Static("/img", "./resources/img")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())

	var errorMessage string
	switch {
	case errors.Is(err, repository.ErrNotFound):
		errorMessage = "Не найден"
	case errors.Is(err, repository.ErrAlreadyExists):
		errorMessage = "Уже существует"
	case errors.Is(err, repository.ErrNotAllowed):
		errorMessage = "Доступ запрещен"
	case errors.Is(err, repository.ErrNoDraft):
		errorMessage = "Черновик не найден"
	default:
		errorMessage = err.Error()
	}

	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": errorMessage,
	})
}