package handler

import (
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/repository"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeletePaintFromRequest godoc
// @Summary Удалить краску из заявки
// @Description Удаляет краску из указанной заявки
// @Tags request-paints
// @Produce json
// @Param request_id path int true "ID заявки"
// @Param paint_id path int true "ID краски"
// @Success 200 {object} apitypes.PaintRequestJSON
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /request-paints/{request_id}/{paint_id} [delete]
func (h *Handler) DeletePaintFromRequest(ctx *gin.Context) {
	requestId, err := strconv.Atoi(ctx.Param("request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	paintId, err := strconv.Atoi(ctx.Param("paint_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.DeletePaintFromRequest(uint(requestId), uint(paintId))
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

// ChangeRequestPaint godoc
// @Summary Изменить краску в заявке
// @Description Обновляет данные краски в заявке (площадь, слои, количество)
// @Tags request-paints
// @Accept json
// @Produce json
// @Param request_id path int true "ID заявки"
// @Param paint_id path int true "ID краски"
// @Param requestPaint body apitypes.RequestsPaintJSON true "Новые данные краски в заявке"
// @Success 200 {object} apitypes.RequestsPaintJSON
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /request-paints/{request_id}/{paint_id} [put]
func (h *Handler) ChangeRequestPaint(ctx *gin.Context) {
	requestId, err := strconv.Atoi(ctx.Param("request_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	paintId, err := strconv.Atoi(ctx.Param("paint_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var requestPaintJSON apitypes.RequestsPaintJSON
	if err := ctx.BindJSON(&requestPaintJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	requestPaint, err := h.Repository.ChangeRequestPaint(uint(requestId), uint(paintId), requestPaintJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, apitypes.RequestsPaintToJSON(requestPaint))
}