package model

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Surname      string    `json:"surname"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"password_hash"`
}
