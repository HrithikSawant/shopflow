package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HrithikSawant/shopflow-api/internal/database"
	"github.com/HrithikSawant/shopflow-api/internal/repository"
)

func ListProducts(c *gin.Context) {
	products, err := repository.ListProducts(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}