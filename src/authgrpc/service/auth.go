package service

import (
	pbAuth "BE2_API_Shop/gen/auth"
	"context"
	"fmt"
)

type AuthServer struct {
	pbAuth.UnimplementedAuthServiceServer
	userService *UserService
	jwtProvider *JWTProvider
}

func NewAuthServer(userService *UserService, jwtProvider *JWTProvider,
) *AuthServer {
	return &AuthServer{
		userService: userService,
		jwtProvider: jwtProvider,
	}
}

func (auth *AuthServer) Registration(ctx context.Context,
	req *pbAuth.RegistrationRequest) (*pbAuth.RegistrationResponse, error) {
	input := CreateUserInput{
		Email:    req.Email,
		Name:     req.Name,
		Surname:  req.Surname,
		Phone:    req.Phone,
		Password: req.Password,
	}

	user, err := auth.userService.Create(ctx, input)
	if err != nil {
		return nil, toGRPCError(err)
	}
	token, err := auth.jwtProvider.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pbAuth.RegistrationResponse{AccessToken: token}, nil
}

func (auth *AuthServer) Login(ctx context.Context,
	req *pbAuth.LoginRequest,
) (*pbAuth.LoginResponse, error) {

	user, err := auth.userService.Authenticate(
		ctx,
		req.Email,
		req.Password,
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	token, err := auth.jwtProvider.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pbAuth.LoginResponse{
		AccessToken: token,
	}, nil
}

func (auth *AuthServer) RestorePassword(
	ctx context.Context,
	req *pbAuth.RestorePasswordRequest,
) (*pbAuth.RestorePasswordResponse, error) {

	if req.Email == "" {
		return nil, toGRPCError(ErrInvalidData)
	}

	user, err := auth.userService.GetByEmail(
		ctx,
		req.Email,
	)
	if err != nil {
		return nil, toGRPCError(err)
	}


	newPassword, err := generatePassword(12)
	if err != nil {
		return nil, toGRPCError(err)
	}

	
	err = auth.userService.ChangePassword(
		ctx,
		user.ID.String(),
		newPassword,
	)
	if err != nil {
		return nil, toGRPCError(err)
	}

	
	fmt.Printf(
		"Password reset for %s. New password: %s\n",
		user.Email,
		newPassword,
	)

	return &pbAuth.RestorePasswordResponse{
		IsRestored: true,
	}, nil
}
func (auth *AuthServer) ChangePassword(ctx context.Context,
	req *pbAuth.ChangePasswordRequest) (*pbAuth.ChangePasswordResponse, error) {
	userID, err := auth.jwtProvider.ValidateTokenAndGetUserID(req.AccessToken)
	if err != nil {
		return nil, toGRPCError(err)
	}
	err = auth.userService.ChangePassword(ctx, userID, req.NewPassword)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pbAuth.ChangePasswordResponse{
		IsChanged: true,
	}, nil
}

func (auth *AuthServer) ValidateToken(
	ctx context.Context,
	req *pbAuth.ValidateTokenRequest,
) (*pbAuth.ValidateTokenResponse, error) {

	userID, err := auth.jwtProvider.ValidateTokenAndGetUserID(req.Token)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pbAuth.ValidateTokenResponse{
		IsValid: true,
		UserId:  userID,
	}, nil
}
