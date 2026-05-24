package usecase

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/ken/connect-microservice/services/user/internal/domain"
)

// ---- モック ----------------------------------------------------------------

type mockUserRepo struct {
	users          map[string]domain.User
	byEmail        map[string]domain.User
	createFn       func(ctx context.Context, u domain.User) (domain.User, error)
	lastListLimit  int
	lastListOffset int
}

func (m *mockUserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, u)
	}
	u.ID = "user-1"
	return u, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) List(_ context.Context, limit, offset int) ([]domain.User, int, error) {
	m.lastListLimit = limit
	m.lastListOffset = offset
	return nil, 0, nil
}

func (m *mockUserRepo) Update(_ context.Context, id, name, email string) (domain.User, error) {
	return domain.User{ID: id, Name: name, Email: email}, nil
}

func (m *mockUserRepo) SoftDelete(_ context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return domain.ErrNotFound
	}
	return nil
}

// ---- CreateUser -------------------------------------------------------------

func TestCreateUser_HashesPassword(t *testing.T) {
	t.Parallel()

	var captured domain.User
	repo := &mockUserRepo{
		createFn: func(_ context.Context, u domain.User) (domain.User, error) {
			captured = u
			u.ID = "user-1"
			return u, nil
		},
	}
	uc := NewUserUsecase(repo)

	_, err := uc.CreateUser(context.Background(), "a@example.com", "Alice", "secret", "user")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if captured.PasswordHash == "secret" {
		t.Error("password must not be stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(captured.PasswordHash), []byte("secret")); err != nil {
		t.Errorf("stored hash does not match original password: %v", err)
	}
}

func TestCreateUser_RepoError(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ domain.User) (domain.User, error) {
			return domain.User{}, domain.ErrAlreadyExists
		},
	}
	uc := NewUserUsecase(repo)

	_, err := uc.CreateUser(context.Background(), "dup@example.com", "Dup", "pass", "user")

	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("want ErrAlreadyExists, got %v", err)
	}
}

// ---- GetUser ----------------------------------------------------------------

func TestGetUser_Success(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{
		users: map[string]domain.User{
			"u1": {ID: "u1", Name: "Alice", Email: "a@example.com"},
		},
	}
	uc := NewUserUsecase(repo)

	u, err := uc.GetUser(context.Background(), "u1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if u.Name != "Alice" {
		t.Errorf("Name = %q, want %q", u.Name, "Alice")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{users: map[string]domain.User{}})

	_, err := uc.GetUser(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// ---- ListUsers --------------------------------------------------------------

func TestListUsers_PageNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pageSize, page int
		wantLimit      int
		wantOffset     int
	}{
		{"zero values default to 20/0", 0, 0, 20, 0},
		{"negative values default to 20/0", -1, -1, 20, 0},
		{"page 2 with size 10", 10, 2, 10, 10},
		{"page 3 with size 5", 5, 3, 5, 10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &mockUserRepo{}
			uc := NewUserUsecase(repo)

			uc.ListUsers(context.Background(), tt.pageSize, tt.page)

			if repo.lastListLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", repo.lastListLimit, tt.wantLimit)
			}
			if repo.lastListOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", repo.lastListOffset, tt.wantOffset)
			}
		})
	}
}

// ---- UpdateUser -------------------------------------------------------------

func TestUpdateUser_Success(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{})

	u, err := uc.UpdateUser(context.Background(), "u1", "Bob", "b@example.com")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if u.Name != "Bob" {
		t.Errorf("Name = %q, want %q", u.Name, "Bob")
	}
	if u.Email != "b@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "b@example.com")
	}
}

// ---- DeleteUser -------------------------------------------------------------

func TestDeleteUser_Success(t *testing.T) {
	t.Parallel()

	repo := &mockUserRepo{
		users: map[string]domain.User{"u1": {ID: "u1"}},
	}
	uc := NewUserUsecase(repo)

	err := uc.DeleteUser(context.Background(), "u1")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{users: map[string]domain.User{}})

	err := uc.DeleteUser(context.Background(), "missing")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// ---- Authenticate -----------------------------------------------------------

func TestAuthenticate_Success(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("setup: bcrypt failed: %v", err)
	}
	repo := &mockUserRepo{
		byEmail: map[string]domain.User{
			"alice@example.com": {ID: "u1", Email: "alice@example.com", PasswordHash: string(hash)},
		},
	}
	uc := NewUserUsecase(repo)

	u, err := uc.Authenticate(context.Background(), "alice@example.com", "correct-pass")

	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if u.ID != "u1" {
		t.Errorf("ID = %q, want %q", u.ID, "u1")
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{byEmail: map[string]domain.User{}})

	_, err := uc.Authenticate(context.Background(), "nobody@example.com", "pass")

	// タイミング攻撃対策で「メール不在」「パスワード不一致」を区別しない
	if err == nil {
		t.Fatal("want error for unknown email, got nil")
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("setup: bcrypt failed: %v", err)
	}
	repo := &mockUserRepo{
		byEmail: map[string]domain.User{
			"alice@example.com": {ID: "u1", Email: "alice@example.com", PasswordHash: string(hash)},
		},
	}
	uc := NewUserUsecase(repo)

	_, err = uc.Authenticate(context.Background(), "alice@example.com", "wrong-pass")

	if err == nil {
		t.Fatal("want error for wrong password, got nil")
	}
}

func TestAuthenticate_SameErrorForBothFailures(t *testing.T) {
	t.Parallel()

	// タイミング攻撃対策として、メール不在とパスワード不一致で同じエラーメッセージを返す
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)
	repo := &mockUserRepo{
		byEmail: map[string]domain.User{
			"exists@example.com": {Email: "exists@example.com", PasswordHash: string(hash)},
		},
	}
	uc := NewUserUsecase(repo)

	_, errNoEmail := uc.Authenticate(context.Background(), "nobody@example.com", "pass")
	_, errWrongPw := uc.Authenticate(context.Background(), "exists@example.com", "wrong")

	if errNoEmail == nil || errWrongPw == nil {
		t.Fatal("both cases must return errors")
	}
	if errNoEmail.Error() != errWrongPw.Error() {
		t.Errorf("error messages differ: %q vs %q — timing attack mitigation broken", errNoEmail.Error(), errWrongPw.Error())
	}
}
