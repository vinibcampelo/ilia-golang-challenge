package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"ilia-golang-challenge/ms-users/internal/domain/user"
)

const postgresUniqueViolation = "23505"

// PostgresUserRepository persists users in PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository returns a repository backed by Postgres (pgx stdlib driver).
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, u *user.User) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (
			id,
			first_name,
			last_name,
			email,
			password
		)
		VALUES ($1, $2, $3, $4, $5)
	`, u.ID, u.FirstName, u.LastName, u.Email, u.Password)
	if err != nil {
		if isPostgresUniqueViolation(err) {
			return fmt.Errorf("save user: %w", user.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func isPostgresUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation
}
