package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

func TestRegisterHashesPasswordAndReturnsToken(t *testing.T) {
	repo := newMemoryRepo()
	service := NewService(repo, "test-secret", time.Hour)

	session, err := service.Register(context.Background(), "Student@Example.com", "password123")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if session.Token == "" {
		t.Fatal("expected token")
	}
	if session.User.Email != "student@example.com" {
		t.Fatalf("expected normalized email, got %q", session.User.Email)
	}

	stored := repo.usersByEmail["student@example.com"]
	if stored.PasswordHash == "password123" {
		t.Fatal("password was stored without hashing")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("password123")); err != nil {
		t.Fatalf("stored hash does not match password: %v", err)
	}

	userID, err := service.ParseToken(session.Token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if userID != session.User.ID {
		t.Fatalf("expected user id %d, got %d", session.User.ID, userID)
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	repo := newMemoryRepo()
	service := NewService(repo, "test-secret", time.Hour)

	if _, err := service.Register(context.Background(), "student@example.com", "password123"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, err := service.Login(context.Background(), "student@example.com", "wrong-password")
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestRegisterDuplicateEmailReturnsConflict(t *testing.T) {
	repo := newMemoryRepo()
	service := NewService(repo, "test-secret", time.Hour)

	if _, err := service.Register(context.Background(), "student@example.com", "password123"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, err := service.Register(context.Background(), "STUDENT@example.com", "password123")
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestParseTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	service := NewService(newMemoryRepo(), "test-secret", time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   "1",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}

	_, err = service.ParseToken(signed)
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

type memoryRepo struct {
	nextID       int64
	usersByID    map[int64]UserWithPassword
	usersByEmail map[string]UserWithPassword
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		nextID:       1,
		usersByID:    map[int64]UserWithPassword{},
		usersByEmail: map[string]UserWithPassword{},
	}
}

func (r *memoryRepo) CreateUser(_ context.Context, email, passwordHash string) (User, error) {
	if _, exists := r.usersByEmail[email]; exists {
		return User{}, app.Conflict("Пользователь с таким email уже существует")
	}
	user := UserWithPassword{
		User: User{
			ID:        r.nextID,
			Email:     email,
			CreatedAt: time.Now(),
		},
		PasswordHash: passwordHash,
	}
	r.nextID++
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user
	return user.User, nil
}

func (r *memoryRepo) GetUserByEmail(_ context.Context, email string) (UserWithPassword, error) {
	user, exists := r.usersByEmail[email]
	if !exists {
		return UserWithPassword{}, app.NotFound("Пользователь не найден")
	}
	return user, nil
}

func (r *memoryRepo) GetUserByID(_ context.Context, id int64) (User, error) {
	user, exists := r.usersByID[id]
	if !exists {
		return User{}, app.NotFound("Пользователь не найден")
	}
	return user.User, nil
}
