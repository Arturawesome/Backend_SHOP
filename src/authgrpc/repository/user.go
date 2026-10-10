package repository

import (
	"BE2_API_Shop/authgrpc/model"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) error
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByID(ctx context.Context, id string) (model.User, error)
	ChangePassword(ctx context.Context, userID string, passwordHash string) error
}
