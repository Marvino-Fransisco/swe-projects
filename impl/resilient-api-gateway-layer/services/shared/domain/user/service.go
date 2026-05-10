package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// UserService provides domain logic for user operations.
// It handles registration, authentication, and profile management.
// Token generation/validation is handled outside this service.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new UserService with the given repositories.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Register creates a new user account.
// It validates the email and name, hashes the password, and persists the user.
// Returns the created user.
func (s *UserService) Register(ctx context.Context, email, name, password string) (*User, error) {
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	nameVO, err := NewFullName(name)
	if err != nil {
		return nil, err
	}

	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	existing, err := s.repo.FindByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		FullName:     nameVO,
		Email:        emailVO,
		PasswordHash: string(hash),
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate validates email and password credentials.
// Returns the authenticated user on success.
// Token generation is handled by the caller.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*User, error) {
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// GetProfile retrieves a user's profile by ID.
// Attempts cache first; on cache miss, queries the database and updates the cache.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*User, error) {
	// Cache miss — query database.
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateProfile updates a user's name, email, address, and phone.
// Returns the updated user.
func (s *UserService) UpdateProfile(ctx context.Context, userID string, name, email, address, phone string) (*User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	nameVO, err := NewFullName(name)
	if err != nil {
		return nil, err
	}

	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Check if the new email is already taken by another user.
	if emailVO != user.Email {
		existing, err := s.repo.FindByEmail(ctx, emailVO)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != user.ID {
			return nil, errors.New("email already in use")
		}
	}

	user.FullName = nameVO
	user.Email = emailVO
	user.Address = address

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword verifies the old password and sets the new password.
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.repo.Update(ctx, user)
}
