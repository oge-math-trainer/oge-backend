package auth

import (
	"context"
	"errors"
	"net/url"
	"strings"
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

func TestOAuthStartBuildsGoogleAuthorizationURL(t *testing.T) {
	service := NewService(newMemoryRepo(), "test-secret", time.Hour)
	service.ConfigureOAuth(OAuthConfig{
		Google: OAuthProviderConfig{
			ClientID:     "google-client-id",
			ClientSecret: "google-client-secret",
			RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/google/callback",
		},
	})

	start, err := service.OAuthStart(context.Background(), ProviderGoogle)
	if err != nil {
		t.Fatalf("OAuthStart returned error: %v", err)
	}
	parsed, err := url.Parse(start.URL)
	if err != nil {
		t.Fatalf("invalid auth url: %v", err)
	}
	query := parsed.Query()
	if got := query.Get("client_id"); got != "google-client-id" {
		t.Fatalf("expected client id, got %q", got)
	}
	if got := query.Get("response_type"); got != "code" {
		t.Fatalf("expected code response type, got %q", got)
	}
	if scope := query.Get("scope"); !strings.Contains(scope, "openid") || !strings.Contains(scope, "email") {
		t.Fatalf("unexpected scope %q", scope)
	}
	if err := service.verifyOAuthState(ProviderGoogle, query.Get("state")); err != nil {
		t.Fatalf("state did not verify: %v", err)
	}
}

func TestOAuthCallbackRejectsTamperedState(t *testing.T) {
	service := NewService(newMemoryRepo(), "test-secret", time.Hour)
	service.ConfigureOAuth(OAuthConfig{
		Yandex: OAuthProviderConfig{
			ClientID:     "yandex-client-id",
			ClientSecret: "yandex-client-secret",
			RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/yandex/callback",
		},
	})

	_, err := service.OAuthCallback(context.Background(), ProviderYandex, "code", "bad-state")
	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *app.Error
	if !errors.As(err, &appErr) || appErr.Code != app.CodeUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
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
	identities   map[string]int64
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		nextID:       1,
		usersByID:    map[int64]UserWithPassword{},
		usersByEmail: map[string]UserWithPassword{},
		identities:   map[string]int64{},
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

func (r *memoryRepo) FindOrCreateOAuthUser(_ context.Context, identity OAuthIdentity) (User, error) {
	key := identity.Provider + ":" + identity.ProviderUserID
	if userID, exists := r.identities[key]; exists {
		user, exists := r.usersByID[userID]
		if !exists {
			return User{}, app.NotFound("user not found")
		}
		return user.User, nil
	}

	if existing, exists := r.usersByEmail[identity.Email]; exists {
		r.identities[key] = existing.ID
		return existing.User, nil
	}

	user := UserWithPassword{
		User: User{
			ID:        r.nextID,
			Email:     identity.Email,
			CreatedAt: time.Now(),
		},
	}
	r.nextID++
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user
	r.identities[key] = user.ID
	return user.User, nil
}

func (r *memoryRepo) GetUserByID(_ context.Context, id int64) (User, error) {
	user, exists := r.usersByID[id]
	if !exists {
		return User{}, app.NotFound("Пользователь не найден")
	}
	return user.User, nil
}
