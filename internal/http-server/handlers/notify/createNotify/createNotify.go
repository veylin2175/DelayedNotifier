package createNotify

import (
	"DelayedNotifier/internal/lib/api/response"
	"DelayedNotifier/internal/lib/logger/sl"
	"DelayedNotifier/internal/storage"
	"errors"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

type Request struct {
	RecipientID int64  `json:"recipient_id" validate:"required"`
	Date        string `json:"date" validate:"required"`
	Text        string `json:"text" validate:"required"`
}

type Response struct {
	response.Response
	NotificationID int64 `json:"notification_id"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.51.1 --name=CreateNotification
type CreateNotification interface {
	CreateNotification(recipientID int64, dateStr, text string) (int64, error)
}

func New(log *slog.Logger, notify CreateNotification) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.notify.createNotify.New"

		log = log.With(
			slog.String("op", op),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		notifyId, err := notify.CreateNotification(req.RecipientID, req.Date, req.Text)
		if errors.Is(err, storage.ErrNotifyExists) {
			log.Info("notify already exists", slog.Int64("notification_id", notifyId))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, response.Error("notify already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add notify", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to add notify"))

			return
		}

		log.Info("notify added", slog.Int64("notification_id", notifyId))

		responseOK(w, r, notifyId)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, notificationId int64) {
	render.JSON(w, r, Response{
		Response:       response.OK(),
		NotificationID: notificationId,
	})
}
