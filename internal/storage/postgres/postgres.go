package postgres

import (
	"DelayedNotifier/internal/config"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func InitDB(cfg *config.Config) (*Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("couldn't connect to the DB: %v", err)
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateNotification(recipientID int64, dateStr, text string) (int64, error) {
	date, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %v", err)
	}

	var notificationId int64
	err = s.db.QueryRow(
		`INSERT INTO notifications (recipient_id, date, text) VALUES ($1, $2, $3) RETURNING id`,
		recipientID, date, text,
	).Scan(&notificationId)

	if err != nil {
		return 0, fmt.Errorf("failed to create notification: %v", err)
	}

	return notificationId, nil
}

func (s *Storage) GetNotificationStatus(notificationID int64) (string, error) {
	var notificationStatus string

	err := s.db.QueryRow(
		`SELECT status FROM notifications WHERE id = $1`,
		notificationID,
	).Scan(&notificationStatus)

	if err != nil {
		return "", fmt.Errorf("failed to get notification status: %v", err)
	}

	return notificationStatus, nil
}

func (s *Storage) DeleteNotificationStatus(notificationID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM notifications WHERE id = $1`,
		notificationID)

	if err != nil {
		return fmt.Errorf("failed to delete notification: %v", err)
	}

	return nil
}

func (s *Storage) UpdateNotificationStatus(notificationID int64, status string) error {
	_, err := s.db.Exec(
		`UPDATE notifications SET status = $1 WHERE id = $2`,
		status, notificationID)

	if err != nil {
		return fmt.Errorf("failed to update notification status: %v", err)
	}

	return nil
}

func (s *Storage) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	return nil
}
