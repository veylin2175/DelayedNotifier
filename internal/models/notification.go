package models

import "time"

type Notification struct {
	ID          int64     `json:"id"`
	RecipientID int64     `json:"recipient_id"`
	Date        time.Time `json:"date"`
	Text        string    `json:"text"`
	Status      string    `json:"status"`
}
