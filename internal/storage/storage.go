package storage

import "errors"

var (
	ErrNotifyNotFound = errors.New("notification not found")
	ErrNotifyExists   = errors.New("notification already exists")
)
