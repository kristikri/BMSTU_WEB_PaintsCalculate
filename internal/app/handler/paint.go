package handler

import (
    "fmt"
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
    var currentRequestID uint
    request, err := h.Repository.GetOrCreateDraftRequest(userID)
    if err != nil {
		logrus.Error("Failed to get draft request:", err)
	} else {
		currentRequestID = request.ID
	}
	paintCount := h.Repository.GetPaintCount(userID)

	ctx.HTML(http.StatusOK, "paints.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"Paints":    paints,
		"query":     searchQuery,
		"paintCount": paintCount,
        "currentRequestID": currentRequestID,
	})
}

func (h *Handler) GetPaintCalculate(ctx *gin.Context) {
    userID := uint(1)
    
    idParam := ctx.Param("id")
    if idParam == "" {
        request, err := h.Repository.GetOrCreateDraftRequest(userID)
        if err != nil {
            logrus.Error(err)
            ctx.HTML(http.StatusOK, "paints_calculate.html", gin.H{
                "Paints":     []ds.Paint{},
                "paintsCount": 0,
                "HasRequest": false,
            })
            return
        }
        ctx.Redirect(http.StatusFound, fmt.Sprintf("/paints_calculate/%d", request.ID))
        return
    }
    
    id, err := strconv.Atoi(idParam)
    if err != nil {
        logrus.Error("Invalid request ID:", err)
        request, err := h.Repository.GetOrCreateDraftRequest(userID)
        if err != nil {
            logrus.Error(err)
            ctx.HTML(http.StatusOK, "paints_calculate.html", gin.H{
                "Paints":     []ds.Paint{},
                "paintsCount": 0,
                "HasRequest": false,
            })
            return
        }
        ctx.Redirect(http.StatusFound, fmt.Sprintf("/paints_calculate/%d", request.ID))
        return
    }
    
    requestID := uint(id)
    
    _, err = h.Repository.GetRequestWithPaints(requestID)
    if err != nil {
        logrus.Error("Request not found:", err)
        request, err := h.Repository.GetOrCreateDraftRequest(userID)
        if err != nil {
            logrus.Error(err)
            ctx.HTML(http.StatusOK, "paints_calculate.html", gin.H{
                "Paints":     []ds.Paint{},
                "paintsCount": 0,
                "HasRequest": false,
            })
            return
        }
        ctx.Redirect(http.StatusFound, fmt.Sprintf("/paints_calculate/%d", request.ID))
        return
    }

    requestWithPaints, err := h.Repository.GetRequestWithPaints(requestID)
    var paints []ds.Paint
    hasRequest := false

    if err == nil {
        for _, rp := range requestWithPaints.RequestPaints {
            paints = append(paints, rp.Paint)
        }
        hasRequest = true
    }

    paintsCount := h.Repository.GetPaintCount(userID)

    ctx.HTML(http.StatusOK, "paints_calculate.html", gin.H{
        "Paints":      paints,
        "paintsCount": paintsCount,
        "HasRequest":  hasRequest,
        "RequestID":   requestID,
    })
}