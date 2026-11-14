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

	if req.ProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid product id",
		})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "quantity must be greater than zero",
		})
		return
	}

	product, err := repository.GetProductByID(
		database.DB,
		req.ProductID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "product not found",
		})
		return
	}

	totalAmount :=
		product.Price * float64(req.Quantity)

	orderID, err := repository.CreateOrder(
		database.DB,
		req.ProductID,
		req.Quantity,
		totalAmount,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)
		return
	}

	err = rabbitmq.PublishOrderCreated(
		rabbitmq.URL,
		orderID,
		req.ProductID,
		req.Quantity,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"error": err.Error(),
			},
		)
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
