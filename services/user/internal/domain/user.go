package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Email        string
	Name         string
	Role         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	Create(ctx context.Context, u User) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	List(ctx context.Context, limit, offset int) ([]User, int, error)
	Update(ctx context.Context, id, name, email string) (User, error)
	SoftDelete(ctx context.Context, id string) error
}
