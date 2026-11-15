package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-api/internal/handlers"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/healthz", handlers.Health)

	req := httptest.NewRequest(
		http.MethodGet,
		"/healthz",
		nil,
	)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(
		t,
		`{"status":"ok"}`,
		w.Body.String(),
	)
}
