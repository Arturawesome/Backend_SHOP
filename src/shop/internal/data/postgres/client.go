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

var _ repository.ClientRepository = (*ClientRepository)(nil)

type ClientRepository struct {
	db *pgxpool.Pool
}

func NewClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) AddClient(
	ctx context.Context,
	client model.Client,
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
		INSERT INTO client (
			id,
			client_name,
			client_surname,
			birthday,
			gender,
			registration_date,
			address_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		client.ID,
		client.Name,
		client.Surname,
		client.Birthday,
		client.Gender,
		client.RegistrationDate,
		client.AddressID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ClientRepository) RemoveByID(
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
		FROM client
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
		DELETE FROM client
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
func (r *ClientRepository) GetByNameSurname(
	ctx context.Context,
	name string,
	surname string,
) ([]model.Client, error) {

	query := `
		SELECT 
			id,
			client_name,
			client_surname,
			birthday,
			gender,
			registration_date,
			address_id
		FROM client
		WHERE client_name = $1
		  AND client_surname = $2
	`

	rows, err := r.db.Query(ctx, query, name, surname)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[model.Client],
	)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *ClientRepository) GetAllClients(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]model.Client, error) {

	query := `
		SELECT
			id,
			client_name,
			client_surname,
			birthday,
			gender,
			registration_date,
			address_id
		FROM client
	`

	var args []any

	if limit != nil {
		args = append(args, *limit)
		query += " LIMIT $1"

		if offset != nil {
			args = append(args, *offset)
			query += " OFFSET $2"
		}
	} else if offset != nil {
		args = append(args, *offset)
		query += " OFFSET $1"
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[model.Client],
	)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *ClientRepository) PatchClientAddress(
	ctx context.Context,
	clientID uuid.UUID,
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
			FROM client
			WHERE id = $4
		)
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		newAddress.Country,
		newAddress.City,
		newAddress.Street,
		clientID,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
