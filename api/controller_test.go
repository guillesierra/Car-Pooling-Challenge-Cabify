package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/service"
)

// TestStatus checks /status health endpoint.
func TestStatus(t *testing.T) {
	router := NewController(service.New_CarPool())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/status", nil)
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAPI checks the main happy path of the API.
func TestAPI(t *testing.T) {
	router := NewController(service.New_CarPool())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/cars", strings.NewReader(`
	[
		{ "id": 1, "seats": 4 },
		{ "id": 2, "seats": 6 }
	]`))
	req.Header = map[string][]string{"Content-Type": {"application/json"}}
	router.engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/journey", strings.NewReader(`
	{ "id": 1, "people": 4 }
	`))
	req.Header = map[string][]string{"Content-Type": {"application/json"}}
	router.engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/locate", strings.NewReader("ID=1"))
	req.Header = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}
	router.engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"id":1,"seats":4}`, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/dropoff", strings.NewReader("ID=1"))
	req.Header = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}
	router.engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
