package worker

import (
	"DelayedNotifier/internal/service"
	"log/slog"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

type Worker struct {
	service *service.Service // Используем сервис
	log     *slog.Logger
}

func New(service *service.Service, log *slog.Logger) *Worker {
	return &Worker{
		service: service,
		log:     log,
	}
}

func (w *Worker) Start(msgs <-chan amqp.Delivery) {
	w.log.Info("Starting RabbitMQ worker")

	for d := range msgs {
		notificationID, err := strconv.ParseInt(string(d.Body), 10, 64)
		if err != nil {
			w.log.Error("Failed to parse notification ID", "error", err)
			err = d.Ack(false)
			if err != nil {
				return
			}
			continue
		}

		notification, err := w.service.GetNotificationByID(notificationID)
		if err != nil {
			w.log.Error("Failed to get notification by ID", "error", err, "notification_id", notificationID)
			err = d.Ack(false)
			if err != nil {
				return
			}
			continue
		}

		if time.Now().After(notification.Date) {
			err = w.service.SendNotification(notification.RecipientID, notification.Text)
			newStatus := "sent"
			if err != nil {
				newStatus = "failed"
				w.log.Error("Failed to send message to Telegram", "error", err, "notification_id", notificationID)
			}

			err = w.service.UpdateNotificationStatus(notificationID, newStatus)
			if err != nil {
				return
			}
			err = d.Ack(false)
			if err != nil {
				return
			}
		} else {
			err = d.Nack(false, true)
			if err != nil {
				return
			}
		}
	}
	w.log.Info("RabbitMQ worker stopped")
}
