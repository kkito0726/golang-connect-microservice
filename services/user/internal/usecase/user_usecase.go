package usecase

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/ken/connect-microservice/services/user/internal/domain"
)

// Authenticate はメールとパスワードでユーザーを認証して返す。
// メール不在・パスワード不一致どちらも同じエラーを返す（タイミング攻撃対策）。

type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) CreateUser(ctx context.Context, email, name, password, role string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	return uc.repo.Create(ctx, domain.User{
		Email:        email,
		Name:         name,
		Role:         role,
		PasswordHash: string(hash),
	})
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) ListUsers(ctx context.Context, pageSize, page int) ([]domain.User, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return uc.repo.List(ctx, pageSize, (page-1)*pageSize)
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, id, name, email string) (domain.User, error) {
	return uc.repo.Update(ctx, id, name, email)
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	return uc.repo.SoftDelete(ctx, id)
}

func (uc *UserUsecase) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	u, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, fmt.Errorf("invalid email or password")
	}
	return u, nil
}
