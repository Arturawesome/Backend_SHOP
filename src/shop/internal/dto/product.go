package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProductRequest struct {
	Name           string    `json:"name" example:"Washing machine"`
	Category       string    `json:"category" example:"washing_machine"`
	Price          int64     `json:"price" example:"59990"`
	SupplierID     uuid.UUID `json:"supplier_id"`
	AvailableStock int       `json:"available_stock" example:"15"`
}

type ProductResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name           string     `json:"name"  example:"Washing machine"`
	Category       string     `json:"category" example:"washing_machine"`
	Price          int64      `json:"price" example:"59990"`
	AvailableStock int        `json:"available_stock" example:"15"`                               // число закупленных экземпляров товара
	LastUpdateDate time.Time  `json:"last_update_date" example:"1998-05-12T00:00:00Z"`            // дата последней закупки
	SupplierID     uuid.UUID  `json:"supplier_id" example:"550e8400-e29b-41d4-a716-446655440000"` //
	ImageID        *uuid.UUID `json:"image_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}
