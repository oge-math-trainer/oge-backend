package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/oge-math-trainer/oge-backend.git/internal/app"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Session struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (UserWithPassword, error)
	GetUserByID(ctx context.Context, id int64) (User, error)
}

type UserWithPassword struct {
	User
	PasswordHash string
}

type Service struct {
	repo   Repository
	secret string
	ttl    time.Duration
	now    func() time.Time
}

func NewService(repo Repository, authSecret string, tokenTTL time.Duration) *Service {
	return &Service{
		repo:   repo,
		secret: authSecret,
		ttl:    tokenTTL,
		now:    time.Now,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (Session, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, app.Internal(err)
	}

	user, err := s.repo.CreateUser(ctx, normalizeEmail(email), string(hash))
	if err != nil {
		return Session{}, err
	}

	token, err := s.createToken(user.ID)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, User: user}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	user, err := s.repo.GetUserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if isNotFound(err) {
			consumePasswordCompare(password)
			return Session{}, app.Unauthorized("Неверный email или пароль")
		}
		return Session{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return Session{}, app.Unauthorized("Неверный email или пароль")
	}

	token, err := s.createToken(user.ID)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, User: user.User}, nil
}

func (s *Service) Me(ctx context.Context, userID int64) (User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) ParseToken(token string) (int64, error) {
	if strings.TrimSpace(s.secret) == "" {
		return 0, app.Unauthorized("Требуется авторизация")
	}

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected jwt signing method: %v", t.Header["alg"])
			}
			return []byte(s.secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
		jwt.WithTimeFunc(s.now),
	)
	if err != nil {
		return 0, app.Unauthorized("Требуется авторизация")
	}
	if parsed == nil || !parsed.Valid {
		return 0, app.Unauthorized("Требуется авторизация")
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, app.Unauthorized("Требуется авторизация")
	}
	return userID, nil
}

func (s *Service) createToken(userID int64) (string, error) {
	if strings.TrimSpace(s.secret) == "" {
		return "", app.Internal(errors.New("AUTH_SECRET is empty"))
	}

	now := s.now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", app.Internal(err)
	}
	return signed, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isNotFound(err error) bool {
	var appErr *app.Error
	return errors.As(err, &appErr) && appErr.Code == app.CodeNotFound
}

func (u User) String() string {
	return fmt.Sprintf("%d:%s", u.ID, u.Email)
}

var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

func consumePasswordCompare(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
}
