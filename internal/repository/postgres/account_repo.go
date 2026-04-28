package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/internal/domain"
)

type accountRepo struct {
	db *pgxpool.Pool
}

func NewAccountRepo(db *pgxpool.Pool) domain.AccountRepository {
	return &accountRepo{
		db: db,
	}
}

func (r *accountRepo) Create(ctx context.Context, acc *domain.Account) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO account (username, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	err = tx.QueryRow(ctx, query, acc.Username, acc.PasswordHash, acc.Role, acc.IsActive).Scan(&acc.ID)
	if err != nil {
		return fmt.Errorf("failed to insert account: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *accountRepo) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	query := `
		SELECT id, username, password_hash, role, is_active, created_at, updated_at
		FROM account
		WHERE id = $1
	`
	var account domain.Account
	err := r.db.QueryRow(ctx, query, id).Scan(
		&account.ID, &account.Username, &account.PasswordHash,
		&account.Role, &account.IsActive, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return &account, nil
}

func (r *accountRepo) GetByUsername(ctx context.Context, username string) (*domain.Account, error) {
	query := `
		SELECT id, username, password_hash, role, is_active, created_at, updated_at
		FROM account
		WHERE username = $1
	`
	var acc domain.Account
	err := r.db.QueryRow(ctx, query, username).Scan(
		&acc.ID, &acc.Username, &acc.PasswordHash,
		&acc.Role, &acc.IsActive, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account by username: %w", err)
	}
	return &acc, nil
}

func (r *accountRepo) Update(ctx context.Context, acc *domain.Account) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE account
		SET username = $1, password_hash = $2, role = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5
	`
	tag, err := tx.Exec(ctx, query, acc.Username, acc.PasswordHash, acc.Role, acc.IsActive, acc.ID)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}

	return tx.Commit(ctx)
}

func (r *accountRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM account WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}
