package deleteNotify

import (
	"DelayedNotifier/internal/http-server/handlers/notify/deleteNotify/mocks"
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

func TestHandler_DeleteNotify_Success(t *testing.T) {
	mockStorage := new(mocks.DeleteNotification)
	mockStorage.On("DeleteNotification", int64(1)).Return(nil)

	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodDelete, "/notify/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestHandler_DeleteNotify_InvalidID(t *testing.T) {
	mockStorage := new(mocks.DeleteNotification)
	h := New(slog.Default(), mockStorage)

	// ID = "abc" (не число)
	req := httptest.NewRequest(http.MethodDelete, "/notify/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockStorage.AssertNotCalled(t, "DeleteNotification")
}

func TestHandler_DeleteNotify_NotFound(t *testing.T) {
	mockStorage := new(mocks.DeleteNotification)
	mockStorage.On("DeleteNotification", int64(999)).Return(storage.ErrNotifyNotFound)

	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodDelete, "/notify/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_DeleteNotify_InternalError(t *testing.T) {
	mockStorage := new(mocks.DeleteNotification)
	mockStorage.On("DeleteNotification", int64(1)).Return(errors.New("database error"))

	h := New(slog.Default(), mockStorage)

	req := httptest.NewRequest(http.MethodDelete, "/notify/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockStorage.AssertExpectations(t)
}
