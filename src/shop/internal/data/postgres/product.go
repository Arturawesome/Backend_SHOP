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

var _ repository.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) AddProduct(
	ctx context.Context,
	product model.Product,
) error {
	query := `
		INSERT INTO product (
			id,
			name,
			category,
			price,
			available_stock,
			last_update_date,
			supplier_id,
			image_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		product.ID,
		product.Name,
		product.Category,
		product.Price,
		product.AvailableStock,
		product.LastUpdateDate,
		product.SupplierID,
		product.ImageID,
	)

	return err
}

func (r *ProductRepository) DecreaseProduct(
	ctx context.Context,
	id uuid.UUID,
	num int,
) error {
	query := `
	UPDATE product
	SET available_stock = available_stock - $1
	WHERE id = $2
	  AND available_stock >= $1
`

	tag, err := r.db.Exec(
		ctx,
		query,
		num,
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

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.Product, error) {

	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			name,
			category,
			price,
			available_stock,
			last_update_date,
			supplier_id,
			image_id
		FROM product
		WHERE id = $1`,
		id,
	)

	if err != nil {
		return model.Product{}, err
	}
	defer rows.Close()

	product, err := pgx.CollectOneRow(
		rows,
		pgx.RowToStructByName[model.Product],
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, repository.ErrNotFound
		}

		return model.Product{}, err
	}

	return product, nil
}

func (r *ProductRepository) GetAvailableProducts(
	ctx context.Context,
) ([]model.Product, error) {
	query := `
		SELECT
			id,
			name,
			category,
			price,
			available_stock,
			last_update_date,
			supplier_id,
			image_id
		FROM product
		WHERE available_stock > 0
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[model.Product],
	)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) RemoveByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var imageID *uuid.UUID

	err = tx.QueryRow(
		ctx,
		`
		SELECT image_id
		FROM product
		WHERE id = $1
		FOR UPDATE
		`,
		id,
	).Scan(&imageID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}

		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		DELETE FROM product
		WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	if imageID != nil {
		_, err = tx.Exec(
			ctx,
			`
			DELETE FROM image
			WHERE id = $1
			`,
			*imageID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ProductRepository) GetProductIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id
		 FROM product`,
	)
	if err != nil {
		return nil, err
	}

	listID, err := pgx.CollectRows(
		rows,
		pgx.RowTo[uuid.UUID],
	)
	if err != nil {
		return nil, err
	}

	return listID, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context,
	event *model.ProductUpdateEvent) error {

	tag, err := r.db.Exec(
		ctx,
		`
       UPDATE product
		SET
			price = $1,
			available_stock = $2,
			last_update_date = $3
		WHERE id = $4
        `,
		event.Price,
		event.AvailableStock,
		event.Timestamp,
		event.ID,
	)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
