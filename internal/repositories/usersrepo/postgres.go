package usersrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AdventurerAmer/todo-api/internal/core/domain"
	"github.com/AdventurerAmer/todo-api/internal/core/ports"
)

type postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) ports.UsersRepository {
	return &postgres{
		db: db,
	}
}

func (repo *postgres) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, password_hash, is_activated)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id, created_at, version`

	row := repo.db.QueryRowContext(ctx, query, user.Name, user.Email, user.PasswordHash, user.IsActivated)
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return nil
}

func (repo *postgres) Get(ctx context.Context, id string) (domain.User, error) {
	query := `SELECT created_at, name, email, password_hash, is_activated, version
			  FROM users
			  WHERE id = $1`
	row := repo.db.QueryRowContext(ctx, query, id)
	u := domain.User{ID: id}
	if err := row.Scan(&u.CreatedAt, &u.Name, &u.Email, &u.PasswordHash, &u.IsActivated, &u.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ports.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return u, nil
}

func (repo *postgres) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT id, created_at, name, password_hash, is_activated, version
			  FROM users
			  WHERE email = $1`
	row := repo.db.QueryRowContext(ctx, query, email)
	u := domain.User{Email: email}
	if err := row.Scan(&u.ID, &u.CreatedAt, &u.Name, &u.PasswordHash, &u.IsActivated, &u.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ports.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return u, nil
}

func (repo *postgres) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users 
			  SET name = $1, email = $2, password_hash = $3, is_activated = $4, version = version + 1
			  WHERE id = $5 and version = $6
			  RETURNING version`

	row := repo.db.QueryRowContext(ctx, query, user.Name, user.Email, user.PasswordHash, user.IsActivated, user.ID, user.Version)
	if err := row.Scan(&user.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}

	return nil
}

func (repo *postgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users
			  WHERE id = $1`

	if _, err := repo.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}
