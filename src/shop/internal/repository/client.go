package repository

import (
	"BE2_API_Shop/shop/internal/model"
	"context"

	"github.com/google/uuid"
)

// check
type ClientRepository interface {
	AddClient(ctx context.Context, client model.Client, address model.Address) error
	RemoveByID(ctx context.Context, ID uuid.UUID) error

	GetByNameSurname(ctx context.Context, name string,
		surname string) ([]model.Client, error)

	GetAllClients(ctx context.Context, limit *int,
		offset *int) ([]model.Client, error)

	PatchClientAddress(ctx context.Context, clientID uuid.UUID,
		newAddress model.Address) error
}
