package getStatus

import (
	"DelayedNotifier/internal/lib/api/response"
	"DelayedNotifier/internal/lib/logger/sl"
	"DelayedNotifier/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type Response struct {
	response.Response
	Status string `json:"status"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.51.1 --name=GetNotificationStatus
type GetNotificationStatus interface {
	GetNotificationStatus(notificationID int64) (string, error)
}

func New(log *slog.Logger, notify GetNotificationStatus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.notify.getStatus.New"

		notifyID := chi.URLParam(r, "id")
		if notifyID == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "notifyID is required"})
			return
		}

		id, err := strconv.ParseInt(notifyID, 10, 64)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid notifyID"})
			return
		}

		log = log.With(
			slog.String("op", op),
			slog.Int64("notification_id", id),
		)

		status, err := notify.GetNotificationStatus(id)
		if errors.Is(err, storage.ErrNotifyNotFound) {
			log.Info("notify not found", slog.Int64("notification_id", id))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error("notify not found"))

			return
		}
		if err != nil {
			log.Error("failed to get notify status", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get notify status"))

			return
		}

		log.Info("notify status received", slog.Int64("notification_id", id))

		responseOK(w, r, status)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, status string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Status:   status,
	})
}
