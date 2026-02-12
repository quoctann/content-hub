package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
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
		INSERT INTO contents (title, search_data, link, file_name, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`
	err = tx.QueryRow(ctx, query, c.Title, c.SearchData, c.Link, c.FileName, c.Type).Scan(&c.ID)
	if err != nil {
		return fmt.Errorf("failed to insert content: %w", err)
	}

	// Handle tags
	if err := r.upsertTags(ctx, tx, c); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *contentRepo) upsertTags(ctx context.Context, tx pgx.Tx, c *domain.Content) error {
	if len(c.Tags) == 0 {
		return nil
	}

	for i, tag := range c.Tags {
		var tagID int64
		// If ID is set, verify existence? No, assuming valid if ID presented.
		// Usually we check by name to avoid duplicates if ID is missing.
		if tag.ID != 0 {
			tagID = tag.ID
		} else if tag.Name != "" {
			// Find or Create
			tagQuery := `
				INSERT INTO tags (name, created_at, updated_at)
				VALUES ($1, NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
				RETURNING id
			`
			err := tx.QueryRow(ctx, tagQuery, tag.Name).Scan(&tagID)
			if err != nil {
				return fmt.Errorf("failed to upsert tag '%s': %w", tag.Name, err)
			}
			c.Tags[i].ID = tagID
		} else {
			continue // Skip empty tags
		}

		linkQuery := `
			INSERT INTO content_tags (content_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		_, err := tx.Exec(ctx, linkQuery, c.ID, tagID)
		if err != nil {
			return fmt.Errorf("failed to link tag id %d: %w", tagID, err)
		}
	}
	return nil
}

func (r *contentRepo) Update(ctx context.Context, c *domain.Content) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE contents
		SET title = $1, search_data = $2, link = $3, file_name = $4, type = $5, updated_at = NOW()
		WHERE id = $6
	`
	tag, err := tx.Exec(ctx, query, c.Title, c.SearchData, c.Link, c.FileName, c.Type, c.ID)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("content not found")
	}

	// Remove existing tags mapping
	_, err = tx.Exec(ctx, "DELETE FROM content_tags WHERE content_id = $1", c.ID)
	if err != nil {
		return fmt.Errorf("failed to clear existing tags: %w", err)
	}

	// Re-add tags
	if err := r.upsertTags(ctx, tx, c); err != nil {
		return err
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

	selectClause := "DISTINCT c.id, c.title, c.search_data, c.link, c.type, c.created_at, c.updated_at"
	if filter.Query != "" {
		// Calculate queryArgIdx for rank calculation in SELECT
		// Matches the logic below for arg indices
		queryArgIdx := 1
		if len(filter.Tags) > 0 {
			queryArgIdx = 2
		}
		selectClause += fmt.Sprintf(", ts_rank(search_vector, plainto_tsquery('simple', unaccent($%d))) as rank", queryArgIdx)
	}

	baseQuery := fmt.Sprintf(`
		SELECT %s
		FROM contents c
	`, selectClause)
	var args []interface{}
	var conditions []string
	argID := 1

	// Join content_tags and tags tables when filtering by tags
	if len(filter.Tags) > 0 {
		baseQuery += `
		JOIN content_tags ct ON c.id = ct.content_id
		JOIN tags t ON ct.tag_id = t.id
		`
		conditions = append(conditions, fmt.Sprintf("LOWER(t.name) = ANY($%d)", argID))
		// Lowercase all tag names for case-insensitive matching
		lowerTags := make([]string, len(filter.Tags))
		for i, tag := range filter.Tags {
			lowerTags[i] = strings.ToLower(tag)
		}
		args = append(args, lowerTags)
		argID++
	}

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
	// Map to hold content IDs to fetch tags
	contentMap := make(map[int64]*domain.Content)
	var contentIDs []int64

	for rows.Next() {
		var c domain.Content
		var err error
		if filter.Query != "" {
			var rank float64
			err = rows.Scan(&c.ID, &c.Title, &c.SearchData, &c.Link, &c.Type, &c.CreatedAt, &c.UpdatedAt, &rank)
		} else {
			err = rows.Scan(&c.ID, &c.Title, &c.SearchData, &c.Link, &c.Type, &c.CreatedAt, &c.UpdatedAt)
		}
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		contents = append(contents, c)
	}

	if len(contents) > 0 {
		// Populate IDs and Map
		for i := range contents {
			contentIDs = append(contentIDs, contents[i].ID)
			contentMap[contents[i].ID] = &contents[i]
		}

		// Fetch Tags
		tagsQuery := `
			SELECT ct.content_id, t.id, t.name, t.created_at, t.updated_at
			FROM tags t
			JOIN content_tags ct ON t.id = ct.tag_id
			WHERE ct.content_id = ANY($1)
		`
		tagRows, err := r.db.Query(ctx, tagsQuery, contentIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var contentID int64
			var t domain.Tag
			if err := tagRows.Scan(&contentID, &t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
				return nil, fmt.Errorf("tag scan failed: %w", err)
			}
			if c, ok := contentMap[contentID]; ok {
				c.Tags = append(c.Tags, t)
			}
		}
	}

	return contents, nil
}
