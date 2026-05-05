package profile

import (
	"context"

	"web-backend/apperror"

	"shared/domain/user"
)

type profileUseCase struct {
	userSvc *user.UserService
}

func NewProfileUseCase(userSvc *user.UserService) ProfileUseCase {
	return &profileUseCase{userSvc: userSvc}
}

func (uc *profileUseCase) GetProfile(ctx context.Context, req GetProfileRequest) (*ProfileResponse, error) {
	u, err := uc.userSvc.GetProfile(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewNotFound(err.Error())
	}

	return mapUserToProfileResponse(u), nil
}

func (uc *profileUseCase) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*ProfileResponse, error) {
	u, err := uc.userSvc.UpdateProfile(ctx, req.UserID, req.Name, req.Email, req.Address, "")
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return mapUserToProfileResponse(u), nil
}

func (uc *profileUseCase) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	if err := uc.userSvc.ChangePassword(ctx, req.UserID, req.OldPassword, req.NewPassword); err != nil {
		return apperror.NewBadRequest(err.Error())
	}
	return nil
}

func mapUserToProfileResponse(u *user.User) *ProfileResponse {
	return &ProfileResponse{
		ID:       u.ID,
		FullName: u.FullName.String(),
		Email:    u.Email.String(),
		Address:  u.Address,
	}
}
