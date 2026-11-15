package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/HrithikSawant/shopflow-api/internal/repository"
)

func post(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/orders", CreateOrder)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func stubCreateOrder(t *testing.T, fn func(productID, quantity int) (int64, error)) {
	t.Helper()
	orig := createOrder
	createOrder = fn
	t.Cleanup(func() { createOrder = orig })
}

func TestCreateOrderRejectsBadInput(t *testing.T) {
	stubCreateOrder(t, func(int, int) (int64, error) {
		t.Fatal("database must not be called for invalid input")
		return 0, nil
	})

	cases := map[string]string{
		"invalid json":      `invalid-json`,
		"empty object":      `{}`,
		"zero product":      `{"productId":0,"quantity":1}`,
		"negative product":  `{"productId":-1,"quantity":1}`,
		"zero quantity":     `{"productId":1,"quantity":0}`,
		"negative quantity": `{"productId":1,"quantity":-5}`,
		"wrong type":        `{"productId":"abc","quantity":1}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, http.StatusBadRequest, post(t, body).Code)
		})
	}
}

func TestCreateOrderSuccess(t *testing.T) {
	stubCreateOrder(t, func(productID, quantity int) (int64, error) {
		assert.Equal(t, 1, productID)
		assert.Equal(t, 2, quantity)
		return 42, nil
	})

	w := post(t, `{"productId":1,"quantity":2}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t,
		`{"message":"order created","order":{"id":42,"productId":1,"quantity":2,"status":"PENDING"}}`,
		w.Body.String())
}

func TestCreateOrderErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not found", repository.ErrProductNotFound, http.StatusNotFound},
		{"no stock", repository.ErrInsufficientStock, http.StatusConflict},
		{"db failure", errors.New("connection refused to 10.0.0.5"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubCreateOrder(t, func(int, int) (int64, error) { return 0, tc.err })

			w := post(t, `{"productId":1,"quantity":1}`)

			assert.Equal(t, tc.want, w.Code)
			assert.NotContains(t, w.Body.String(), "10.0.0.5", "internal errors must not leak")
		})
	}
}
