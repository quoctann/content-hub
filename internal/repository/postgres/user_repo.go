package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/internal/domain"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) domain.UserRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) FetchActive(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE status = 'active'
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Email,
			&u.Password,
			&u.Status,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	var u domain.User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return domain.User{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	return u, nil
}
