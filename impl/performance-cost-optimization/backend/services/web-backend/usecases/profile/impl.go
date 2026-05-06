package profile

import (
	"context"
	"log"

	"web-backend/apperror"

	"shared/config"
	"shared/domain/user"
)

type profileUseCase struct {
	userSvc       *user.UserService
	userCacheRepo user.UserCacheRepository
	dbTx          config.DBTransaction
	redisTx       config.RedisTransaction
}

func NewProfileUseCase(userSvc *user.UserService, userCacheRepo user.UserCacheRepository, dbTx config.DBTransaction, redisTx config.RedisTransaction) ProfileUseCase {
	return &profileUseCase{
		userSvc:       userSvc,
		userCacheRepo: userCacheRepo,
		dbTx:          dbTx,
		redisTx:       redisTx,
	}
}

func (uc *profileUseCase) GetProfile(ctx context.Context, req GetProfileRequest) (*ProfileResponse, error) {
	u, err := uc.userSvc.GetProfile(ctx, req.UserID)
	if err != nil {
		return nil, apperror.NewNotFound(err.Error())
	}

	return mapUserToProfileResponse(u), nil
}

func (uc *profileUseCase) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*ProfileResponse, error) {
	var u *user.User

	err := uc.dbTx(ctx, func(txCtx context.Context) error {
		var err error
		u, err = uc.userSvc.UpdateProfile(ctx, req.UserID, req.Name, req.Email, req.Address, "")
		if err != nil {
			return apperror.NewBadRequest(err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = uc.redisTx(ctx, func(txCtx context.Context) error {
		var err error

		err = uc.userCacheRepo.Delete(txCtx, u.ID)
		if err != nil {
			return err
		}

		err = uc.userCacheRepo.Set(txCtx, u)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("cache error: %v", err)
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
