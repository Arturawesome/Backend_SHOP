package dto

import (
	"time"

	"github.com/google/uuid"
)

type ClientRequest struct {
	Name     string         `json:"name" example:"Artur"`
	Surname  string         `json:"surname" example:"Nasyrov"`
	Birthday time.Time      `json:"birthday" example:"1998-05-12T00:00:00Z"`
	Gender   string         `json:"gender" example:"male"`
	Address  AddressRequest `json:"address"`
}

type ClientResponse struct {
	ID               uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name             string    `json:"name"  example:"Artur"`
	Surname          string    `json:"surname"  example:"Nasyrov"`
	Birthday         time.Time `json:"birthday" example:"1998-05-12T00:00:00Z"`
	Gender           string    `json:"gender" example:"male"`
	RegistrationDate time.Time `json:"registration_date" example:"1998-05-12T00:00:00Z"`
}
