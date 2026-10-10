package postgres

// user.go
import (
	"BE2_API_Shop/authgrpc/model"
	"BE2_API_Shop/authgrpc/repository"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

var _ (repository.UserRepository) = (*PostgresUserRepository)(nil)

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}
func (repo *PostgresUserRepository) ChangePassword(
	ctx context.Context,
	userID string,
	passwordHash string,
) error {
	query := `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
	`

	commandTag, err := repo.db.Exec(
		ctx,
		query,
		passwordHash,
		userID,
	)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
func (repo *PostgresUserRepository) Create(
	ctx context.Context, user model.User) error {
	query := `
    INSERT INTO users (
        id,
        email,
        name,
    	surname,
    	phone,
    	password_hash 
    )
    VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := repo.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Surname,
		user.Phone,
		user.PasswordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrConflict
		}

		return err
	}

	return nil
}

func (repo *PostgresUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (model.User, error) {
	query := `
    SELECT     
	id,
    email,
    name,
    surname,
    phone,
    password_hash
	 FROM users 
	WHERE email = $1
	`
	var user model.User
	err := repo.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Surname,
		&user.Phone,
		&user.PasswordHash,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, repository.ErrNotFound
	}
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (repo *PostgresUserRepository) GetByID(
	ctx context.Context,
	id string,
) (model.User, error) {
	query := `
    SELECT
        id,
        email,
        name,
        surname,
        phone,
        password_hash
    FROM users
    WHERE id = $1
`

	var user model.User

	err := repo.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Surname,
		&user.Phone,
		&user.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, repository.ErrNotFound
	}

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
