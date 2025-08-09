package postgres

import (
	"DelayedNotifier/internal/config"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db  *sql.DB
	rdb *redis.Client
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

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  cfg.Redis.PoolTimeout,
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to Redis: %v", err)
	}

	return &Storage{
		db:  db,
		rdb: rdb,
	}, nil
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

	err = s.rdb.Set(fmt.Sprintf("notification:%d", notificationId), "pending", 48*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to set Redis key: %v", err)
	}

	return notificationId, nil
}

func (s *Storage) GetNotificationStatus(notificationID int64) (string, error) {
	status, err := s.rdb.Get(fmt.Sprintf("notification:%d", notificationID)).Result()

	if err == nil {
		return status, nil
	}

	if !errors.Is(err, redis.Nil) {
		log.Printf("Redis error: %v", err)
	}

	err = s.db.QueryRow(
		`SELECT status FROM notifications WHERE id = $1`,
		notificationID,
	).Scan(&status)

	if err != nil {
		return "", fmt.Errorf("failed to get notification status: %v", err)
	}

	return status, nil
}

func (s *Storage) DeleteNotification(notificationID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM notifications WHERE id = $1`,
		notificationID)

	if err != nil {
		return fmt.Errorf("failed to delete notification: %v", err)
	}

	key := fmt.Sprintf("notification:%d", notificationID)
	err = s.rdb.Del(key).Err()
	if err != nil {
		log.Printf("Failed to delete notification from Redis: %v", err)
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

	s.rdb.Set(fmt.Sprintf("notification:%d", notificationID), status, 48*time.Hour)

	return nil
}

func (s *Storage) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	return nil
}
