package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"oge-ai-trainer/backend/internal/app"
	"oge-ai-trainer/backend/internal/auth"
	"oge-ai-trainer/backend/internal/diagnostic"
	"oge-ai-trainer/backend/internal/progress"
	"oge-ai-trainer/backend/internal/tasks"
)

func TestHealthReturnsDBStatus(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), "", true)
}

func TestRequestIDAndSecurityHeaders(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(requestIDHeader, "test-request-id")
	router.ServeHTTP(resp, req)

	if got := resp.Header().Get(requestIDHeader); got != "test-request-id" {
		t.Fatalf("expected request id header, got %q", got)
	}
	if got := resp.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", got)
	}
	if got := resp.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected DENY frame header, got %q", got)
	}
	if got := resp.Header().Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'none'") {
		t.Fatalf("unexpected CSP header %q", got)
	}
}

func TestHealthReturnsDBUnavailable(t *testing.T) {
	router := testRouter(testDeps{healthErr: app.DBUnavailable(errors.New("down"))})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeDBUnavailable, false)
}

func TestProtectedRouteRequiresBearerToken(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeUnauthorized, false)
}

func TestRegisterRejectsUnknownJSONField(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := jsonRequest(http.MethodPost, "/api/v1/auth/register", `{"email":"a@example.com","password":"password123","extra":true}`)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeValidation, false)
}

func TestAuthAllowsMissingContentType(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"a@example.com","password":"password123"}`))
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	assertJSONCode(t, resp.Body.Bytes(), "", true)
}

func TestRegisterRejectsPasswordLongerThanBcryptLimit(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := jsonRequest(http.MethodPost, "/api/v1/auth/register", `{"email":"a@example.com","password":"`+strings.Repeat("a", 73)+`"}`)
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeValidation, false)
}

func TestTaskCheckRejectsModeField(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := jsonRequest(http.MethodPost, "/api/v1/tasks/check", `{"task_id":1,"student_answer":"3","mode":"weak"}`)
	req.Header.Set("Authorization", "Bearer good-token")
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeValidation, false)
}

func TestTaskGenerateCustomRequiresTargetFields(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := jsonRequest(http.MethodPost, "/api/v1/tasks/generate", `{"mode":"custom"}`)
	req.Header.Set("Authorization", "Bearer good-token")
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeValidation, false)
}

func TestCORSAllowsLocalhost(t *testing.T) {
	router := testRouter(testDeps{})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.Code)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("unexpected CORS origin %q", got)
	}
}

func TestAuthRateLimit(t *testing.T) {
	router := NewRouter(Dependencies{
		Auth:                  fakeAuthService{},
		Tasks:                 fakeTaskService{},
		Diagnostic:            fakeDiagnosticService{},
		Progress:              fakeProgressService{},
		Health:                fakeHealth{},
		CORSAllowedOrigins:    []string{"http://localhost:5173"},
		AuthRateLimitRequests: 2,
		AuthRateLimitWindow:   time.Minute,
	})

	for i := 0; i < 2; i++ {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"a@example.com","password":"password123"}`))
		req.RemoteAddr = "192.0.2.10:12345"
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("request %d expected 200, got %d", i+1, resp.Code)
		}
	}

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"a@example.com","password":"password123"}`))
	req.RemoteAddr = "192.0.2.10:12345"
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.Code)
	}
	assertJSONCode(t, resp.Body.Bytes(), app.CodeRateLimited, false)
}

func testRouter(deps testDeps) http.Handler {
	return NewRouter(Dependencies{
		Auth:               fakeAuthService{},
		Tasks:              fakeTaskService{},
		Diagnostic:         fakeDiagnosticService{},
		Progress:           fakeProgressService{},
		Health:             fakeHealth{err: deps.healthErr},
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	})
}

func jsonRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func assertJSONCode(t *testing.T, raw []byte, code string, success bool) {
	t.Helper()
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("invalid JSON response: %v\n%s", err, string(raw))
	}
	if body.Success != success {
		t.Fatalf("expected success=%v, got %v", success, body.Success)
	}
	if code != "" && body.Error.Code != code {
		t.Fatalf("expected code %q, got %q", code, body.Error.Code)
	}
}

type testDeps struct {
	healthErr error
}

type fakeHealth struct {
	err error
}

func (f fakeHealth) Ping(context.Context) error {
	return f.err
}

type fakeAuthService struct{}

func (fakeAuthService) Register(context.Context, string, string) (auth.Session, error) {
	return auth.Session{Token: "token", User: auth.User{ID: 1, Email: "a@example.com"}}, nil
}

func (fakeAuthService) Login(context.Context, string, string) (auth.Session, error) {
	return auth.Session{Token: "token", User: auth.User{ID: 1, Email: "a@example.com"}}, nil
}

func (fakeAuthService) Me(context.Context, int64) (auth.User, error) {
	return auth.User{ID: 42, Email: "student@example.com"}, nil
}

func (fakeAuthService) ParseToken(token string) (int64, error) {
	if token == "good-token" {
		return 42, nil
	}
	return 0, app.Unauthorized("Требуется авторизация")
}

type fakeTaskService struct{}

func (fakeTaskService) Generate(context.Context, tasks.GenerateRequest) (tasks.Task, error) {
	return tasks.Task{ID: 1, OgeNumber: 9, SubtypeCode: "linear_equation", Question: "x + 2 = 5"}, nil
}

func (fakeTaskService) Check(context.Context, tasks.CheckRequest) (tasks.CheckResult, error) {
	return tasks.CheckResult{IsCorrect: true}, nil
}

func (fakeTaskService) Hint(context.Context, int64, int64) (tasks.HintResult, error) {
	return tasks.HintResult{Hint: "hint"}, nil
}

func (fakeTaskService) Explain(context.Context, int64, int64) (tasks.ExplainResult, error) {
	return tasks.ExplainResult{Explanation: "explain"}, nil
}

type fakeDiagnosticService struct{}

func (fakeDiagnosticService) Start(context.Context, int64) (diagnostic.StartResult, error) {
	return diagnostic.StartResult{SessionID: 1}, nil
}

func (fakeDiagnosticService) Submit(context.Context, int64, int64, []diagnostic.AnswerInput) (diagnostic.SubmitResult, error) {
	return diagnostic.SubmitResult{SessionID: 1}, nil
}

type fakeProgressService struct{}

func (fakeProgressService) GetProgress(context.Context, int64) ([]progress.Item, error) {
	return nil, nil
}

func (fakeProgressService) GetStats(context.Context, int64) (progress.Stats, error) {
	return progress.Stats{}, nil
}

func (fakeProgressService) GetRecommendations(context.Context, int64) ([]progress.Recommendation, error) {
	return nil, nil
}
