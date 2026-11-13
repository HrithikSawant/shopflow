// internal/handlers/order.go

package handlers

import "github.com/gin-gonic/gin"

type CreateOrderRequest struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"message": "order created",
		"order": gin.H{
			"id":        1,
			"productId": req.ProductID,
			"quantity":  req.Quantity,
			"status":    "PENDING",
		},
	})
}