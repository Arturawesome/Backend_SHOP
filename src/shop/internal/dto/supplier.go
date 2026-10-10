package dto

import "github.com/google/uuid"

type SupplierRequest struct {
	Name        string         `json:"name" example:"Bosch"`
	PhoneNumber string         `json:"phone_number" example:"+74951234567"`
	Address     AddressRequest `json:"address"`
}

type SupplierResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"Bosch"`
	AddressID   uuid.UUID `json:"address_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	PhoneNumber string    `json:"phone_number" example:"+74951234567"`
}
