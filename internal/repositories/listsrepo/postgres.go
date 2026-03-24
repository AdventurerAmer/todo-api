package listsrepo

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

func NewPostgres(db *sql.DB) ports.ListsRepository {
	return &postgres{
		db: db,
	}
}

func (repo *postgres) Create(ctx context.Context, list *domain.List) error {
	query := `INSERT INTO lists (title, user_id)
			  VALUES ($1, $2)
			  RETURNING id, created_at, updated_at, version`

	row := repo.db.QueryRowContext(ctx, query, list.Title, list.UserID)
	if err := row.Scan(&list.ID, &list.CreatedAt, &list.UpdatedAt, &list.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return nil
}

func (repo *postgres) Get(ctx context.Context, id string) (domain.List, error) {
	query := `SELECT created_at, updated_at, user_id, title, version
			  FROM lists
			  WHERE id = $1`
	row := repo.db.QueryRowContext(ctx, query, id)
	list := domain.List{ID: id}
	if err := row.Scan(&list.CreatedAt, &list.UpdatedAt, &list.UserID, &list.Title, &list.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.List{}, ports.ErrListNotFound
		}
		return domain.List{}, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return list, nil
}

func (repo *postgres) Update(ctx context.Context, list *domain.List) error {
	query := `UPDATE lists 
			  SET title = $1, description = $2, updated_at = NOW(), version = version + 1
			  WHERE id = $3 and version = $4
			  RETURNING version`

	row := repo.db.QueryRowContext(ctx, query, list.Title, list.Description, list.ID)
	if err := row.Scan(&list.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}

	return nil
}

func (repo *postgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM lists
			  WHERE id = $1`

	if _, err := repo.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}
