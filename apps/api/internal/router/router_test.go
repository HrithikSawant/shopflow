package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-api/internal/router"
)

func TestHealthRoute(t *testing.T) {
	r := router.New()

	req := httptest.NewRequest(
		http.MethodGet,
		"/healthz",
		nil,
	)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
