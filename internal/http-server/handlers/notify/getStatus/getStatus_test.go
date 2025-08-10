package getStatus

import (
	"DelayedNotifier/internal/http-server/handlers/notify/getStatus/mocks"
	"DelayedNotifier/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetStatus_Success(t *testing.T) {
	mockStorage := new(mocks.GetNotificationStatus)
	mockStorage.On("GetNotificationStatus", int64(1)).Return("delivered", nil)

	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/status/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "delivered", resp.Status)

	mockStorage.AssertExpectations(t)
}

func TestHandler_GetStatus_InvalidID(t *testing.T) {
	mockStorage := new(mocks.GetNotificationStatus)
	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/status/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockStorage.AssertNotCalled(t, "GetNotificationStatus")
}

func TestHandler_GetStatus_MissingID(t *testing.T) {
	mockStorage := new(mocks.GetNotificationStatus)
	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/status/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockStorage.AssertNotCalled(t, "GetNotificationStatus")
}

func TestHandler_GetStatus_NotFound(t *testing.T) {
	mockStorage := new(mocks.GetNotificationStatus)
	mockStorage.On("GetNotificationStatus", int64(999)).Return("", storage.ErrNotifyNotFound)

	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/status/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockStorage.AssertExpectations(t)
}
func TestHandler_GetStatus_InternalError(t *testing.T) {
	mockStorage := new(mocks.GetNotificationStatus)
	mockStorage.On("GetNotificationStatus", int64(1)).Return("", errors.New("database error"))

	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/status/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockStorage.AssertExpectations(t)
}
