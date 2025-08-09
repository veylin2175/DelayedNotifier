package service

import (
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/models"
	"DelayedNotifier/internal/rabbitMQ/broker"
	"DelayedNotifier/internal/storage/postgres"
	"DelayedNotifier/internal/telegram/notifier"
	"fmt"
	"strconv"
)

type Service struct {
	storage  *postgres.Storage
	broker   *broker.RabbitMQBroker
	cfg      *config.Config
	notifier *notifier.Notifier
}

func New(storage *postgres.Storage, broker *broker.RabbitMQBroker, cfg *config.Config, notifier *notifier.Notifier) *Service {
	return &Service{
		storage:  storage,
		broker:   broker,
		cfg:      cfg,
		notifier: notifier,
	}
}

func (s *Service) CreateNotification(recipientID int64, dateStr, text string) (int64, error) {
	notificationID, err := s.storage.CreateNotification(recipientID, dateStr, text)
	if err != nil {
		return 0, fmt.Errorf("service failed to create notification: %w", err)
	}

	err = s.publishNotificationID(notificationID)
	if err != nil {
		return notificationID, fmt.Errorf("service failed to publish notification ID: %w", err)
	}

	return notificationID, nil
}

func (s *Service) GetNotificationStatus(notificationID int64) (string, error) {
	return s.storage.GetNotificationStatus(notificationID)
}

func (s *Service) GetNotificationByID(notificationID int64) (*models.Notification, error) {
	return s.storage.GetNotificationByID(notificationID)
}

func (s *Service) UpdateNotificationStatus(notificationID int64, status string) error {
	return s.storage.UpdateNotificationStatus(notificationID, status)
}

func (s *Service) DeleteNotification(notificationID int64) error {
	return s.storage.DeleteNotification(notificationID)
}

func (s *Service) publishNotificationID(id int64) error {
	message := []byte(strconv.FormatInt(id, 10))
	return s.broker.Publish(s.cfg.Rabbit.QueueName, message)
}

func (s *Service) SendNotification(recipientID int64, text string) error {
	return s.notifier.SendNotification(recipientID, text)
}
