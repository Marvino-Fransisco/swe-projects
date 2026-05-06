package user

import "context"

// UserCacheRepository defines the interface for user cache operations.
type UserCacheRepository interface {
	// GetByID retrieves a user from cache by their ID.
	// Returns nil if not found in cache.
	GetByID(ctx context.Context, id string) (*User, error)

	// Set persists a user to cache.
	Set(ctx context.Context, u *User) error

	// Delete removes a user from cache by their ID.
	Delete(ctx context.Context, id string) error
}
