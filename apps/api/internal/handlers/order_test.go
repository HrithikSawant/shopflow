package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-api/internal/handlers"
)

func TestCreateOrderInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/orders", handlers.CreateOrder)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(`invalid-json`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}