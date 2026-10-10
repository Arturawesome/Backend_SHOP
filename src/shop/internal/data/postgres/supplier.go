package postgres

import (
	"BE2_API_Shop/shop/internal/model"
	"BE2_API_Shop/shop/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.SupplierRepository = (*SupplierRepository)(nil)

type SupplierRepository struct {
	db *pgxpool.Pool
}

func NewSupplierRepository(db *pgxpool.Pool) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) AddSupplier(
	ctx context.Context,
	supplier model.Supplier,
	address model.Address,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO address (id, country, city, street)
		VALUES ($1, $2, $3, $4)
		`,
		address.ID,
		address.Country,
		address.City,
		address.Street,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO supplier (
			id,
			name,
			address_id,
			phone_number
		)
		VALUES ($1, $2, $3, $4)
		`,
		supplier.ID,
		supplier.Name,
		supplier.AddressID,
		supplier.PhoneNumber,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *SupplierRepository) PatchAddress(
	ctx context.Context,
	id uuid.UUID,
	newAddress model.Address,
) error {
	query := `
		UPDATE address
		SET
			country = $1,
			city = $2,
			street = $3
		WHERE id = (
			SELECT address_id
			FROM supplier
			WHERE id = $4
		)
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		newAddress.Country,
		newAddress.City,
		newAddress.Street,
		id,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *SupplierRepository) RemoveByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var addressID uuid.UUID

	err = tx.QueryRow(
		ctx,
		`
		SELECT address_id
		FROM supplier
		WHERE id = $1
		`,
		id,
	).Scan(&addressID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}

		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM supplier
		WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM address
		WHERE id = $1
		`,
		addressID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *SupplierRepository) GetAllSuppliers(
	ctx context.Context,
) ([]model.Supplier, error) {
	query := `
		SELECT
			id,
			name,
			address_id,
			phone_number
		FROM supplier
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[model.Supplier],
	)
	if err != nil {
		return nil, err
	}

	return suppliers, nil
}

func (r *SupplierRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.Supplier, error) {
	query := `
		SELECT
			id,
			name,
			address_id,
			phone_number
		FROM supplier
		WHERE id = $1
	`

	rows, err := r.db.Query(
		ctx,
		query,
		id,
	)
	if err != nil {
		return model.Supplier{}, err
	}
	defer rows.Close()

	supplier, err := pgx.CollectOneRow(
		rows,
		pgx.RowToStructByName[model.Supplier],
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supplier{}, repository.ErrNotFound
		}

		return model.Supplier{}, err
	}

	return supplier, nil
}
