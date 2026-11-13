package router

import (
	"github.com/gin-gonic/gin"

	"github.com/HrithikSawant/shopflow-api/internal/handlers"
)

func New() *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", handlers.Health)
	r.GET("/products", handlers.ListProducts)
	r.POST("/orders", handlers.CreateOrder)
	return r
}
