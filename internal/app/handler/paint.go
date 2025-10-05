package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"ssr_immemory/internal/app/ds"
)

func (h *Handler) GetPaint(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid paint ID"})
		return
	}

	paint, err := h.Repository.GetPaint(id)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Paint not found"})
		return
	}

	ctx.HTML(http.StatusOK, "paint.html", gin.H{
		"Paint": paint,
	})
}

func (h *Handler) GetPaints(ctx *gin.Context) {
	var paints []ds.Paint
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		paints, err = h.Repository.GetPaints()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		paints, err = h.Repository.GetPaintsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	userID := uint(1)
	paintCount := h.Repository.GetPaintCount(userID)

	ctx.HTML(http.StatusOK, "paints.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"Paints":    paints,
		"query":     searchQuery,
		"paintCount": paintCount,
	})
}

func (h *Handler) GetPaintCalculate(ctx *gin.Context) {
    userID := uint(1)
    request, err := h.Repository.GetDraftRequest(userID)    
    var paints []ds.Paint
    hasRequest := false

    if err == nil {
        requestWithPaints, err := h.Repository.GetRequestWithPaints(request.ID)
        if err == nil {
            for _, rp := range requestWithPaints.RequestPaints {
                paints = append(paints, rp.Paint)
            }
            hasRequest = true
        }
    }

    paintsCount := h.Repository.GetPaintCount(userID)

    ctx.HTML(http.StatusOK, "paints_calculate.html", gin.H{
        "Paints":    paints,
        "paintsCount": paintsCount,
        "HasRequest": hasRequest,
    })
}