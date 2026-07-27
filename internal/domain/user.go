package domain

import "time"

type User struct {
	ID           int64
	PublicID     string
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
