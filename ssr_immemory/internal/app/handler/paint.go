package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/ds"
	"ssr_immemory/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// GetPaints godoc
// @Summary Получить список красок
// @Description Возвращает список всех красок или фильтрует по названию
// @Tags paints
// @Produce json
// @Param title query string false "Фильтр по названию"
// @Success 200 {array} apitypes.PaintJSON
// @Failure 500 {object} map[string]string
// @Router /paints [get]
func (h *Handler) GetPaints(ctx *gin.Context) {
	var paints []ds.Paint
	var err error

	searchQuery := ctx.Query("title")
	if searchQuery == "" {
		paints, err = h.Repository.GetPaints()
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	} else {
		paints, err = h.Repository.GetPaintsByTitle(searchQuery)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	resp := make([]apitypes.PaintJSON, 0, len(paints))
	for _, r := range paints {
		resp = append(resp, apitypes.PaintToJSON(r))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetPaint godoc
// @Summary Получить краску по ID
// @Description Возвращает информацию о конкретной краске
// @Tags paints
// @Produce json
// @Param id path int true "ID краски"
// @Success 200 {object} apitypes.PaintJSON
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /paint/{id} [get]
func (h *Handler) GetPaint(ctx *gin.Context) {
	idStr := ctx.Param("id") 
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	paint, err := h.Repository.GetPaint(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, apitypes.PaintToJSON(*paint))
}

// CreatePaint godoc
// @Summary Создать краску
// @Description Создает новую запись о краске
// @Tags paints
// @Accept json
// @Produce json
// @Param paint body apitypes.PaintJSON true "Данные краски"
// @Success 201 {object} apitypes.PaintJSON
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /paint/create-paint [post]
func (h *Handler) CreatePaint(ctx *gin.Context) {
	var paintJSON apitypes.PaintJSON
	if err := ctx.BindJSON(&paintJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	paint, err := h.Repository.CreatePaint(paintJSON)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Header("Location", fmt.Sprintf("/paints/%v", paint.ID))
	ctx.JSON(http.StatusCreated, apitypes.PaintToJSON(paint))
}

// DeletePaint godoc
// @Summary Удалить краску
// @Description Помечает краску как удаленную
// @Tags paints
// @Produce json
// @Param id path int true "ID краски"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /paint/{id}/delete-paint [delete]
func (h *Handler) DeletePaint(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeletePaint(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "deleted",
	})
}

// ChangePaint godoc
// @Summary Изменить краску
// @Description Обновляет данные о краске
// @Tags paints
// @Accept json
// @Produce json
// @Param id path int true "ID краски"
// @Param paint body apitypes.PaintJSON true "Новые данные краски"
// @Success 200 {object} apitypes.PaintJSON
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /paint/{id}/change-paint [put]
func (h *Handler) ChangePaint(ctx *gin.Context){
	var paintJSON apitypes.PaintJSON
	if err := ctx.BindJSON(&paintJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	paint, err := h.Repository.ChangePaint(id, paintJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, apitypes.PaintToJSON(paint))
}

// AddPaintToRequest godoc
// @Summary Добавить краску в заявку
// @Description Добавляет краску в текущую заявку-черновик пользователя
// @Tags paints
// @Accept json
// @Produce json
// @Param id path int true "ID краски"
// @Param paintData body AddPaintToRequestData true "Данные для добавления"
// @Success 200 {object} apitypes.PaintRequestJSON
// @Success 201 {object} apitypes.PaintRequestJSON
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security ApiKeyAuth
// @Router /paint/{id}/add-to [post]
func (h *Handler) AddPaintToRequest(ctx *gin.Context) {
	userID, err := getUserIDFromContext(ctx) 
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	request, created, err := h.Repository.GetRequestDraft(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	requestId := request.ID

	paintId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var addPaintData struct {
		Area   float64 `json:"area" binding:"required"`
		Layers int     `json:"layers" binding:"required"`
	}
	if err := ctx.BindJSON(&addPaintData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.AddPaintToRequest(int(requestId), paintId, addPaintData.Area, addPaintData.Layers)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	
	status := http.StatusOK
	
	if created {
		ctx.Header("Location", fmt.Sprintf("/requests/%v", request.ID))
		status = http.StatusCreated
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(request)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(status, apitypes.PaintRequestToJSON(request, creatorLogin, moderatorLogin))
}

// UploadImage godoc
// @Summary Загрузить изображение краски
// @Description Загружает изображение для конкретной краски
// @Tags paints
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID краски"
// @Param image formData file true "Изображение краски"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /paint/{id}/upload-image [post]
func (h *Handler) UploadImage(ctx *gin.Context) {
	paintId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	paint, err := h.Repository.UploadImage(ctx, paintId, file)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "uploaded",
		"paint": apitypes.PaintToJSON(paint),
	})
}

type AddPaintToRequestData struct {
	Area   float64 `json:"area" binding:"required"`
	Layers int     `json:"layers" binding:"required"`
}