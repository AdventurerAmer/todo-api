package domain

import "time"

type TokenScope string

const (
	TokenScopeActivation = "activation"
)

type Token struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userID"`
	ExpiresAt time.Time  `json:"expiresAt"`
	Scope     TokenScope `json:"scope"`
}
