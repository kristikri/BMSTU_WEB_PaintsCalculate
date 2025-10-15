package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	

	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// GetPaintRequests godoc
// @Summary Получить список заявок
// @Description Возвращает список заявок с фильтрацией по дате и статусу
// @Tags requests
// @Produce json
// @Param from-date query string false "Начальная дата (YYYY-MM-DD)"
// @Param to-date query string false "Конечная дата (YYYY-MM-DD)"
// @Param status query string false "Статус заявки"
// @Success 200 {array} apitypes.PaintRequestJSON
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests [get]
func (h *Handler) GetPaintRequests(ctx *gin.Context) {
	fromDate := ctx.Query("from-date")
	var from = time.Time{}
	var to = time.Time{}
	if fromDate != "" {
		from1, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		from = from1
	}

	toDate := ctx.Query("to-date")
	if toDate != "" {
		to1, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		to = to1
	}

	status := ctx.Query("status")

	requests, err := h.Repository.GetPaintRequests(from, to, status)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]apitypes.PaintRequestJSON, 0, len(requests))
	for _, c := range requests {
		creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(c)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		resp = append(resp, apitypes.PaintRequestToJSON(c, creatorLogin, moderatorLogin))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetRequestCart godoc
// @Summary Получить корзину заявки
// @Description Возвращает информацию о текущей корзине заявки пользователя
// @Tags requests
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests/paints-calculate [get]
func (h *Handler) GetRequestCart(ctx *gin.Context){
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	paintsCount := h.Repository.GetPaintCount(userID)

	if paintsCount == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"status":       "no_draft",
			"paints_count": paintsCount,
		})
		return
	}

	request, err := h.Repository.CheckCurrentRequestDraft(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusUnauthorized, err)
		} else if errors.Is(err, repository.ErrNoDraft) {
			ctx.JSON(http.StatusOK, gin.H{
				"status":       "no_draft",
				"paints_count": 0,
			})
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":           request.ID,
		"paints_count": h.Repository.GetPaintCount(userID),
	})
}

// GetRequest godoc
// @Summary Получить заявку
// @Description Возвращает детальную информацию о заявке
// @Tags requests
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests/{id} [get]
func (h *Handler) GetRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	paints, request, err := h.Repository.GetRequestPaintsList(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	resp := make([]apitypes.PaintJSON, 0, len(paints))
	for _, r := range paints {
		resp = append(resp, apitypes.PaintToJSON(r))
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(request)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	requestPaints, _ := h.Repository.GetRequestPaints(request.ID)
	
	resp2 := make([]apitypes.RequestsPaintJSON, 0, len(requestPaints))
	for _, r := range requestPaints {
		resp2 = append(resp2, apitypes.RequestsPaintToJSON(r))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"request":       apitypes.PaintRequestToJSON(request, creatorLogin, moderatorLogin),
		"paints":        resp,
		"requestPaints": resp2,
	})
}

// FormRequest godoc
// @Summary Сформировать заявку
// @Description Переводит заявку из черновика в статус "сформирована"
// @Tags requests
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} apitypes.PaintRequestJSON
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests/{id}/form-paint_request [put]
func (h *Handler) FormRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	status := "сформирована"

	request, err := h.Repository.FormRequest(id, status)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(request)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, apitypes.PaintRequestToJSON(request, creatorLogin, moderatorLogin))
}

// ChangeRequest godoc
// @Summary Изменить заявку
// @Description Обновляет данные заявки
// @Tags requests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param request body apitypes.PaintRequestJSON true "Новые данные заявки"
// @Success 200 {object} apitypes.PaintRequestJSON
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests/{id}/change-paint_request [put]
func (h *Handler) ChangeRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var requestJSON apitypes.PaintRequestJSON
	if err := ctx.BindJSON(&requestJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.ChangeRequest(id, requestJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(request)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, apitypes.PaintRequestToJSON(request, creatorLogin, moderatorLogin))
}

// DeleteRequest godoc
// @Summary Удалить заявку
// @Description Переводит заявку в статус "удалён"
// @Tags requests
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests/{id}/delete-paint_request [delete]
func (h *Handler) DeleteRequest(ctx *gin.Context){
	idStr := ctx.Param("id")
	requestId, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	status := "удалён"
	
	_, err = h.Repository.FormRequest(requestId, status)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request deleted"})
}

// ModerateRequest godoc
// @Summary Модерировать заявку
// @Description Завершает или отклоняет заявку (только для модераторов)
// @Tags requests
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param status body apitypes.StatusJSON true "Статус ('завершена' или 'отклонена')"
// @Success 200 {object} apitypes.PaintRequestJSON
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /requests/{id}/complete-paint_request [put]
func (h *Handler) ModerateRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var statusJSON apitypes.StatusJSON
	if err := ctx.BindJSON(&statusJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	request, err := h.Repository.ModerateRequest(id, statusJSON.Status, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(request)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, apitypes.PaintRequestToJSON(request, creatorLogin, moderatorLogin))
}