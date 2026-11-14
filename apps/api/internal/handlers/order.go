package handlers

import (
	"net/http"

	"github.com/HrithikSawant/shopflow-api/internal/database"
	"github.com/HrithikSawant/shopflow-api/internal/rabbitmq"
	"github.com/HrithikSawant/shopflow-api/internal/repository"
	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	orderID, err := repository.CreateOrder(
		database.DB,
		9999,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = rabbitmq.PublishOrderCreated(
		rabbitmq.URL,
		orderID,
		req.ProductID,
		req.Quantity,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "order created",
		"order": gin.H{
			"id":        orderID,
			"productId": req.ProductID,
			"quantity":  req.Quantity,
			"status":    "PENDING",
		},
	})
}
