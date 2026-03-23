package main

import "time"

type task struct {
	ID          int       `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      string    `json:"user_id"`
	Content     string    `json:"content"`
	IsCompleted bool      `json:"is_completed"`
	Version     int       `json:"-"`
}
