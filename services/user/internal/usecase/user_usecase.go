package usecase

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/ken/connect-microservice/services/user/internal/repository"
)

type UserRepository interface {
	Create(ctx context.Context, u repository.User) (repository.User, error)
	GetByID(ctx context.Context, id string) (repository.User, error)
	List(ctx context.Context, limit, offset int) ([]repository.User, int, error)
	Update(ctx context.Context, id, name, email string) (repository.User, error)
	SoftDelete(ctx context.Context, id string) error
}

type UserUsecase struct {
	repo UserRepository
}

func NewUserUsecase(repo UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) CreateUser(ctx context.Context, email, name, password, role string) (repository.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return repository.User{}, fmt.Errorf("hash password: %w", err)
	}
	return uc.repo.Create(ctx, repository.User{
		Email:        email,
		Name:         name,
		Role:         role,
		PasswordHash: string(hash),
	})
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (repository.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) ListUsers(ctx context.Context, pageSize, page int) ([]repository.User, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, pageSize, (page-1)*pageSize)
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, id, name, email string) (repository.User, error) {
	return uc.repo.Update(ctx, id, name, email)
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	return uc.repo.SoftDelete(ctx, id)
}
