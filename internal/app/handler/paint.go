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

func (h *Handler) AddPaintToRequest(ctx *gin.Context) {
	request, created, err := h.Repository.GetRequestDraft(h.Repository.GetUserID())
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