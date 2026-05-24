package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/ken/connect-microservice/services/user/db/sqlc"
	"github.com/ken/connect-microservice/services/user/internal/domain"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: db.New(pool)}
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u domain.User) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        u.Email,
		Name:         u.Name,
		Role:         u.Role,
		PasswordHash: u.PasswordHash,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return userFromRow(row.ID, row.Email, row.Name, row.Role, row.PasswordHash, row.CreatedAt, row.UpdatedAt), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("get user %s: %w", id, domain.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return userFromRow(row.ID, row.Email, row.Name, row.Role, row.PasswordHash, row.CreatedAt, row.UpdatedAt), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("get user by email: %w", domain.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return userFromRow(row.ID, row.Email, row.Name, row.Role, row.PasswordHash, row.CreatedAt, row.UpdatedAt), nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	total, err := r.queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	users := make([]domain.User, len(rows))
	for i, row := range rows {
		users[i] = userFromRow(row.ID, row.Email, row.Name, row.Role, row.PasswordHash, row.CreatedAt, row.UpdatedAt)
	}
	return users, int(total), nil
}

func (r *UserRepository) Update(ctx context.Context, id, name, email string) (domain.User, error) {
	row, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		Name:  name,
		Email: email,
		ID:    id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("update user %s: %w", id, domain.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return userFromRow(row.ID, row.Email, row.Name, row.Role, row.PasswordHash, row.CreatedAt, row.UpdatedAt), nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	affected, err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("delete user %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

func userFromRow(id, email, name, role, passwordHash string, createdAt, updatedAt pgtype.Timestamptz) domain.User {
	return domain.User{
		ID:           id,
		Email:        email,
		Name:         name,
		Role:         role,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}
}
