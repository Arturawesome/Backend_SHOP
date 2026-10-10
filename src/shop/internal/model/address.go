package model

import "github.com/google/uuid"

type Address struct {
	ID      uuid.UUID `db:"id"`
	Country string    `db:"country"`
	City    string    `db:"city"`
	Street  string    `db:"street"`
}
