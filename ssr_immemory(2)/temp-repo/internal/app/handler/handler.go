package handler

import (
	"net/http"
	"strconv"
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
	router.GET("/", h.GetPaints)
	router.GET("/paint/:id", h.GetPaint)
	router.GET("/paints_calculate", h.GetPaintCalculate)        
	router.POST("/add-to-paint_request", h.AddToPaintRequest) 
	router.POST("/delete-paint_request", h.DeletePaintRequest) 
	
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) GetPaintRequest(ctx *gin.Context) {
	userID := uint(1)

	request, err := h.Repository.GetOrCreateDraftRequest(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	requestWithPaints, err := h.Repository.GetRequestWithPaints(request.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	cartCount := h.Repository.GetPaintCount(userID)

	ctx.HTML(http.StatusOK, "paints_calculate.html", gin.H{
		"Request":   requestWithPaints,
		"cartCount": cartCount,
	})
}

func (h *Handler) AddToPaintRequest(ctx *gin.Context) {
    userID := uint(1)

    paintIDStr := ctx.PostForm("paint_id")
    areaStr := ctx.PostForm("area")
    layersStr := ctx.PostForm("layers")

    paintID, err := strconv.Atoi(paintIDStr)
    if err != nil {
        h.errorHandler(ctx, http.StatusBadRequest, err)
        return
    }

    area := 10.0
    if areaStr != "" {
        area, err = strconv.ParseFloat(areaStr, 64)
        if err != nil {
            h.errorHandler(ctx, http.StatusBadRequest, err)
            return
        }
    }

    layers := 2
    if layersStr != "" {
        layers, err = strconv.Atoi(layersStr)
        if err != nil {
            h.errorHandler(ctx, http.StatusBadRequest, err)
            return
        }
    }

    request, err := h.Repository.GetOrCreateDraftRequest(userID)
    if err != nil {
        h.errorHandler(ctx, http.StatusInternalServerError, err)
        return
    }

    err = h.Repository.AddPaintToRequest(request.ID, uint(paintID), area, layers)
    if err != nil {
        h.errorHandler(ctx, http.StatusInternalServerError, err)
        return
    }

    ctx.Redirect(http.StatusFound, "/paints_calculate")
}

func (h *Handler) DeletePaintRequest(ctx *gin.Context) {
	userID := uint(1)

	request, err := h.Repository.GetDraftRequest(userID)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	err = h.Repository.DeleteRequestSQL(request.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

