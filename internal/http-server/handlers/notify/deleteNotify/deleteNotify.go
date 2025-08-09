package deleteNotify

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
}

//go:generate go run github.com/vektra/mockery/v2@v2.51.1 --name=DeleteNotification
type DeleteNotification interface {
	DeleteNotification(notificationID int64) error
}

func New(log *slog.Logger, notify DeleteNotification) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.notify.deleteNotify.New"

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
			slog.Int64("notifyID", id),
		)

		err = notify.DeleteNotification(id)
		if errors.Is(err, storage.ErrNotifyNotFound) {
			log.Info("notify not found", slog.Int64("notify", id))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, response.Error("notify not found"))

			return
		}
		if err != nil {
			log.Error("failed to delete notify", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to delete notify"))

			return
		}

		log.Info("notify deleted", slog.Int64("id", id))

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: response.OK(),
	})
}
