package auth

import (
	"context"
	"fmt"

	"web-backend/apperror"

	"shared/domain/user"
	"shared/util"
)

type authUseCase struct {
	userSvc *user.UserService
	jwtSvc  *util.JWTService
}

func NewAuthUseCase(userSvc *user.UserService, jwtSvc *util.JWTService) AuthUseCase {
	return &authUseCase{
		userSvc: userSvc,
		jwtSvc:  jwtSvc,
	}
}

func (uc *authUseCase) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	u, err := uc.userSvc.Register(ctx, req.Email, req.Name, req.Password)
	if err != nil {
		return nil, apperror.NewBadRequest(err.Error())
	}

	return &RegisterResponse{
		ID:    u.ID,
		Email: u.Email.String(),
		Name:  u.FullName.String(),
	}, nil
}

func (uc *authUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := uc.userSvc.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, apperror.NewUnauthorized(err.Error())
	}

	accessToken, refreshToken, err := uc.generateTokenPair(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *authUseCase) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	newAccess, newRefresh, err := uc.jwtSvc.RefreshTokens(req.RefreshToken)
	if err != nil {
		return nil, apperror.NewUnauthorized("invalid or expired refresh token")
	}

	return &RefreshResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}

func (uc *authUseCase) generateTokenPair(u *user.User) (string, string, error) {
	role := "user"
	accessToken, err := uc.jwtSvc.GenerateAccessToken(u.ID, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := uc.jwtSvc.GenerateRefreshToken(u.ID, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
