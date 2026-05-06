package user

import (
	"context"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// Save persists a new user to the data store.
	Save(ctx context.Context, u *User) error

	// FindByEmail retrieves a user by their email address.
	// Returns nil if no user is found.
	FindByEmail(ctx context.Context, email Email) (*User, error)

	// FindByID retrieves a user by their ID.
	// Returns nil if no user is found.
	FindByID(ctx context.Context, id string) (*User, error)

	// Update persists changes to an existing user.
	Update(ctx context.Context, u *User) error
}
