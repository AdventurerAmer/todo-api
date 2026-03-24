package domain

import "time"

type Task struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ListID      string    `json:"list_id"`
	Content     string    `json:"content"`
	IsCompleted bool      `json:"is_completed"`
	Version     int       `json:"version"`
}
