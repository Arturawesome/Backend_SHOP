package repository

import (
	"BE2_API_Shop/shop/internal/model"
	"context"

	"github.com/google/uuid"
)

// check
type ProductRepository interface {
	AddProduct(ctx context.Context, product model.Product) error
	DecreaseProduct(ctx context.Context, ID uuid.UUID, num int) error
	GetByID(ctx context.Context, ID uuid.UUID) (model.Product, error)
	GetAvailableProducts(ctx context.Context) ([]model.Product, error)
	RemoveByID(ctx context.Context, ID uuid.UUID) error
	UpdateProduct(ctx context.Context, event *model.ProductUpdateEvent) error
}
