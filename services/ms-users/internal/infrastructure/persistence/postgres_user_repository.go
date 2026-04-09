package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"ilia-golang-challenge/services/ms-users/internal/application/user/usecase"
	domainuser "ilia-golang-challenge/services/ms-users/internal/domain/user"
)

const postgresUniqueViolationSQLState = "23505"

type PostgresUserRepository struct {
	database *sql.DB
}

func NewPostgresUserRepository(database *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{database: database}
}

func (r *PostgresUserRepository) Save(ctx context.Context, userEntity *domainuser.User) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	_, err := r.database.ExecContext(ctx, `
		INSERT INTO users (
			id,
			first_name,
			last_name,
			email,
			password
		)
		VALUES ($1, $2, $3, $4, $5)
	`, userEntity.ID, userEntity.FirstName, userEntity.LastName, userEntity.Email, userEntity.Password)
	if err != nil {
		if isPostgresUniqueViolationError(err) {
			return fmt.Errorf("save user: %w", usecase.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, userID string) (*domainuser.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	row := r.database.QueryRowContext(ctx, `
		SELECT id, first_name, last_name, email, password
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)
	var foundUser domainuser.User
	if err := row.Scan(&foundUser.ID, &foundUser.FirstName, &foundUser.LastName, &foundUser.Email, &foundUser.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("find user: %w", usecase.ErrNotFound)
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &foundUser, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	row := r.database.QueryRowContext(ctx, `
		SELECT id, first_name, last_name, email, password
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email)
	var foundUser domainuser.User
	if err := row.Scan(&foundUser.ID, &foundUser.FirstName, &foundUser.LastName, &foundUser.Email, &foundUser.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("find user by email: %w", usecase.ErrNotFound)
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &foundUser, nil
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]domainuser.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	rows, err := r.database.QueryContext(ctx, `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domainuser.User
	for rows.Next() {
		var rowUser domainuser.User
		if err := rows.Scan(&rowUser.ID, &rowUser.FirstName, &rowUser.LastName, &rowUser.Email); err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		users = append(users, rowUser)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, userEntity *domainuser.User) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	execResult, err := r.database.ExecContext(ctx, `
		UPDATE users
		SET first_name = $1,
		    last_name = $2,
		    email = $3,
		    password = $4
		WHERE id = $5 AND deleted_at IS NULL
	`, userEntity.FirstName, userEntity.LastName, userEntity.Email, userEntity.Password, userEntity.ID)
	if err != nil {
		if isPostgresUniqueViolationError(err) {
			return fmt.Errorf("update user: %w", usecase.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("update user: %w", err)
	}
	rowsAffected, err := execResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("update user: %w", usecase.ErrNotFound)
	}
	return nil
}

func (r *PostgresUserRepository) DeleteByID(ctx context.Context, userID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	execResult, err := r.database.ExecContext(ctx, `
		UPDATE users
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	rowsAffected, err := execResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("delete user: %w", usecase.ErrNotFound)
	}
	return nil
}

func isPostgresUniqueViolationError(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == postgresUniqueViolationSQLState
}
