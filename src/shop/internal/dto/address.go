package dto

import "github.com/google/uuid"

type AddressRequest struct {
	Country string `json:"country" example:"Russia"`
	City    string `json:"city" example:"Moscow"`
	Street  string `json:"street" example:"Baumanskaya street"`
}

type AddressResponce struct {
	ID      uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Country string    `json:"country" example:"Russia"`
	City    string    `json:"city" example:"Moscow"`
	Street  string    `json:"street" example:"Baumanskaya street"`
}
