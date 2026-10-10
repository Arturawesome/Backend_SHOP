package repository

import (
	"BE2_API_Shop/shop/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type ImageRepository interface {
	AddImage(ctx context.Context, image model.Image, productID uuid.UUID) error
	ChangeImage(ctx context.Context, imageID uuid.UUID, newImage []byte) error
	RemoveByID(ctx context.Context, ID uuid.UUID) error
	GetByProductID(ctx context.Context, productID uuid.UUID) (model.Image, error)
	GetByID(ctx context.Context, ID uuid.UUID) (model.Image, error)
}

type ImageCashe interface {
	GetByID(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
