package tasksrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

func (repo *postgres) Get(ctx context.Context, userID, id string) (domain.Task, error) {
	query := `SELECT created_at, updated_at, list_id, content, is_completed, version
			  FROM lists
			  WHERE user_id = $1 AND id = $2`
	row := repo.db.QueryRowContext(ctx, query, userID, id)
	task := domain.Task{ID: id}
	if err := row.Scan(&task.CreatedAt, &task.UpdatedAt, &task.ListID, &task.Content, &task.IsCompleted, &task.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, ports.ErrTaskNotFound
		}
		return domain.Task{}, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return task, nil
}

func (repo *postgres) GetAll(ctx context.Context, userID, listID string, page, pageSize int, sort, content string, isCompleted *bool) ([]domain.Task, int, error) {
	order := "ASC"
	if strings.HasPrefix(sort, "-") {
		order = "DESC"
		sort, _ = strings.CutPrefix(sort, "-")
	}
	sortStr := fmt.Sprintf("%s %s", sort, order)
	if sort != "id" {
		sortStr = fmt.Sprintf("%s %s, id ASC", sort, order)
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
			  SELECT count(*) OVER(), id, created_at, updated_at, content, is_completed, version
			  FROM tasks
			  WHERE list_id = $1 AND user_id = $2 AND ($3 = '' OR to_tsvector('simple', content) @@ plainto_tsquery('simple', $3)) AND ($4 = NULL OR is_completed = $4)
			  ORDER BY %s
			  LIMIT $5 OFFSET $6`, sortStr)

	rows, err := repo.db.QueryContext(ctx, query, userID, listID, content, isCompleted, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("'QueryContext' failed: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	total := 0
	for rows.Next() {
		task := domain.Task{
			ListID: listID,
			UserID: userID,
		}
		if err := rows.Scan(&total, &task.ID, &task.CreatedAt, &task.UpdatedAt, &task.Content, &task.IsCompleted, &task.Version); err != nil {
			return nil, 0, fmt.Errorf("'rows.Scan' failed: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("'rows' failed: %w", err)
	}
	return tasks, total, nil
}

func (repo *postgres) Update(ctx context.Context, task *domain.Task) error {
	query := `UPDATE tasks 
			  SET content = $1, is_completed = $2, updated_at = NOW(), version = version + 1
			  WHERE user_id = $3 AND id = $4 AND version = $5
			  RETURNING version`

	row := repo.db.QueryRowContext(ctx, query, task.Content, task.IsCompleted, task.UserID, task.ID)
	if err := row.Scan(&task.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}

	return nil
}

func (repo *postgres) Delete(ctx context.Context, userID, id string) error {
	query := `DELETE FROM tasks
			  WHERE user_id = $1 AND id = $2`

	if _, err := repo.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}
