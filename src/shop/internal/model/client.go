package model

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID               uuid.UUID `db:"id"`
	Name             string    `db:"client_name"`
	Surname          string    `db:"client_surname"`
	Birthday         time.Time `db:"birthday"`
	Gender           string    `db:"gender"`
	RegistrationDate time.Time `db:"registration_date"`
	AddressID        uuid.UUID `db:"address_id"`
}
