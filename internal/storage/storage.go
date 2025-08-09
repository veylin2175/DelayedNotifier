package storage

import "errors"

var (
	ErrNoteNotFound = errors.New("notification not found")
	ErrNoteExists   = errors.New("notification already exists")
)
