package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quoctann/content-hub/internal/database/migrator"
	"github.com/quoctann/content-hub/internal/domain"
)

// newTestPool migrates a throwaway schema and returns a pool bound to it.
// Set TEST_DATABASE_URL (e.g. postgres://postgres:x@localhost:5432/postgres?sslmode=disable)
// to run; otherwise the test is skipped.
func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	baseURL := os.Getenv("TEST_DATABASE_URL")
	if baseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())

	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		_ = admin.Close(context.Background())
	})
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}

	url := baseURL + "&search_path=" + schema
	m, err := migrator.NewMigrator(url, "../../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Up(); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func ptr(s string) *string { return &s }

func TestContentRepoUpdatePatchSemantics(t *testing.T) {
	pool := newTestPool(t)
	repo := NewContentRepo(pool)
	ctx := context.Background()

	c := &domain.Content{Title: ptr("title"), Caption: ptr("caption"), Link: ptr("https://x/y.png"), Type: domain.Image}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE content SET file_name = 'y.png' WHERE id = $1", c.ID); err != nil {
		t.Fatal(err)
	}

	updated, err := repo.Update(ctx, c.ID, domain.ContentPatch{
		Caption:  domain.Some[*string](nil), // clear
		TextData: domain.Some(ptr("hello")), // set
		IsHidden: domain.Some(true),
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Caption != nil {
		t.Errorf("caption = %q, want NULL", *updated.Caption)
	}
	if updated.Title == nil || *updated.Title != "title" {
		t.Errorf("title = %v, want unchanged \"title\"", updated.Title)
	}
	if updated.TextData == nil || *updated.TextData != "hello" || !updated.IsHidden {
		t.Errorf("text_data/is_hidden not applied: %+v", updated)
	}

	var fileName *string
	if err := pool.QueryRow(ctx, "SELECT file_name FROM content WHERE id = $1", c.ID).Scan(&fileName); err != nil {
		t.Fatal(err)
	}
	if fileName == nil || *fileName != "y.png" {
		t.Errorf("file_name = %v, want untouched \"y.png\"", fileName)
	}

	same, err := repo.Update(ctx, c.ID, domain.ContentPatch{})
	if err != nil || same.ID != c.ID {
		t.Errorf("empty patch = (%v, %v), want current row", same, err)
	}
}

func TestContentRepoNotFound(t *testing.T) {
	pool := newTestPool(t)
	repo := NewContentRepo(pool)
	ctx := context.Background()
	const missing = int64(987654321)

	if _, err := repo.GetByID(ctx, missing); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID error = %v, want ErrNotFound", err)
	}
	if _, err := repo.Update(ctx, missing, domain.ContentPatch{Title: domain.Some(ptr("x"))}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update error = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, missing); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete error = %v, want ErrNotFound", err)
	}
	if err := repo.SetHidden(ctx, missing, true); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("SetHidden error = %v, want ErrNotFound", err)
	}
}
