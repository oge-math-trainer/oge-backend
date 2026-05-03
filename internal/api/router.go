package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/oge-math-trainer/oge-backend.git/internal/auth"
	"github.com/oge-math-trainer/oge-backend.git/internal/diagnostic"
	"github.com/oge-math-trainer/oge-backend.git/internal/progress"
	"github.com/oge-math-trainer/oge-backend.git/internal/tasks"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (auth.Session, error)
	Login(ctx context.Context, email, password string) (auth.Session, error)
	Me(ctx context.Context, userID int64) (auth.User, error)
	ParseToken(token string) (int64, error)
}

type TaskService interface {
	Generate(ctx context.Context, req tasks.GenerateRequest) (tasks.Task, error)
	Check(ctx context.Context, req tasks.CheckRequest) (tasks.CheckResult, error)
	Hint(ctx context.Context, userID, taskID int64) (tasks.HintResult, error)
	Explain(ctx context.Context, userID, taskID int64) (tasks.ExplainResult, error)
}

type DiagnosticService interface {
	Start(ctx context.Context, userID int64) (diagnostic.StartResult, error)
	Submit(ctx context.Context, userID, sessionID int64, answers []diagnostic.AnswerInput) (diagnostic.SubmitResult, error)
}

type ProgressService interface {
	GetProgress(ctx context.Context, userID int64) ([]progress.Item, error)
	GetStats(ctx context.Context, userID int64) (progress.Stats, error)
	GetRecommendations(ctx context.Context, userID int64) ([]progress.Recommendation, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type Dependencies struct {
	Auth                  AuthService
	Tasks                 TaskService
	Diagnostic            DiagnosticService
	Progress              ProgressService
	Health                HealthChecker
	CORSAllowedOrigins    []string
	AuthRateLimitRequests int
	AuthRateLimitWindow   time.Duration
}

type Server struct {
	auth     AuthService
	tasks    TaskService
	diag     DiagnosticService
	progress ProgressService
	health   HealthChecker
	cors     map[string]struct{}

	authLimiter *rateLimiter
}

func NewRouter(deps Dependencies) http.Handler {
	authRateLimitRequests := deps.AuthRateLimitRequests
	if authRateLimitRequests <= 0 {
		authRateLimitRequests = 10
	}
	authRateLimitWindow := deps.AuthRateLimitWindow
	if authRateLimitWindow <= 0 {
		authRateLimitWindow = time.Minute
	}

	server := &Server{
		auth:     deps.Auth,
		tasks:    deps.Tasks,
		diag:     deps.Diagnostic,
		progress: deps.Progress,
		health:   deps.Health,
		cors:     make(map[string]struct{}),

		authLimiter: newRateLimiter(authRateLimitRequests, authRateLimitWindow),
	}
	for _, origin := range deps.CORSAllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			server.cors[origin] = struct{}{}
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", server.handleHealth)

	mux.Handle("POST /api/v1/auth/register", server.authRateLimit(http.HandlerFunc(server.handleRegister)))
	mux.Handle("POST /api/v1/auth/login", server.authRateLimit(http.HandlerFunc(server.handleLogin)))

	mux.Handle("GET /api/v1/auth/me", server.requireAuth(http.HandlerFunc(server.handleMe)))
	mux.Handle("POST /api/v1/diagnostic/start", server.requireAuth(http.HandlerFunc(server.handleDiagnosticStart)))
	mux.Handle("POST /api/v1/diagnostic/submit", server.requireAuth(http.HandlerFunc(server.handleDiagnosticSubmit)))
	mux.Handle("POST /api/v1/tasks/generate", server.requireAuth(http.HandlerFunc(server.handleTaskGenerate)))
	mux.Handle("POST /api/v1/tasks/check", server.requireAuth(http.HandlerFunc(server.handleTaskCheck)))
	mux.Handle("POST /api/v1/tasks/hint", server.requireAuth(http.HandlerFunc(server.handleTaskHint)))
	mux.Handle("POST /api/v1/tasks/explain", server.requireAuth(http.HandlerFunc(server.handleTaskExplain)))
	mux.Handle("GET /api/v1/progress", server.requireAuth(http.HandlerFunc(server.handleProgress)))
	mux.Handle("GET /api/v1/progress/stats", server.requireAuth(http.HandlerFunc(server.handleProgressStats)))
	mux.Handle("GET /api/v1/progress/recommendations", server.requireAuth(http.HandlerFunc(server.handleProgressRecommendations)))

	return server.requestIDMiddleware(server.securityHeadersMiddleware(server.corsMiddleware(server.recoverMiddleware(mux))))
}
