package repository

import (
	"BE2_API_Shop/shop/internal/model"
	"context"

	"github.com/google/uuid"
)

//check

type SupplierRepository interface {
	AddSupplier(ctx context.Context, supplier model.Supplier, address model.Address) error
	PatchAddress(ctx context.Context, ID uuid.UUID, newAddress model.Address) error
	RemoveByID(ctx context.Context, ID uuid.UUID) error
	GetAllSuppliers(ctx context.Context) ([]model.Supplier, error)
	GetByID(ctx context.Context, ID uuid.UUID) (model.Supplier, error)
}
