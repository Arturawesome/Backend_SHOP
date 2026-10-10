package model

import "github.com/google/uuid"

type Supplier struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	AddressID   uuid.UUID `db:"address_id"`
	PhoneNumber string    `db:"phone_number"`
}
