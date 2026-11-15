package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HrithikSawant/shopflow-api/internal/database"
	"github.com/HrithikSawant/shopflow-api/internal/repository"
)

type CreateOrderRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

// createOrder is a variable so tests can replace the database layer.
var createOrder = func(productID, quantity int) (int64, error) {
	return repository.CreateOrder(database.DB, productID, quantity)
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	if req.ProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be greater than zero"})
		return
	}

	orderID, err := createOrder(req.ProductID, req.Quantity)

	switch {
	case errors.Is(err, repository.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	case errors.Is(err, repository.ErrInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient stock"})
		return
	case err != nil:
		log.Printf("create order failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// The ORDER_CREATED event is written to the outbox in the same database
	// statement and published to RabbitMQ by the relay (see internal/outbox).
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
