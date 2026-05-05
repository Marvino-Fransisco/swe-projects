package repository

import (
	"context"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/user"
)

// userRepository implements user.UserRepository using GORM.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &userRepository{db: db}
}

// Save persists a new user to the database.
func (r *userRepository) Save(ctx context.Context, u *user.User) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(u).Error
}

// FindByEmail retrieves a user by their email address.
// Returns nil if no user is found.
func (r *userRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	db := config.DBFromContext(ctx, r.db)
	var u user.User
	err := db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// FindByID retrieves a user by their ID.
// Returns nil if no user is found.
func (r *userRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	db := config.DBFromContext(ctx, r.db)
	var u user.User
	err := db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Update persists changes to an existing user.
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(u).Error
}
