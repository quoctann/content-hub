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
		UPDATE content SET
			title = $1,
			text_data = $2,
			ocr_text = $3,
			caption = $4,
			link = $5,
			file_name = $6,
			type = $7,
			is_hidden = $8,
			updated_at = NOW()
		WHERE id = $9
	`
	tag, err := tx.Exec(ctx, query, c.Title, c.TextData, c.OCRText, c.Caption, c.Link, c.FileName, c.Type, c.IsHidden, c.ID)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found")
	}

	return tx.Commit(ctx)
}

func (r *contentRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM content WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found")
	}
	return nil
}

func (r *contentRepo) GetByID(ctx context.Context, id int64) (*domain.Content, error) {
	query := `
		SELECT id, title, text_data, ocr_text, caption, link, type, is_hidden, created_at, updated_at
		FROM content
		WHERE id = $1
	`
	var c domain.Content
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Title, &c.TextData, &c.OCRText, &c.Caption, &c.Link, &c.Type, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	return &c, nil
}

func (r *contentRepo) SetHidden(ctx context.Context, id int64, hidden bool) error {
	query := `UPDATE content SET is_hidden = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, hidden, id)
	if err != nil {
		return fmt.Errorf("failed to set hidden: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found")
	}
	return nil
}

func (r *contentRepo) Search(ctx context.Context, filter domain.SearchFilter, cursor string, num int64) ([]domain.Content, int64, error) {
	var offset int
	if cursor != "" {
		fmt.Sscanf(cursor, "%d", &offset)
	}
	if offset < 0 {
		offset = 0
	}

	limit := num
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Determine if we have a search query (from Keywords or legacy Query)
	hasSearchQuery := len(filter.Keywords) > 0 || filter.Query != ""

	// Build args slice and track parameter index
	var args []interface{}
	var tsqueryFunc string
	argID := 1 // Will be incremented as we add parameters

	if len(filter.Keywords) > 0 {
		searchArgs, funcName := r.buildSearchArgs(filter.Keywords, filter.MatchType)
		args = searchArgs
		tsqueryFunc = funcName
		argID = len(args) + 1 // Move to next available param slot
	} else if filter.Query != "" {
		args = []interface{}{filter.Query}
		tsqueryFunc = "plainto_tsquery"
		argID = 2 // Next slot after $1
	}

	// Build SELECT clause with ts_rank always included (0 if no search query)
	selectClause := "DISTINCT c.id, c.title, c.text_data, c.ocr_text, c.caption, c.link, c.type, c.is_hidden, c.created_at, c.updated_at"
	if hasSearchQuery {
		selectClause += fmt.Sprintf(", ts_rank(search_vector, %s('simple', unaccent($1))) as rank", tsqueryFunc)
	} else {
		selectClause += ", 0.0 as rank"
	}

	baseQuery := fmt.Sprintf(`
		SELECT %s
		FROM content c
	`, selectClause)

	var conditions []string

	// Visibility filter logic:
	// - IncludeHidden false (public API): always exclude hidden
	// - IncludeHidden true + VisibilityFilter "true": only visible (is_hidden = false)
	// - IncludeHidden true + VisibilityFilter "false": only hidden (is_hidden = true)
	// - IncludeHidden true + VisibilityFilter "": include all (no condition)
	if !filter.IncludeHidden {
		conditions = append(conditions, "c.is_hidden = false")
	} else if filter.VisibilityFilter == "true" {
		conditions = append(conditions, "c.is_hidden = false")
	} else if filter.VisibilityFilter == "false" {
		conditions = append(conditions, "c.is_hidden = true")
	}

	// Build search condition from Keywords (preferred) or legacy Query
	if len(filter.Keywords) > 0 {
		conditions = append(conditions, fmt.Sprintf("search_vector @@ %s('simple', unaccent($1))", tsqueryFunc))
	} else if filter.Query != "" {
		conditions = append(conditions, "search_vector @@ plainto_tsquery('simple', unaccent($1))")
	}

	if filter.ContentType != "" {
		conditions = append(conditions, fmt.Sprintf("c.type = $%d", argID))
		args = append(args, string(filter.ContentType))
		argID++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by rank DESC if query is present, then by created_at
	orderClause := " ORDER BY c.created_at DESC"
	if hasSearchQuery {
		orderClause = " ORDER BY rank DESC, c.created_at DESC"
	}

	// First, get total count
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT c.id) FROM content c%s", whereClause)
	countArgs := args[:] // Use same args (without limit/offset)
	var totalCount int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count results: %w", err)
	}

	// Then, get paginated results
	resultQuery := baseQuery + whereClause + orderClause + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	resultArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, resultQuery, resultArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var contents []domain.Content

	for rows.Next() {
		var c domain.Content
		var rank float64
		err := rows.Scan(&c.ID, &c.Title, &c.TextData, &c.OCRText, &c.Caption, &c.Link, &c.Type, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt, &rank)
		if err != nil {
			return nil, 0, fmt.Errorf("scan failed: %w", err)
		}
		c.Rank = rank
		contents = append(contents, c)
	}

	return contents, totalCount, nil
}

func (r *contentRepo) buildSearchArgs(keywords []string, matchType string) ([]interface{}, string) {
	// Default to OR if not specified
	if matchType == "" {
		matchType = "or"
	}

	// Filter out empty keywords
	var validKeywords []string
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			validKeywords = append(validKeywords, kw)
		}
	}

	if len(validKeywords) == 0 {
		return nil, "plainto_tsquery"
	}

	if len(validKeywords) == 1 {
		return []interface{}{validKeywords[0]}, "plainto_tsquery"
	}

	// Multiple keywords: build combined query string for to_tsquery
	separator := " & " // AND
	if matchType == "or" {
		separator = " | "
	}

	// Escape single quotes in keywords for to_tsquery
	var escapedKeywords []string
	for _, kw := range validKeywords {
		escaped := strings.ReplaceAll(kw, "'", "''")
		escapedKeywords = append(escapedKeywords, escaped)
	}
	queryStr := strings.Join(escapedKeywords, separator)

	return []interface{}{queryStr}, "to_tsquery"
}
