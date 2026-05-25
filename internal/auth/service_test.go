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

func TestPasswordResetSendsCodeAndUpdatesPassword(t *testing.T) {
	repo := newMemoryRepo()
	service := NewService(repo, "test-secret", time.Hour)
	mailer := &fakeResetMailer{}
	service.ConfigurePasswordReset(mailer, 15*time.Minute)

	if _, err := service.Register(context.Background(), "student@example.com", "old-password"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if err := service.RequestPasswordReset(context.Background(), "student@example.com"); err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if mailer.email != "student@example.com" {
		t.Fatalf("expected reset email, got %q", mailer.email)
	}
	if len(mailer.code) != 6 {
		t.Fatalf("expected six digit code, got %q", mailer.code)
	}

	session, err := service.ResetPassword(context.Background(), "student@example.com", mailer.code, "new-password")
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if session.Token == "" {
		t.Fatal("expected token")
	}

	if _, err := service.Login(context.Background(), "student@example.com", "old-password"); err == nil {
		t.Fatal("expected old password to be rejected")
	}
	if _, err := service.Login(context.Background(), "student@example.com", "new-password"); err != nil {
		t.Fatalf("expected login with new password: %v", err)
	}
}

func TestPasswordResetDoesNotSendEmailForUnknownUser(t *testing.T) {
	service := NewService(newMemoryRepo(), "test-secret", time.Hour)
	mailer := &fakeResetMailer{}
	service.ConfigurePasswordReset(mailer, 15*time.Minute)

	if err := service.RequestPasswordReset(context.Background(), "unknown@example.com"); err != nil {
		t.Fatalf("RequestPasswordReset returned error: %v", err)
	}
	if mailer.email != "" || mailer.code != "" {
		t.Fatalf("expected no email to be sent, got email=%q code=%q", mailer.email, mailer.code)
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
	resetCodes   []passwordResetRecord
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		nextID:       1,
		usersByID:    map[int64]UserWithPassword{},
		usersByEmail: map[string]UserWithPassword{},
		identities:   map[string]int64{},
		resetCodes:   []passwordResetRecord{},
	}
}

type passwordResetRecord struct {
	UserID    int64
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type fakeResetMailer struct {
	email     string
	code      string
	expiresAt time.Time
}

func (m *fakeResetMailer) SendPasswordResetCode(_ context.Context, email, code string, expiresAt time.Time) error {
	m.email = email
	m.code = code
	m.expiresAt = expiresAt
	return nil
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

func (r *memoryRepo) CreatePasswordResetCode(_ context.Context, email, codeHash string, expiresAt time.Time) (bool, error) {
	user, exists := r.usersByEmail[email]
	if !exists {
		return false, nil
	}
	now := time.Now()
	for i := range r.resetCodes {
		if r.resetCodes[i].UserID == user.ID && r.resetCodes[i].UsedAt == nil {
			r.resetCodes[i].UsedAt = &now
		}
	}
	r.resetCodes = append(r.resetCodes, passwordResetRecord{
		UserID:    user.ID,
		CodeHash:  codeHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	})
	return true, nil
}

func (r *memoryRepo) ResetPasswordWithCode(_ context.Context, email, codeHash, passwordHash string, now time.Time) (User, error) {
	user, exists := r.usersByEmail[email]
	if !exists {
		return User{}, app.Unauthorized("Invalid or expired password reset code")
	}
	for i := len(r.resetCodes) - 1; i >= 0; i-- {
		code := &r.resetCodes[i]
		if code.UserID != user.ID || code.CodeHash != codeHash || code.UsedAt != nil || !code.ExpiresAt.After(now) {
			continue
		}
		code.UsedAt = &now
		user.PasswordHash = passwordHash
		r.usersByID[user.ID] = user
		r.usersByEmail[user.Email] = user
		return user.User, nil
	}
	return User{}, app.Unauthorized("Invalid or expired password reset code")
}

func (r *memoryRepo) GetUserByID(_ context.Context, id int64) (User, error) {
	user, exists := r.usersByID[id]
	if !exists {
		return User{}, app.NotFound("Пользователь не найден")
	}
	return user.User, nil
}
