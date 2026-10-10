package service

import (
	"BE2_API_Shop/authgrpc/repository"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidData):
		return status.Error(
			codes.InvalidArgument,
			"invalid input data",
		)

	case errors.Is(err, ErrInvalidCredentials):
		return status.Error(
			codes.Unauthenticated,
			"invalid email or password",
		)

	case errors.Is(err, ErrInvalidToken):
		return status.Error(
			codes.Unauthenticated,
			"invalid or expired token",
		)

	case errors.Is(err, repository.ErrNotFound):
		return status.Error(
			codes.NotFound,
			"user not found",
		)

	case errors.Is(err, repository.ErrConflict):
		return status.Error(
			codes.AlreadyExists,
			"user already exists",
		)

	default:
		return status.Error(
			codes.Internal,
			"internal server error",
		)
	}
}
