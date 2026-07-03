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
		WHERE id = $9 AND deleted_at IS NULL
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
	query := `UPDATE content SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft-delete content: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found or already deleted")
	}
	return nil
}

func (r *contentRepo) DeleteMany(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE content SET deleted_at = NOW(), updated_at = NOW() WHERE id = ANY($1) AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("failed to bulk soft-delete content: %w", err)
	}
	return nil
}

func (r *contentRepo) GetByID(ctx context.Context, id int64) (*domain.Content, error) {
	query := `
		SELECT id, title, text_data, ocr_text, caption, link, type, is_hidden, created_at, updated_at
		FROM content
		WHERE id = $1 AND deleted_at IS NULL
	`
	var c domain.Content
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Title, &c.TextData, &c.OCRText, &c.Caption, &c.Link, &c.Type, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	return &c, nil
}

func (r *contentRepo) SetHidden(ctx context.Context, id int64, hidden bool) error {
	query := `UPDATE content SET is_hidden = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, query, hidden, id)
	if err != nil {
		return fmt.Errorf("failed to set hidden: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found")
	}
	return nil
}

// tsQuery holds the resolved tsquery function name and its single argument,
// ready to be embedded as tsFunc('simple', unaccent($1)) in SQL.
type tsQuery struct {
	fn  string // "plainto_tsquery" or "to_tsquery"
	arg string // the query string to pass as $1
}

// buildTSQuery resolves the correct PostgreSQL tsquery variant from the filter.
//
//   - No keywords: zero value (no FTS condition added)
//   - Single keyword: plainto_tsquery (handles tokenisation automatically)
//   - Multiple keywords: to_tsquery with explicit & / | operators
//
// Single-quote escaping is applied only for the multi-keyword to_tsquery path,
// which needs the operator syntax, plainto_tsquery never needs it.
func buildTSQuery(filter domain.SearchFilter) (tsQuery, bool) {
	keywords := filter.Keywords

	// Normalize: strip blank entries produced by trailing commas or extra spaces.
	cleaned := keywords[:0]
	for _, kw := range keywords {
		if trimmed := strings.TrimSpace(kw); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	keywords = cleaned

	switch {
	case len(keywords) == 0:
		return tsQuery{}, false // no FTS - browse / list mode

	case len(keywords) == 1:
		// Single keyword; plainto_tsquery is simpler and handles all edge cases.
		return tsQuery{fn: "plainto_tsquery", arg: keywords[0]}, true

	default:
		// Multiple keywords: build an explicit operator expression for to_tsquery.
		op := " | "
		if filter.MatchType == "and" {
			op = " & "
		}
		// Escape single quotes so they are valid inside a to_tsquery string literal.
		escaped := make([]string, len(keywords))
		for i, kw := range keywords {
			escaped[i] = strings.ReplaceAll(kw, "'", "''")
		}
		return tsQuery{fn: "to_tsquery", arg: strings.Join(escaped, op)}, true
	}
}

// visibilityCondition returns the SQL fragment (if any) that enforces the
// requested visibility policy.
//
// Public API (IncludeHidden=false): always exclude hidden rows
//
// Admin API (IncludeHidden=true): respect VisibilityFilter:
//
//	"true": visible only
//	"false": hidden only
//	"": no restriction
func visibilityCondition(filter domain.SearchFilter) string {
	if !filter.IncludeHidden {
		return "c.is_hidden = false"
	}
	switch filter.VisibilityFilter {
	case "true":
		return "c.is_hidden = false"
	case "false":
		return "c.is_hidden = true"
	default:
		return ""
	}
}

func (r *contentRepo) Search(ctx context.Context, filter domain.SearchFilter, cursor string, num int64) ([]domain.Content, int64, error) {
	offset := parseCursor(cursor)
	if num <= 0 {
		num = 10
	}

	tsq, hasFTS := buildTSQuery(filter)

	// args holds positional query parameters ($1, $2, …).
	// When FTS is active, $1 is always the tsquery argument so the SELECT and
	// WHERE clauses can reference it with a fixed placeholder.
	var args []any
	if hasFTS {
		args = append(args, tsq.arg) // $1
	}
	nextArg := len(args) + 1 // next available $N slot

	const baseColumns = "DISTINCT c.id, c.title, c.text_data, c.ocr_text, c.caption, c.link, c.type, c.is_hidden, c.created_at, c.updated_at"
	rankExpr := "0.0"
	if hasFTS {
		rankExpr = fmt.Sprintf("ts_rank(search_vector, %s('simple', public.unaccent($1)))", tsq.fn)
	}
	selectSQL := fmt.Sprintf("SELECT %s, %s AS rank FROM content c", baseColumns, rankExpr)

	conditions := []string{"c.deleted_at IS NULL"}

	if vis := visibilityCondition(filter); vis != "" {
		conditions = append(conditions, vis)
	}
	if hasFTS {
		conditions = append(conditions, fmt.Sprintf("search_vector @@ %s('simple', public.unaccent($1))", tsq.fn))
	}
	if filter.ContentType != "" {
		conditions = append(conditions, fmt.Sprintf("c.type = $%d", nextArg))
		args = append(args, string(filter.ContentType))
		nextArg++
	}

	whereSQL := " WHERE " + strings.Join(conditions, " AND ")

	orderSQL := " ORDER BY c.created_at DESC"
	if hasFTS {
		orderSQL = " ORDER BY rank DESC, c.created_at DESC"
	}

	countSQL := "SELECT COUNT(DISTINCT c.id) FROM content c" + whereSQL
	var totalCount int64
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count results: %w", err)
	}

	dataSQL := selectSQL + whereSQL + orderSQL + fmt.Sprintf(" LIMIT $%d OFFSET $%d", nextArg, nextArg+1)
	dataArgs := append(args, num, offset)

	rows, err := r.db.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var contents []domain.Content
	for rows.Next() {
		var c domain.Content
		var rank float64
		if err := rows.Scan(&c.ID, &c.Title, &c.TextData, &c.OCRText, &c.Caption, &c.Link, &c.Type, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt, &rank); err != nil {
			return nil, 0, fmt.Errorf("scan failed: %w", err)
		}
		c.Rank = rank
		contents = append(contents, c)
	}

	return contents, totalCount, nil
}

// parseCursor decodes a cursor string (a stringified integer offset) into an
// int64. An empty or invalid cursor returns 0.
func parseCursor(cursor string) int64 {
	if cursor == "" {
		return 0
	}
	var offset int64
	_, _ = fmt.Sscanf(cursor, "%d", &offset)
	if offset < 0 {
		return 0
	}
	return offset
}
