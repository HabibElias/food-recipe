package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Username     string
	FirstName    string
	LastName     string
	PasswordHash string
	AvatarURL    string
	Bio          string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.FirstName + u.LastName
}
