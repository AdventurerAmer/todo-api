package tasksrepo

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

func NewPostgres(db *sql.DB) ports.TasksRepository {
	return &postgres{
		db: db,
	}
}

func (repo *postgres) Create(ctx context.Context, task *domain.Task) error {
	query := `INSERT INTO lists (list_id, content, is_completed)
			  VALUES ($1, $2, $3)
			  RETURNING id, created_at, updated_at, version`

	row := repo.db.QueryRowContext(ctx, query, task.ListID, task.Content, task.IsCompleted)
	if err := row.Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt, &task.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return nil
}

func (repo *postgres) Get(ctx context.Context, id string) (domain.Task, error) {
	query := `SELECT created_at, updated_at, list_id, content, is_completed, version
			  FROM lists
			  WHERE id = $1`
	row := repo.db.QueryRowContext(ctx, query, id)
	task := domain.Task{ID: id}
	if err := row.Scan(&task.CreatedAt, &task.UpdatedAt, &task.ListID, &task.Content, &task.IsCompleted, &task.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, ports.ErrTaskNotFound
		}
		return domain.Task{}, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return task, nil
}

func (repo *postgres) Update(ctx context.Context, task *domain.Task) error {
	query := `UPDATE tasks 
			  SET content = $1, is_completed = $2, updated_at = NOW(), version = version + 1
			  WHERE id = $3 and version = $4
			  RETURNING version`

	row := repo.db.QueryRowContext(ctx, query, task.Content, task.IsCompleted, task.ID)
	if err := row.Scan(&task.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}

	return nil
}

func (repo *postgres) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks
			  WHERE id = $1`

	if _, err := repo.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}
