package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID             uuid.UUID  `db:"id"`
	Name           string     `db:"name"`
	Category       string     `db:"category"`
	Price          int64      `db:"price"`
	AvailableStock int        `db:"available_stock"`  // число закупленных экземпляров товара
	LastUpdateDate time.Time  `db:"last_update_date"` // дата последней закупки
	SupplierID     uuid.UUID  `db:"supplier_id"`      //
	ImageID        *uuid.UUID `db:"image_id"`
}

type ProductUpdateEvent struct {
	ID             uuid.UUID `json:"id"`
	Price          int64     `json:"price"`
	AvailableStock int       `json:"available_stock"` // число закупленных экземпляров товара
	Timestamp      time.Time `json:"timestamp"`
}
