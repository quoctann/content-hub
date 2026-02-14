package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/internal/domain"
)

type contentRepo struct {
	db *pgxpool.Pool
}

func NewContentRepo(db *pgxpool.Pool) domain.ContentRepository {
	return &contentRepo{
		db: db,
	}
}

func (r *contentRepo) Create(ctx context.Context, c *domain.Content) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO content (title, text_data, ocr_text, caption, link, file_name, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`
	err = tx.QueryRow(ctx, query, c.Title, c.TextData, c.OCRText, c.Caption, c.Link, c.FileName, c.Type).Scan(&c.ID)
	if err != nil {
		return fmt.Errorf("failed to insert content: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *contentRepo) Update(ctx context.Context, c *domain.Content) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE content
		SET title = $1, text_data = $2, ocr_text = $3, caption = $4, link = $5, file_name = $6, type = $7, updated_at = NOW()
		WHERE id = $8
	`
	tag, err := tx.Exec(ctx, query, c.Title, c.TextData, c.OCRText, c.Caption, c.Link, c.FileName, c.Type, c.ID)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found")
	}

	return tx.Commit(ctx)
}

func (r *contentRepo) Search(ctx context.Context, filter domain.SearchFilter, cursor string, num int64) ([]domain.Content, error) {
	var offset int
	if cursor != "" {
		fmt.Sscanf(cursor, "%d", &offset)
	}
	if offset < 0 {
		offset = 0
	}

	limit := num
	if limit <= 0 {
		limit = 20 // Default limit
	}

	selectClause := "DISTINCT c.id, c.title, c.text_data, c.ocr_text, c.caption, c.link, c.type, c.created_at, c.updated_at"
	if filter.Query != "" {
		// Calculate queryArgIdx for rank calculation in SELECT
		queryArgIdx := 1
		selectClause += fmt.Sprintf(", ts_rank(search_vector, plainto_tsquery('simple', unaccent($%d))) as rank", queryArgIdx)
	}

	baseQuery := fmt.Sprintf(`
		SELECT %s
		FROM content c
	`, selectClause)
	var args []interface{}
	var conditions []string
	argID := 1

	if filter.Query != "" {
		// Use unaccent for both the document (in search_vector) and the query
		// 'simple' dictionary is used to avoid stemming which might conflict with vietnamese unaccenting
		conditions = append(conditions, fmt.Sprintf("search_vector @@ plainto_tsquery('simple', unaccent($%d))", argID))
		args = append(args, filter.Query)
		argID++
	}

	if filter.ContentType != "" {
		conditions = append(conditions, fmt.Sprintf("c.type = $%d", argID))
		args = append(args, string(filter.ContentType))
		argID++
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by rank if query is present, otherwise by created_at desc
	if filter.Query != "" {
		baseQuery += " ORDER BY rank DESC, c.created_at DESC"
	} else {
		baseQuery += " ORDER BY c.created_at DESC"
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var contents []domain.Content

	for rows.Next() {
		var c domain.Content
		var err error
		if filter.Query != "" {
			var rank float64
			err = rows.Scan(&c.ID, &c.Title, &c.TextData, &c.OCRText, &c.Caption, &c.Link, &c.Type, &c.CreatedAt, &c.UpdatedAt, &rank)
		} else {
			err = rows.Scan(&c.ID, &c.Title, &c.TextData, &c.OCRText, &c.Caption, &c.Link, &c.Type, &c.CreatedAt, &c.UpdatedAt)
		}
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		contents = append(contents, c)
	}

	return contents, nil
}
