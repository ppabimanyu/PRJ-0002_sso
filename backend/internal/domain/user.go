package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is one person's global identity in the Instance.
type User struct {
	ID              uuid.UUID
	Email           string
	Name            string
	PictureURL      string
	IsAdmin         bool
	EmailVerifiedAt *time.Time
	DisabledAt      *time.Time
	CreatedAt       time.Time
}

// Active reports whether the User may authenticate and hold a Session.
func (u User) Active() bool {
	return u.DisabledAt == nil
}

// DisplayName returns the name when set, otherwise the email local part.
func (u User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	for index, character := range u.Email {
		if character == '@' {
			return u.Email[:index]
		}
	}
	return u.Email
}
