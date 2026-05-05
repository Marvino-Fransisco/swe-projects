package profile

import "context"

type ProfileUseCase interface {
	GetProfile(ctx context.Context, req GetProfileRequest) (*ProfileResponse, error)
	UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*ProfileResponse, error)
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error
}
