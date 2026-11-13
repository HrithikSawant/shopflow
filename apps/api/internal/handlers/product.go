package handlers

import "github.com/gin-gonic/gin"

func ListProducts(c *gin.Context) {
	c.JSON(200, gin.H{
		"products": []gin.H{
			{
				"id":          1,
				"name":        "Mechanical Keyboard",
				"description": "Hot swappable keyboard",
				"price":       4999,
			},
			{
				"id":          2,
				"name":        "Gaming Mouse",
				"description": "Wireless mouse",
				"price":       2999,
			},
		},
	})
}