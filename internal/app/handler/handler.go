package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"ssr_immemory/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/api", h.GetPaints)
	router.GET("/api/paint/:id", h.GetPaint)
	router.POST("/api/paint/create-paint", h.CreatePaint)
	router.PUT("/api/paint/:id/change-paint", h.ChangePaint)
	router.DELETE("/api/paint/:id/delete-paint", h.DeletePaint)
	router.POST("api/paint/:id/add-to",h.AddPaintToRequest)
	router.POST("/api/paint/:id/upload-image", h.UploadImage)

	router.GET("/api/requests/paints-calculate", h.GetRequestCart)	
	router.GET("/api/requests", h.GetPaintRequests)
	router.GET("/api/requests/:id", h.GetRequest)
	router.PUT("/api/requests/:id/change-paint_request", h.ChangeRequest)
	router.PUT("/api/requests/:id/form-paint_request", h.FormRequest)
	router.PUT("/api/requests/:id/complete-paint_request", h.ModerateRequest)
	router.DELETE("/api/requests/:id/delete-paint_request", h.DeleteRequest)

	router.DELETE("/api/request-paints/:request_id/:paint_id", h.DeletePaintFromRequest)
	router.PUT("/api/request-paints/:request_id/:paint_id", h.ChangeRequestPaint)

	router.POST("/api/users/register", h.CreateUser)
	router.GET("/api/users/profile", h.GetProfile)
	router.PUT("/api/users/profile", h.ChangeProfile)
	router.POST("/api/users/signin", h.SignIn)
	router.POST("/api/users/signout", h.SignOut)
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