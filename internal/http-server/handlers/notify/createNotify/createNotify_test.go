package createNotify

import (
	"DelayedNotifier/internal/http-server/handlers/notify/createNotify/mocks"
	"DelayedNotifier/internal/storage"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateNotify_Success(t *testing.T) {
	mockStorage := new(mocks.CreateNotification)
	mockStorage.On(
		"CreateNotification",
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(int64(1), nil)

	h := New(slog.Default(), mockStorage)

	reqBody := `{"recipient_id": 123, "date": "2024-01-01", "text": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp Response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.NotificationID)

	mockStorage.AssertExpectations(t)
}

func TestHandler_CreateNotify_ValidationError(t *testing.T) {
	mockStorage := new(mocks.CreateNotification)
	h := New(slog.Default(), mockStorage)

	reqBody := `{"date": "2024-01-01", "text": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	mockStorage.AssertNotCalled(t, "CreateNotification")
}

func TestHandler_CreateNotify_Conflict(t *testing.T) {
	mockStorage := new(mocks.CreateNotification)
	mockStorage.On(
		"CreateNotification",
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(int64(0), storage.ErrNotifyExists)

	h := New(slog.Default(), mockStorage)

	reqBody := `{"recipient_id": 123, "date": "2024-01-01", "text": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockStorage.AssertExpectations(t)
}

func TestHandler_CreateNotify_InternalError(t *testing.T) {
	mockStorage := new(mocks.CreateNotification)
	mockStorage.On(
		"CreateNotification",
		mock.AnythingOfType("int64"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(int64(0), errors.New("some internal error"))

	h := New(slog.Default(), mockStorage)

	reqBody := `{"recipient_id": 123, "date": "2024-01-01", "text": "Test"}`
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockStorage.AssertExpectations(t)
}
