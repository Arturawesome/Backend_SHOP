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

var _ repository.ImageRepository = (*ImageRepository)(nil)

type ImageRepository struct {
	db *pgxpool.Pool
}

func NewImageRepository(db *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) AddImage(
	ctx context.Context,
	image model.Image,
	productID uuid.UUID,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oldImageID *uuid.UUID

	err = tx.QueryRow(
		ctx,
		`
		SELECT image_id
		FROM product
		WHERE id = $1
		FOR UPDATE
		`,
		productID,
	).Scan(&oldImageID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}

		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO image (id, image)
		VALUES ($1, $2)
		`,
		image.ID,
		image.Image,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE product
		SET image_id = $1
		WHERE id = $2
		`,
		image.ID,
		productID,
	)
	if err != nil {
		return err
	}

	if oldImageID != nil {
		_, err = tx.Exec(
			ctx,
			`
			DELETE FROM image
			WHERE id = $1
			`,
			*oldImageID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ImageRepository) ChangeImage(
	ctx context.Context,
	imageID uuid.UUID,
	newImage []byte,
) error {
	query := `
		UPDATE image
		SET image = $1
		WHERE id = $2
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		newImage,
		imageID,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *ImageRepository) RemoveByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		UPDATE product
		SET image_id = NULL
		WHERE image_id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	deleteTag, err := tx.Exec(
		ctx,
		`
		DELETE FROM image
		WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	if deleteTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *ImageRepository) GetByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (model.Image, error) {
	query := `
		SELECT
			i.id,
			i.image
		FROM image AS i
		JOIN product AS p
			ON p.image_id = i.id
		WHERE p.id = $1
	`

	rows, err := r.db.Query(
		ctx,
		query,
		productID,
	)
	if err != nil {
		return model.Image{}, err
	}
	defer rows.Close()

	image, err := pgx.CollectOneRow(
		rows,
		pgx.RowToStructByName[model.Image],
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Image{}, repository.ErrNotFound
		}

		return model.Image{}, err
	}

	return image, nil
}

func (r *ImageRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (model.Image, error) {
	query := `
		SELECT
			id,
			image
		FROM image
		WHERE id = $1
	`

	rows, err := r.db.Query(
		ctx,
		query,
		id,
	)
	if err != nil {
		return model.Image{}, err
	}
	defer rows.Close()

	image, err := pgx.CollectOneRow(
		rows,
		pgx.RowToStructByName[model.Image],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Image{}, repository.ErrNotFound
		}

		return model.Image{}, err
	}

	return image, nil
}
