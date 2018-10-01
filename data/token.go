package data

import (
	"time"
)

// Token represents an auth token
type Token struct {
	ID       string    `json:"token"`
	UserName string    `json:"user"`
	Created  time.Time `json:"created"`
	Expires  time.Time `json:"expires"`
}
