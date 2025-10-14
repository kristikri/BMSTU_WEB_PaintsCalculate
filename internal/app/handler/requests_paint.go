package handler

import (
	apitypes "ssr_immemory/internal/app/api_types"
	"ssr_immemory/internal/app/repository"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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