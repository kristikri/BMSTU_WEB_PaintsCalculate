package handler 

import (
	"ssr_immemory/internal/app/repository"
	"github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
	"net/http"
    "time"
	"strconv"
)

type Handler struct{
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler{
	return &Handler{
		Repository:r,
	}
}


func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id") 
	id, err := strconv.Atoi(idStr) 
	if err != nil {
		logrus.Error(err)
	}

	Order, err := h.Repository.GetOrder(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"Order": Order,
	})
}

func (h *Handler) GetOrders(ctx *gin.Context) {
	var Orders []repository.Order
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {            
		Orders, err = h.Repository.GetOrders()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		Orders, err = h.Repository.GetOrdersByTitle(searchQuery) 
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":   time.Now().Format("15:04:05"),
		"Orders": Orders,
		"query":  searchQuery, 
	})
}


func (h *Handler) GetCalculate(ctx *gin.Context) {
    Orders, err := h.Repository.GetOrders()
    if err != nil {
        logrus.Error(err)
    }
    ctx.HTML(http.StatusOK, "calculate.html", gin.H{
        "Orders": Orders,
    })
}