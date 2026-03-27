package domain

import "time"

type User struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdateAt     time.Time `json:"updated_at"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	IsActivated  bool      `json:"is_activated"`
	Version      int       `json:"-"`
}
