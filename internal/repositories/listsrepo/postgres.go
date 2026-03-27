package listsrepo

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

func (repo *postgres) GetAll(ctx context.Context, userID string, page, pageSize int, sort, title string) ([]domain.List, int, error) {
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
			  SELECT count(*) OVER(), id, created_at, updated_at, title, description, version
			  FROM lists
			  WHERE user_id = $1 AND ($2 = '' OR to_tsvector('simple', title) @@ plainto_tsquery('simple', $2))
			  ORDER BY %s
			  LIMIT $3 OFFSET $4`, sortStr)

	var lists []domain.List
	rows, err := repo.db.QueryContext(ctx, query, userID, title, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		list := domain.List{
			UserID: userID,
		}
		if err := rows.Scan(&total, &list.ID, &list.CreatedAt, &list.UpdatedAt, &list.Title, &list.Version); err != nil {
			return nil, 0, err
		}
		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return lists, total, nil
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
